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
		Send    SendCommand    `cmd:"" name:"send" help:"send different types of messages like text, image, video, audio, document, location, vcard, template, sticker, and file"`
		QrCodes QrcodesCommand `cmd:"" name:"qrcodes" help:"manage qr codes"`
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
