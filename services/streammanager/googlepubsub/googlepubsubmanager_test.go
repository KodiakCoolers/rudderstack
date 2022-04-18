package googlepubsub

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/ory/dockertest"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
)

var (
	hold     bool
	endpoint string
)

const (
	projectId = "my-project-id"
	topic     = "my-topic"
)

func Test_Timeout(t *testing.T) {

	config := Config{
		ProjectId: projectId,
		EventToTopicMap: []map[string]string{
			{"to": topic},
		},
		TestConfig: TestConfig{
			Endpoint: endpoint,
		},
	}

	client, err := NewProducer(config, Opts{Timeout: 1 * time.Microsecond})
	if err != nil {
		t.Fatalf("Expected no error, got: %s.", err)
	}
	json := `{"topicId": "my-topic", "message": "{}"}`
	statusCode, respStatus, responseMessage := Produce([]byte(json), client, nil)

	const expectedStatusCode = 504
	if statusCode != expectedStatusCode {
		t.Errorf("Expected status code %d, got %d.", expectedStatusCode, statusCode)
	}

	const expectedRespStatus = "Failure"
	if respStatus != expectedRespStatus {
		t.Errorf("Expected response status %s, got %s.", expectedRespStatus, respStatus)
	}

	const expectedResponseMessage = "[GooglePubSub] error :: Failed to publish:context deadline exceeded"
	if responseMessage != expectedResponseMessage {
		t.Errorf("Expected response message %s, got %s.", expectedResponseMessage, responseMessage)
	}
}

func TestMain(m *testing.M) {
	flag.BoolVar(&hold, "hold", false, "hold environment clean-up after test execution until Ctrl+C is provided")
	flag.Parse()

	// hack to make defer work, without being affected by the os.Exit in TestMain
	os.Exit(run(m))
}

func run(m *testing.M) int {
	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	pool.MaxWait = 2 * time.Minute
	if err != nil {
		log.Printf("Could not connect to docker: %s", err)
		return -1
	}

	pubsubContainer, err := pool.Run("messagebird/gcloud-pubsub-emulator", "latest", []string{
		"PUBSUB_PROJECT1=my-project-id,my-topic1",
	})
	if err != nil {
		log.Printf("Could not start resource: %s", err)
		return -1
	}
	defer func() {
		_ = recover()
		if err := pool.Purge(pubsubContainer); err != nil {
			log.Printf("Could not purge resource: %s \n", err)
			panic(err)
		}
	}()
	endpoint = fmt.Sprintf("127.0.0.1:%s", pubsubContainer.GetPort("8681/tcp"))
	client, err := pubsub.NewClient(
		context.Background(),
		projectId,
		option.WithoutAuthentication(),
		option.WithGRPCDialOption(grpc.WithInsecure()),
		option.WithEndpoint(endpoint))
	if err != nil {
		log.Printf("Expected no error, got: %s.", err)
		return -1
	}
	if err := pool.Retry(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		_, err = client.CreateTopic(ctx, topic)
		return err
	}); err != nil {
		log.Printf("Could not connect to docker: %s", err)
		return -1
	}

	code := m.Run()
	blockOnHold()

	return code
}

func blockOnHold() {
	if !hold {
		return
	}

	fmt.Println("Test on hold, before cleanup")
	fmt.Println("Press Ctrl+C to exit")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c
}
