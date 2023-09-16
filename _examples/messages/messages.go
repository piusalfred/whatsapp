package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/piusalfred/whatsapp"
	whttp "github.com/piusalfred/whatsapp/http"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
)

func main() {
	err := setup()
	if err != nil {
		fmt.Printf("error setting up: %v\n", err) //nolint:forbidigo
		os.Exit(1)
	}
}

const quitMessage = `Press Ctrl+C to quit at any point to quit`

var ErrInterrupted = fmt.Errorf("interrupted")

// setup runs a simple interactive commandline tool to show that you have successfully
// configured your whatsapp business account to send messages.
func setup() error {
	writer := os.Stdout
	if _, err2 := writer.WriteString(quitMessage); err2 != nil {
		return err2
	}
	logger := slog.New(slog.NewTextHandler(writer, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}))
	hook := whttp.LogRequestHook(logger)
	respHook := whttp.LogResponseHook(logger)

	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		sig := <-sigChan
		err := fmt.Errorf("%w: received signal: %v", ErrInterrupted, sig)
		cancel(err)
	}()
	configer := &configer{reader: bufio.NewReader(os.Stdin)}
	config, err := configer.Read(ctx)
	if err != nil {
		return err
	}
	client, err := whatsapp.NewClientWithConfig(config,
		whatsapp.WithBaseClient(&whatsapp.BaseClient{Client: whttp.NewClient(
			whttp.WithHTTPClient(http.DefaultClient),
			whttp.WithRequestHooks(hook),
			whttp.WithResponseHooks(respHook),
		)}))
	if err != nil {
		return err
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Enter the recipient phone number (this number must be registered in FB portal): ")
	recipient, err := reader.ReadString('\n')
	if err != nil {
		return err
	}

	fmt.Println("Sending Template Message (make sure you reply): ")
	// Send a Template ( We will use the default template called hello_world)
	tmpl := &whatsapp.Template{
		LanguageCode:   "en_US",
		LanguagePolicy: "",
		Name:           "hello_world",
		Components:     nil,
	}

	response, err := client.SendTemplate(ctx, recipient, tmpl)
	if err != nil {
		return err
	}

	logger.LogAttrs(ctx, slog.LevelInfo, "send template", slog.Group("response", response))

	message := &whatsapp.TextMessage{
		Message:    "😺Find me at https://github.com/piusalfred/whatsapp 👩🏻‍🦰",
		PreviewURL: true,
	}

	response, err = client.SendTextMessage(ctx, recipient, message)
	if err != nil {
		return fmt.Errorf("send message: %w", err)
	}

	logger.LogAttrs(ctx, slog.LevelInfo, "send template",
		slog.Group("response", response))

	select {
	case <-ctx.Done():
		return fmt.Errorf("interupted: %w", ctx.Err())

	default:
		return nil
	}
}

var _ whatsapp.ConfigReader = (*configer)(nil)

// configer implements whatsapp.ConfigReaderFunc it basically asks user to enter the required
// configuration values via the commandline.
type configer struct {
	reader *bufio.Reader
}

func (c *configer) Read(ctx context.Context) (*whatsapp.Config, error) {
	doneChan := make(chan struct{}, 1)
	errChan := make(chan error, 1)
	var config whatsapp.Config

	go func() {
		fmt.Println("Enter your access token: ")
		token, err := c.reader.ReadString('\n')
		if err != nil {
			errChan <- err

			return
		}
		config.AccessToken = strings.TrimSpace(token)

		fmt.Println("Enter your phone number ID: ")
		phoneID, err := c.reader.ReadString('\n')
		if err != nil {
			errChan <- err

			return
		}

		config.PhoneNumberID = strings.TrimSpace(phoneID)

		fmt.Println("Enter your business account ID: ")
		businessID, err := c.reader.ReadString('\n')
		if err != nil {
			errChan <- err

			return
		}

		config.BusinessAccountID = strings.TrimSpace(businessID)

		fmt.Println("Enter API version:(Lowest version is v16.0) ")
		version, err := c.reader.ReadString('\n')
		if err != nil {
			errChan <- err

			return
		}

		config.Version = strings.TrimSpace(version)

		doneChan <- struct{}{}
	}()

	select {
	case <-doneChan:
		return &config, nil

	case err := <-errChan:
		return nil, err

	case <-ctx.Done():
		return nil, fmt.Errorf("interrupted: %w", ctx.Err())
	}
}

