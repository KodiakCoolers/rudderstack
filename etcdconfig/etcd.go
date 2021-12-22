package etcdconfig

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/rudderlabs/rudder-server/config"
	"github.com/rudderlabs/rudder-server/rruntime"
	"github.com/rudderlabs/rudder-server/utils"
	"github.com/rudderlabs/rudder-server/utils/logger"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	cli              *clientv3.Client
	etcdHosts        []string
	podPrefix        string
	connectTimeout   time.Duration
	etcdGetTimeout   time.Duration
	etcdWatchTimeout time.Duration
	pkgLogger        logger.LoggerI
	releaseName      string
	serverNumber     string

	podStatus                 string
	podStatusWaitGroup        *sync.WaitGroup
	podStatusLock             sync.RWMutex
	podStatuswatchInitialized bool
)

const PODSTATUS = `ETCD_POD_STATUS`

func Init() {
	loadConfig()
}

type EtcdService struct {
	cli                    *clientv3.Client
	workSpaceConfigChannel chan map[string]string
}

func loadConfig() {
	etcdHosts = strings.Split(config.GetEnv("ETCD_HOST", "127.0.0.1:2379"), `,`)
	releaseName = config.GetEnv("RELEASE_NAME", `multitenantv1`)
	serverNumber = config.GetEnv("SERVER_NUMBER", `1`)
	podPrefix = releaseName + `/SERVER/` + serverNumber
	config.RegisterDurationConfigVariable(time.Duration(15), &etcdGetTimeout, true, time.Second, "ETCD_GET_TIMEOUT")
	config.RegisterDurationConfigVariable(time.Duration(3), &connectTimeout, true, time.Second, "ETCD_CONN_TIMEOUT")
	config.RegisterDurationConfigVariable(time.Duration(3), &etcdWatchTimeout, true, time.Second, "ETCD_WATCH_TIMEOUT")
	pkgLogger = logger.NewLogger().Child("etcd")
	podStatusLock = sync.RWMutex{}
	podStatusWaitGroup = &sync.WaitGroup{}
	connectToETCD()
}

func connectToETCD() {}

//returns a channel watching for changes in workspaces that this pod serves
func WatchForWorkspaces(ctx context.Context) chan map[string]string {
	returnChan := make(chan map[string]string)
	go func(returnChan chan map[string]string, ctx context.Context) {
		defer cli.Close()
		etcdWatchChan := cli.Watch(ctx, podPrefix+`/workspaces`)
		for watchResp := range etcdWatchChan {
			for _, event := range watchResp.Events {
				switch event.Type {
				case mvccpb.PUT:
					returnChan <- map[string]string{
						"type":       "PUT",
						"workSpaces": string(event.Kv.Value),
					}
				case mvccpb.DELETE:
					returnChan <- map[string]string{
						"type":       "DELETE",
						"workSpaces": "",
					}
					//we can close this channel now..?
				}
			}
		}
	}(returnChan, ctx)
	return returnChan
}

//returns the initial workspaces this pod must serve
//
//along with a watchChan to watch further updates in the workspaces
func GetWorkspaces(ctx context.Context) (string, chan map[string]string) {
	clientReturnChan := make(chan *clientv3.Client)
	errChan := make(chan error)
	go GetEtcdClient(ctx, clientReturnChan, errChan)

	select {
	case cli := <-clientReturnChan:
		initialWorkspaces, err := cli.Get(ctx, podPrefix+`/workspaces`)
		if err != nil {
			panic(err)
		}
		var workSpaceString string
		if len(initialWorkspaces.Kvs) > 0 {
			workSpaceString = string(initialWorkspaces.Kvs[0].Value)
		} else {
			workSpaceString = ``
		}
		watchChan := WatchForWorkspaces(ctx)

		return workSpaceString, watchChan
	case <-time.After(connectTimeout):
		panic("Couldn't find etcd Client")
	case err := <-errChan:
		panic(err)
	}
}

