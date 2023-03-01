/*
 * Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of this software
 * and associated documentation files (the “Software”), to deal in the Software without restriction,
 * including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
 * and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
 * subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all copies or substantial
 * portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
 * LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
 * IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
 * WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
 * SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

//nolint:lll
package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/piusalfred/whatsapp"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"text/template"

	"github.com/alecthomas/kong"
	dotenv "github.com/joho/godotenv"
	whttp "github.com/piusalfred/whatsapp/http"
)

type (
	Context struct {
		http        *http.Client
		ctx         context.Context
		logger      io.Writer
		loader      ConfigLoader
		ConfigPath  string `name: "config" help:"Location of client config files" default:".env" type:"path"`
		Debug       bool   `name: "debug" short:"D" help:"Enable debug mode" default:"false"`
		LogLevel    string `name: "log-level" short:"L" help:"Set the logging level (debug|info|warn|error|fatal)" default:"info"`
		Output      string `name: "output" short:"O" help:"Output format (json|text|pretty)" default:"text"`
		ApiVersion  string `name: "api-version" short:"V" help:"the version of Whatsapp Cloud API to use" default:"v16.0"`
		BaseURL     string `name: "base-url" short:"b" help:"the base URL of Whatsapp Cloud API to use" default:"https://graph.facebook.com/"`
		PhoneID     string `name: "phone" short:"p" help:"phone ID of Whatsapp Cloud API to use"`
		WabaID      string `name: "waba" short:"w" help:"whatsapp business account id"`
		AccessToken string `name "token" short:"T" help:"access token of Whatsapp Cloud API to use"`
		Timeout     int    `name: "timeout" short:"t" help:"http timeout for making api calls" default:"30"`
	}

	Config struct {
		BaseURL           string
		Version           string
		PhoneID           string
		BusinessAccountID string
		AccessToken       string
	}

	cli struct {
		Context
		Send    SendCommand    `cmd:"" name:"send" help:"send different types of messages like text, image, video, audio, document, location, vcard, template, sticker, and file"`
		QrCodes QrcodesCommand `cmd:"" name:"qrcodes" help:"manage qr codes"`
	}

	App struct {
		mu        sync.Mutex
		envLoader ConfigLoader
		cli       cli
	}

	Option func(*App)
)

func NewApp(options ...Option) *App {
	app := &App{
		cli: cli{
			Context: Context{
				http:   http.DefaultClient,
				ctx:    context.Background(),
				logger: os.Stdout,
			},
		},
	}

	for _, option := range options {
		option(app)
	}

	return app
}

// SetLogger sets the logger for the app
func SetLogger(logger io.Writer) Option {
	return func(app *App) {
		app.mu.Lock()
		defer app.mu.Unlock()
		app.cli.logger = logger
	}
}

// SetConfigFile sets the config file for the app
func SetConfigFile(config string) Option {
	return func(app *App) {
		app.mu.Lock()
		defer app.mu.Unlock()
		app.cli.ConfigPath = config
		app.envLoader = defaultEnvLoader
		conf, err := app.envLoader(app.cli.ConfigPath)
		if err != nil {
			panic(fmt.Errorf("failed to load config file: %w", err))
		}

		app.cli.BaseURL = conf.BaseURL
		app.cli.ApiVersion = conf.Version
		app.cli.PhoneID = conf.PhoneID
		app.cli.WabaID = conf.BusinessAccountID
		app.cli.AccessToken = conf.AccessToken
		app.cli.Timeout = conf.Timeout
		app.cli.Output = conf.OutputFormat
	}
}

// SetConfigFilePath sets the config file for the Context
func (ctx *Context) SetConfigFilePath(path string) error {
	ctx.ConfigPath = path
	ctx.loader = defaultEnvLoader
	conf, err := ctx.loader(ctx.ConfigPath)
	if err != nil {
		return fmt.Errorf("failed to load config file: %w", err)
	}

	ctx.BaseURL = conf.BaseURL
	ctx.ApiVersion = conf.Version
	ctx.PhoneID = conf.PhoneID
	ctx.WabaID = conf.BusinessAccountID
	ctx.AccessToken = conf.AccessToken
	return nil
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

const (
	JsonOutputFormat       OutputFormat = "json"
	TextOutputFormat       OutputFormat = "text"
	PrettyJsonOutputFormat OutputFormat = "pretty"
)

const responseTextTemplate = `
Status: {{.StatusCode}}
Headers:
{{range $key, $value := .Headers}}
	  {{$key}}: {{$value}}
{{end}}
Message IDs:
{{range $value := .MessageIDs}}
	  - {{$value}}
{{end}}
`

var ErrUnknownOutputFormat = fmt.Errorf("unknown output format")

// printResponse is a function that prettifies and print the response to the writer.
// it takes a io.Writer, OutputFormat, and the whatsapp response as arguments.
func printResponse(w io.Writer, response *whttp.Response, format OutputFormat) error {
	switch format {
	case PrettyJsonOutputFormat:
		b, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal response: %w", err)
		}
		_, err = w.Write(b)

		return fmt.Errorf("failed to write response: %w", err)

	case TextOutputFormat:
		// use text template to print the  response in a human readable format
		// example:
		// Status: 200
		// Headers:
		//    Content-Type: application/json, text/javascript, */*; q=0.01
		//    Date: Mon, 01 Feb 2021 12:00:00 GMT
		//    ETag: "5e1f-5a3e3b6b1b680-gzip"
		// Message IDs:
		//   - 1234567890
		//   - 0987654321
		t := template.Must(template.New("response").Parse(responseTextTemplate))
		err := t.Execute(w, response)
		if err != nil {
			return fmt.Errorf("failed to execute template: %w", err)
		}

		return nil

	case JsonOutputFormat:
		// print the response as json same as PrettyJsonOutputFormat but without indentation
		b, err := json.Marshal(response)
		if err != nil {
			return fmt.Errorf("failed to marshal response: %w", err)
		}
		_, err = w.Write(b)
		if err != nil {
			return fmt.Errorf("failed to write response: %w", err)
		}

		return nil
	}

	return ErrUnknownOutputFormat
}

