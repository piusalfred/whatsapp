package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"github.com/piusalfred/whatsapp/config"
	"github.com/piusalfred/whatsapp/flow"
	"github.com/piusalfred/whatsapp/message"
	"github.com/piusalfred/whatsapp/pkg/crypto"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
)

func LoadConfigFromFile(filepath string) (config.ReaderFunc, string) {
	err := godotenv.Load(filepath) // Load the .env file from the given path
	if err != nil {
		panic(err)
	}
	recipient := os.Getenv("WHATSAPP_CLOUD_API_TEST_RECIPIENT")

	secureRequestsStr := os.Getenv("WHATSAPP_CLOUD_API_SECURE_REQUESTS")

	secureRequests := false

	if secureRequestsStr == "true" {
		secureRequests = true
	}

	conf := &config.Config{
		BaseURL:           os.Getenv("WHATSAPP_CLOUD_API_BASE_URL"),
		APIVersion:        os.Getenv("WHATSAPP_CLOUD_API_API_VERSION"),
		AccessToken:       os.Getenv("WHATSAPP_CLOUD_API_ACCESS_TOKEN"),
		PhoneNumberID:     os.Getenv("WHATSAPP_CLOUD_API_PHONE_NUMBER_ID"),
		BusinessAccountID: os.Getenv("WHATSAPP_CLOUD_API_BUSINESS_ACCOUNT_ID"),
		AppSecret:         os.Getenv("WHATSAPP_CLOUD_API_APP_SECRET"),
		SecureRequests:    secureRequests,
	}

	fn := func(ctx context.Context) (*config.Config, error) {
		return conf, nil
	}

	prrof, err := crypto.GenerateAppSecretProof(conf.AccessToken, conf.AppSecret)
	fmt.Println("PROOOOOF "+prrof, " error: ", err)

	return fn, recipient
}

type Clients struct {
	Message *message.BaseClient
	Flow    *flow.BaseClient
}

func CreateClients() (*Clients, error) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}))

	clientOptions := []whttp.CoreClientOption[message.Message]{
		whttp.WithCoreClientRequestInterceptor[message.Message](
			func(ctx context.Context, req *http.Request) error {
				logger.LogAttrs(ctx, slog.LevelInfo, "request intercepted",
					slog.String("http.request.method", req.Method),
					slog.String("http.request.url", req.URL.String()),
					slog.Any("headers", req.Header),
				)
				return nil
			},
		),
		whttp.WithCoreClientResponseInterceptor[message.Message](
			func(ctx context.Context, resp *http.Response) error {
				logger.LogAttrs(ctx, slog.LevelInfo, "response intercepted",
					slog.String("http.response.status", resp.Status),
					slog.Int("http.response.code", resp.StatusCode),
					slog.Any("headers", resp.Header),
				)

				return nil
			},
		),
	}

	ctx := context.Background()

	coreClient := whttp.NewSender[message.Message](clientOptions...)
	reader, _ := LoadConfigFromFile("/Users/piusalfred/dev/github/whatsapp/examples/api.env")
	baseClient, err := message.NewBaseClient(coreClient, reader)
	if err != nil {
		logger.LogAttrs(ctx, slog.LevelError, "error creating base client", slog.String("error", err.Error()))
	}

	flowSender := whttp.NewAnySender()

	flowClient := flow.NewBaseClient(reader, flowSender)

	clients := &Clients{
		Message: baseClient,
		Flow:    flowClient,
	}

	return clients, nil
}