func GetEtcdClient(ctx context.Context, clientReturnChan chan *clientv3.Client, errChan chan error) {
	var err error
	cli, err = clientv3.New(clientv3.Config{
		Endpoints:   etcdHosts,
		DialTimeout: connectTimeout,
	})
	if err != nil {
		panic(err)
	}

	statusRes, err := cli.Status(ctx, etcdHosts[0])
	if err != nil {
		errChan <- err
		return
	} else if statusRes == nil {
		errChan <- errors.New("statusRes is nil")
		return
	}
	etcdHeartBeat(ctx)
	clientReturnChan <- cli
	go MigrationWatch(ctx)
}

func etcdHeartBeat(ctx context.Context) {
	client := cli
	etcdConnectTimeout := connectTimeout
	rruntime.Go(func() {
		for {
			func(etcdclient *clientv3.Client, etcdConnectTimeout time.Duration) {
				ctxHeartBeat, cancel := context.WithTimeout(context.Background(), etcdConnectTimeout)
				defer cancel()
				heartBeatChan := make(chan bool)

				go func(ctx context.Context, etcdclient *clientv3.Client, heartBeatChan chan bool) {
					heartBeatFunc(ctx, etcdclient, heartBeatChan)
				}(ctxHeartBeat, client, heartBeatChan)

				select {
				case <-ctxHeartBeat.Done():
					panic(ctxHeartBeat.Err())
				case <-heartBeatChan:
					time.Sleep(1 * time.Second)
				}
			}(client, etcdConnectTimeout)
		}
	})
}

func heartBeatFunc(ctxHeartBeat context.Context, client *clientv3.Client, heartBeatChan chan bool) {
	lease, err := client.Lease.Grant(ctxHeartBeat, 2)
	if err != nil {
		panic(err)
	}

	_, err = client.Put(ctxHeartBeat, podPrefix, `alive`, clientv3.WithLease(lease.ID))
	if err != nil {
		panic(err)
	}
	heartBeatChan <- true
}

var Eb utils.PublishSubscriber = new(utils.EventBus)

func WatchForMigration(ctx context.Context, statusWatchChannel chan utils.DataEvent) (string, *sync.WaitGroup) {
	Eb.Subscribe(PODSTATUS, statusWatchChannel)

	podStatusLock.RLock()
	defer podStatusLock.RUnlock()
	if !podStatuswatchInitialized {
		go MigrationWatch(ctx)
	}

	//get current state
	var initialPodState string
	initialState, err := cli.Get(ctx, podPrefix+`/mode`)
	if err != nil {
		panic(err)
	}
	if len(initialState.Kvs) > 0 {
		initialPodState = string(initialState.Kvs[0].Value)
	} else {
		initialPodState = `` //or simply degraded?	works the same now anyway -> processor, router don't start unless initialPodState = `normal`
	}
	return initialPodState, podStatusWaitGroup
}

func MigrationWatch(ctx context.Context) {
	podStatusLock.Lock()
	podStatuswatchInitialized = true
	podStatusLock.Unlock()
	defer cli.Close()
	watchCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	etcdMigrationStatusChannel := cli.Watch(watchCtx, podPrefix+`/mode`)
	for watchResp := range etcdMigrationStatusChannel {
		for _, event := range watchResp.Events {
			switch event.Type {
			case mvccpb.PUT:
				podStatus = string(event.Kv.Value)
			case mvccpb.DELETE:
				//This pod's status has been deleted from etcd store, pod no longer needed..?
				podStatus = `terminated`
			}
			podStatusWaitGroup.Add(Eb.NumSubscribers(PODSTATUS))
			Eb.Publish(PODSTATUS, podStatus)
			podStatusWaitGroup.Wait()
			cli.Put(watchCtx, podPrefix+`/status`, podStatus+`_completed`)
		}
	}
}