type ConfigLoader func(path string) (*Config, error)

func defaultEnvLoader(path string) (*Config, error) {
	if path == "" {
		// take the current working directory and first check if there is a .env file
		// if not, check if there is a whatsapp.env file
		// if not, return an error
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current working directory: %w", err)
		}
		path = filepath.Join(cwd, ".env")

		// check if the .env file exists
		_, err = os.Stat(path)
		if err != nil {
			// if the file does not exist, check if there is a whatsapp.env file
			path = filepath.Join(cwd, "whatsapp.env")
			_, err = os.Stat(path)
			if err != nil {
				// if the file does not exist, return an error
				return nil, fmt.Errorf("failed to find .env or whatsapp.env file: %w", err)
			}
		}
	}
	// load the config from the env file
	config, err := loadConfigFromEnvFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load config from env file: %w", err)
	}

	return config, nil
}

const (
	envWhatsappApiBaseURL     = "WHATSAPP_CLOUD_API_BASE_URL"
	envWhatsappApiVersion     = "WHATSAPP_CLOUD_API_VERSION"
	envWhatsappApiPhoneID     = "WHATSAPP_CLOUD_API_PHONE_ID"
	envWhatsappBusinessAccID  = "WHATSAPP_CLOUD_BUSINESS_ACCOUNT_ID"
	envWhatsappApiAccessToken = "WHATSAPP_CLOUD_API_ACCESS_TOKEN"
)

// loadConfigFromEnvFile loads the config from the env file
func loadConfigFromEnvFile(path string) (*Config, error) {
	// These are the required env variables we are looking for
	// if not found the default value will be used which is an empty string
	// for all except the baseURL which is set to whatsapp.BaseURL and
	// the version which is set to "v16.0"
	var baseURL, version, phoneID, businessAccID, accessToken string
	values, err := dotenv.Read(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read env file (%s): %w", path, err)
	}

	// check if the env variable exists
	if v, ok := values[envWhatsappApiBaseURL]; ok {
		baseURL = v
	} else {
		baseURL = whatsapp.BaseURL
	}

	if v, ok := values[envWhatsappApiVersion]; ok {
		version = v
	} else {
		version = "v16.0"
	}

	if v, ok := values[envWhatsappApiPhoneID]; ok {
		phoneID = v
	}

	if v, ok := values[envWhatsappBusinessAccID]; ok {
		businessAccID = v
	}

	if v, ok := values[envWhatsappApiAccessToken]; ok {
		accessToken = v
	}

	// return the config
	return &Config{
		BaseURL:           baseURL,
		Version:           version,
		PhoneID:           phoneID,
		BusinessAccountID: businessAccID,
		AccessToken:       accessToken,
	}, nil
}
