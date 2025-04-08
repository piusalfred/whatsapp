package main

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"os"

	"github.com/joho/godotenv"

	"github.com/piusalfred/whatsapp/auth"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
)

// Config holds the WhatsApp Cloud API configuration parameters.
type Config struct {
	BaseURL           string // Base URL for the API
	APIVersion        string // API version (e.g., v21.0)
	AccessToken       string // Access token for API requests
	PhoneNumberID     string // Phone number EntryID used in the API
	BusinessAccountID string // Business account EntryID
	TestRecipient     string // Test recipient phone number
	ApplicationID     string // Application EntryID
	ApplicationSecret string // Application secret
	BusinessID        string // Business EntryID
	ClientToken       string // Client token
	ClientID          string // Client EntryID
	SystemUserID      string // System user EntryID
}

// LoadConfigFromFile loads configuration parameters from the provided .env file.
func LoadConfigFromFile(filepath string) (*Config, error) {
	// Load the .env file from the given path
	err := godotenv.Load(filepath)
	if err != nil {
		return nil, err
	}

	// Populate the Config struct with environment variables
	conf := &Config{
		BaseURL:           os.Getenv("WHATSAPP_CLOUD_API_BASE_URL"),
		APIVersion:        os.Getenv("WHATSAPP_CLOUD_API_API_VERSION"),
		AccessToken:       os.Getenv("WHATSAPP_CLOUD_API_ACCESS_TOKEN"),
		PhoneNumberID:     os.Getenv("WHATSAPP_CLOUD_API_PHONE_NUMBER_ID"),
		BusinessAccountID: os.Getenv("WHATSAPP_CLOUD_API_BUSINESS_ACCOUNT_ID"),
		TestRecipient:     os.Getenv("WHATSAPP_CLOUD_API_TEST_RECIPIENT"),
		ApplicationID:     os.Getenv("WHATSAPP_CLOUD_API_APPLICATION_ID"),
		ApplicationSecret: os.Getenv("WHATSAPP_CLOUD_API_APP_SECRET"),
		BusinessID:        os.Getenv("WHATSAPP_CLOUD_API_BUSINESS_ID"),
		ClientToken:       os.Getenv("WHATSAPP_CLOUD_API_CLIENT_TOKEN"),
		ClientID:          os.Getenv("WHATSAPP_CLOUD_API_CLIENT_ID"),
		SystemUserID:      os.Getenv("WHATSAPP_CLOUD_API_SYSTEM_USER_ID"),
	}

	return conf, nil
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}))

	clientOptions := []whttp.CoreClientOption[any]{
		whttp.WithCoreClientRequestInterceptor[any](
			func(ctx context.Context, req *http.Request) error {
				logger.LogAttrs(ctx, slog.LevelInfo, "request intercepted",
					slog.String("http.request.method", req.Method),
					slog.String("http.request.url", req.URL.String()),
				)
				return nil
			},
		),
		whttp.WithCoreClientResponseInterceptor[any](
			func(ctx context.Context, resp *http.Response) error {
				dumpResponse, _ := httputil.DumpResponse(resp, true)

				logger.LogAttrs(ctx, slog.LevelInfo, "response intercepted",
					slog.String("http.response.status", resp.Status),
					slog.Int("http.response.code", resp.StatusCode),
					slog.String("body", string(dumpResponse)),
				)
				return nil
			},
		),
	}
	ctx := context.Background()

	coreClient := whttp.NewAnySender(clientOptions...)

	conf, err := LoadConfigFromFile("api.env")
	if err != nil {
		logger.LogAttrs(ctx, slog.LevelError, "error loading configs", slog.String("error", err.Error()))
		return
	}

	client := auth.NewClient(conf.BaseURL, conf.APIVersion, coreClient)

	response, err := client.GenerateAccessToken(ctx, auth.GenerateAccessTokenParams{
		AccessToken:  conf.AccessToken,
		AppID:        conf.ApplicationID,
		SystemUserID: conf.SystemUserID,
		AppSecret:    conf.ApplicationSecret,
		Scopes: []string{
			auth.TokenScopeWhatsappBusinessManagement,
			auth.TokenScopeWhatsappBusinessMessaging,
		},
		SetTokenExpiresIn60: true,
	})
	if err != nil {
		return
	}

	logger.LogAttrs(ctx, slog.LevelInfo, "token", slog.Any("response", response))
}
