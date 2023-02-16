package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/piusalfred/whatsapp"
	"github.com/urfave/cli/v2"
)

const (
	flagWhatsappBusinessID    = "business-id"
	defaultWhatsappBusinessID = ""
	envWhatsappBusinessID     = "WHATSAPP_BUSINESS_ID"
	descWhatsappBusinessID    = "WhatsApp Business ID"
	flagPhoneNumberID         = "phone-id"
	defaultPhoneNumberID      = ""
	envPhoneNumberID          = "PHONE_NUMBER_ID"
	descPhoneNumberID         = "Phone Number ID"
	flagApiVersion            = "api-version"
	defaultApiVersion         = "v16.0"
	envApiVersion             = "API_VERSION"
	descApiVersion            = "Whatsapp Business Cloud API Version"
	flagAccessToken           = "access-token"
	defaultAccessToken        = ""
	envAccessToken            = "ACCESS_TOKEN"
	descAccessToken           = "Whatsapp Business Cloud API Access Token"
	flagBaseURL               = "base-url"
	defaultBaseURL            = whatsapp.BaseURL
	envBaseURL                = "BASE_URL"
	descBaseURL               = "Whatsapp Business Cloud API Base URL"
	flagConfigPath            = "config-path"
	defaultConfigPath         = ""
	envConfigPath             = "CONFIG_PATH"
	descConfigPath            = "Path to the config file"
)

type App struct {
	HTTP        *http.Client
	businessID  string
	phoneID     string
	apiVersion  string
	accessToken string
	baseURL     string
	configPath  string
	app         *cli.App
}

func main() {
	commander := &cli.App{
		Name:                 "whatsapp",
		Usage:                "use whatsapp from the command line",
		EnableBashCompletion: true,
		Action: func(*cli.Context) error {
			time.Sleep(5 * time.Second)
			fmt.Println("boom! I say!")
			return nil
		},
		Commands: []*cli.Command{},
	}

	app := &App{
		HTTP: http.DefaultClient,
		app:  commander,
	}

	app.initCommonFlags()

	app.initSubCommands()

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

// initSubCommands initializes the sub commands
func (a *App) initSubCommands() {
	sendCommandParams := &CommandParams{
		Name:        "send",
		Usage:       "send a message",
		Description: "send a message to a phone number it can be a text, media or template message",
	}

	texterFlgs := []cli.Flag{
		&cli.StringFlag{
			Name:    "message",
			Aliases: []string{"m"},
			Usage:   "message to send",
		},
		&cli.StringFlag{
			Name:    "phone",
			Aliases: []string{"p"},
			Usage:   "phone number to send to",
		},
		&cli.BoolFlag{
			Name:    "preview-url",
			Aliases: []string{"u"},
			Usage:   "allow to preview url",
		},
	}

	sendTextActionFunc := func(c *cli.Context) error {
		message := c.String("message")
		phone := c.String("phone")
		previewURL := c.Bool("preview-url")
		request := &whatsapp.SendTextRequest{
			Recipient:     phone,
			Message:       message,
			PreviewURL:    previewURL,
			BaseURL:       a.baseURL,
			AccessToken:   a.accessToken,
			PhoneNumberID: a.phoneID,
			ApiVersion:    a.apiVersion,
		}
		response, err := whatsapp.SendText(c.Context, a.HTTP, request)
		if err != nil {
			return err
		}

		fmt.Println(response)
		return nil
	}

	sendTextSubCommandParams := &CommandParams{
		Name:        "text",
		Usage:       "send a text message",
		Description: "send a text message to a phone number",
		Flags:       texterFlgs,
		Action:      sendTextActionFunc,
	}

	sendCommand := NewCommand(sendCommandParams).Get()
	sendTextSubCommand := NewCommand(sendTextSubCommandParams).Get()
	sendCommand.Subcommands = append(sendCommand.Subcommands, sendTextSubCommand)
	a.app.Commands = append(a.app.Commands, sendCommand)
}

// initCommonFlags initializes the common flags
func (a *App) initCommonFlags() {
	a.app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:        flagWhatsappBusinessID,
			Aliases:     []string{"b"},
			Usage:       descWhatsappBusinessID,
			EnvVars:     []string{envWhatsappBusinessID},
			Value:       defaultWhatsappBusinessID,
			Destination: &a.businessID,
		},
		&cli.StringFlag{
			Name:        flagPhoneNumberID,
			Aliases:     []string{"p"},
			Usage:       descPhoneNumberID,
			EnvVars:     []string{envPhoneNumberID},
			Value:       defaultPhoneNumberID,
			Destination: &a.phoneID,
		},
		&cli.StringFlag{
			Name:        flagApiVersion,
			Aliases:     []string{"v"},
			Usage:       descApiVersion,
			EnvVars:     []string{envApiVersion},
			Value:       defaultApiVersion,
			Destination: &a.apiVersion,
		},
		&cli.StringFlag{
			Name:        flagAccessToken,
			Aliases:     []string{"a"},
			Usage:       descAccessToken,
			EnvVars:     []string{envAccessToken},
			Value:       defaultAccessToken,
			Destination: &a.accessToken,
		},
		&cli.StringFlag{
			Name:        flagBaseURL,
			Aliases:     []string{"u"},
			Usage:       descBaseURL,
			EnvVars:     []string{envBaseURL},
			Value:       defaultBaseURL,
			Destination: &a.baseURL,
		},
		&cli.StringFlag{
			Name:        flagConfigPath,
			Aliases:     []string{"c"},
			Usage:       descConfigPath,
			EnvVars:     []string{envConfigPath},
			Value:       defaultConfigPath,
			Destination: &a.configPath,
		},
	}
}

// Run takes context and runs the app
func (a *App) Run(args []string) error {
	ctx, cancel := context.WithCancel(context.TODO())
	terminate := make(chan os.Signal, 1)
	defer func() {
		signal.Stop(terminate)
		cancel()
	}()

	go func() {
		select {
		case <-terminate:
			cancel()
			return // exit
		case <-ctx.Done():
			return // exit
		}
	}()

	return a.app.RunContext(ctx, args)
}
