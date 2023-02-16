package cli

import (
	"context"
	"fmt"
	"io"
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

var (
	_ Commander = (*Command)(nil)
)

type (
	App struct {
		HTTP        *http.Client
		Logger      io.Writer
		businessID  string
		phoneID     string
		apiVersion  string
		accessToken string
		baseURL     string
		configPath  string
		app         *cli.App
	}

	AppOption func(*App)

	Commander interface {
		Init(name, usage, description string, aliases ...string) *cli.Command
		SetRunners(cli.BeforeFunc, cli.ActionFunc, cli.AfterFunc, cli.OnUsageErrorFunc)
		SetSubCommands(...*cli.Command)
		SetFlags(...cli.Flag)
		Get() *cli.Command
	}

	Command struct {
		Logger io.Writer
		C      *cli.Command
	}

	CommandParams struct {
		Name         string
		Usage        string
		Description  string
		Aliases      []string
		Before       cli.BeforeFunc
		Action       cli.ActionFunc
		After        cli.AfterFunc
		OnUsageError cli.OnUsageErrorFunc
		Flags        []cli.Flag
		SubCommands  []*cli.Command
	}

	CommandOption func(*Command)
)

func NewCommand(params *CommandParams, opts ...CommandOption) *Command {
	c := &Command{}
	c.Init(params.Name, params.Usage, params.Description, params.Aliases...)
	c.SetRunners(params.Before, params.Action, params.After, params.OnUsageError)
	c.SetFlags(params.Flags...)
	c.SetSubCommands(params.SubCommands...)

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func (cmd *Command) Init(name, usage, description string, aliases ...string) *cli.Command {
	cmd.C = &cli.Command{
		Name:        name,
		Usage:       usage,
		Description: description,
		Aliases:     aliases,
	}

	return cmd.C
}

func (cmd *Command) SetRunners(before cli.BeforeFunc, action cli.ActionFunc, after cli.AfterFunc, onUsageError cli.OnUsageErrorFunc) {
	cmd.C.Before = before
	cmd.C.Action = action
	cmd.C.After = after
	cmd.C.OnUsageError = onUsageError
}

func (cmd *Command) SetSubCommands(subCommands ...*cli.Command) {
	cmd.C.Subcommands = subCommands
}

func (cmd *Command) SetFlags(flags ...cli.Flag) {
	cmd.C.Flags = flags
}

func (cmd *Command) Get() *cli.Command {
	return cmd.C
}

func New(options ...AppOption) *App {
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

	for _, opt := range options {
		opt(app)
	}

	return app
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
