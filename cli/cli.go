package cli

import (
	"context"
	"net/http"

	"github.com/alecthomas/kong"
)

type (
	Context struct {
		http        *http.Client
		ctx         context.Context
		Config      string `name: "config" help:"Location of client config files" default:".env" type:"path"`
		Debug       bool   `name: "debug" short:"D" help:"Enable debug mode"`
		LogLevel    string `name: "log-level" short:"L" help:"Set the logging level (debug|info|warn|error|fatal)" default:"info"`
		Output      string `name: "output" short:"O" help:"Output format (json|text)" default:"text"`
		ApiVersion  string `name: "api-version" short:"V" help:"the version of Whatsapp Cloud API to use" default:"v16.0"`
		BaseURL     string `name: "base-url" short:"b" help:"the base URL of Whatsapp Cloud API to use" default:"https://graph.facebook.com/"`
		PhoneID     string `name: "phone" short:"p" help:"phone ID of Whatsapp Cloud API to use"`
		WabaID      string `name: "waba" short:"w" help:"whatsapp business account id"`
		AccessToken string `name "token" short:"T" help:"access token of Whatsapp Cloud API to use"`
		Timeout     int    `name: "timeout" short:"t" help:"http timeout for making api calls" default:"30"`
	}

	cli struct {
		Context
		Send SendCommand `cmd:"" name:"send" help:"send different types of messages like text, image, video, audio, document, location, vcard, template, sticker, and file"`
	}

	App struct {
		cli cli
	}
)

func NewApp() *App {
	return &App{
		cli: cli{
			Context: Context{
				http: http.DefaultClient,
				ctx:  context.Background(),
			},
		},
	}
}

func (app *App) Run() error {
	cli := app.cli

	ctx := kong.Parse(&cli,
		kong.Name("whatsapp"),
		kong.Description("using whatsapp cloud api from the command line"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}),
		kong.Vars{
			"version": "0.0.1",
		})
	return ctx.Run(&cli.Context)
}
