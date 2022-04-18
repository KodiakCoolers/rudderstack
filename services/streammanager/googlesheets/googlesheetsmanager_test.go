package googlesheets

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/ory/dockertest"
	mock_logger "github.com/rudderlabs/rudder-server/mocks/utils/logger"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

var (
	hold     bool
	endpoint string
)

const (
	sheetId       = "sheetId"
	sheetName     = "sheetName"
	destinationId = "destinationId"
	header1       = "Product Purchased"
	header2       = "Product Value"
	accessToken   = "cd887efc-7c7d-4e8e-9580-f7502123badf"
	refreshToken  = "bdbbe5ec-6081-4c6c-8974-9c4abfc0fdcc"
)

func Test_Timeout(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	mockLogger := mock_logger.NewMockLoggerI(mockCtrl)
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()
	pkgLogger = mockLogger

	config := Config{
		SheetId:     sheetId,
		SheetName:   sheetName,
		DestID:      destinationId,
		Credentials: "",
		EventKeyMap: []map[string]string{
			{"to": header1},
			{"to": header2},
		},
		TestConfig: TestConfig{
			Endpoint:     endpoint,
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		},
	}
	client, err := NewProducer(config, Opts{Timeout: 10 * time.Second})
	if err != nil {
		t.Fatalf(" %+v", err)
	}
	client.opts = Opts{Timeout: 1 * time.Microsecond}
	json := fmt.Sprintf(`{
		"spreadSheetId": "%s",
		"spreadSheet": "%s",
		"message":{
			"0": { "attributeKey": "%s", "attributeValue": "Realme C3" }
			"1": { "attributeKey": "%s", "attributeValue": "5900"}
		}
	}`, sheetId, sheetName, header1, header2)
	statusCode, respStatus, responseMessage := Produce([]byte(json), client, nil)
	const expectedStatusCode = 504
	if statusCode != expectedStatusCode {
		t.Errorf("Expected status code %d, got %d.", expectedStatusCode, statusCode)
	}

	const expectedRespStatus = "Failure"
	if respStatus != expectedRespStatus {
		t.Errorf("Expected response status %s, got %s.", expectedRespStatus, respStatus)
	}

	const expectedResponseMessage = "[GoogleSheets] error :: Failed to insert Payload :: context deadline exceeded"
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

	dockerContainer, err := pool.Run("atzoum/simulator-google-sheets", "latest", []string{})
	if err != nil {
		log.Printf("Could not start resource: %s", err)
		return -1
	}
	defer func() {
		_ = recover()
		if err := pool.Purge(dockerContainer); err != nil {
			log.Printf("Could not purge resource: %s \n", err)
			panic(err)
		}
	}()
	endpoint = fmt.Sprintf("https://127.0.0.1:%s/", dockerContainer.GetPort("8443/tcp"))
	token := &oauth2.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	config := &tls.Config{
		// skipcq: GSC-G402
		InsecureSkipVerify: true,
	}
	client := oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(token))
	trans := client.Transport.(*oauth2.Transport)
	trans.Base = &http.Transport{TLSClientConfig: config}
	sheetService, err := sheets.NewService(context.Background(), option.WithEndpoint(endpoint), option.WithHTTPClient(client))

	if err := pool.Retry(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		_, err = sheetService.Spreadsheets.Get(sheetId).Context(ctx).Do()
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
