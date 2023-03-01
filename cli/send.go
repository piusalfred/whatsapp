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
	"fmt"
	"time"

	"github.com/piusalfred/whatsapp"
)

type (
	SendCommand struct {
		SendTextCommand     SendTextCommand     `cmd:"" name:"text" help:"send a text message"`
		SendLocationCommand SendLocationCommand `cmd:"" name:"location" help:"send a location message"`
		SendTemplateCommand SendTemplateCommand `cmd:"" name:"template" help:"send a template message"`
	}

	SendTextCommand struct {
		Recipient  string `name:"to" help:"the recipient of the message" type:"string"`
		Message    string `name:"message" short:"m" help:"the text to send" type:"string"`
		PreviewURL bool   `name:"preview-url" short:"P" help:"preview the url" type:"bool" default:"false"`
	}

	SendLocationCommand struct {
		Recipient string  `name:"to" help:"the recipient of the message" type:"string"`
		Latitude  float64 `name:"lat" help:"the latitude of the location" type:"float64"`
		Longitude float64 `name:"long" help:"the longitude of the location" type:"float64"`
		Name      string  `name:"name" help:"the name of the location" type:"string"`
		Address   string  `name:"address" help:"the address of the location" type:"string"`
	}

	SendTemplateCommand struct {
		Recipient      string `name:"to" help:"the recipient of the message" type:"string"`
		TemplateName   string `name:"template-name" help:"the name of the template" type:"string" default:"hello_world"`
		LanguageCode   string `name:"language-code" help:"the language code of the template" type:"string" default:"en_US"`
		LanguagePolicy string `name:"language-policy" help:"the language policy of the template" type:"string"`
	}
)

func (cmd *SendCommand) Run(globals *Context) error {
	return nil
}

func (cmd *SendTextCommand) Run(ctx *Context) error {
	configurations(ctx)

	req := &whatsapp.SendTextRequest{
		Recipient:     cmd.Recipient,
		Message:       cmd.Message,
		PreviewURL:    cmd.PreviewURL,
		ApiVersion:    ctx.ApiVersion,
		BaseURL:       ctx.BaseURL,
		PhoneNumberID: ctx.PhoneID,
		AccessToken:   ctx.AccessToken,
	}

	nctx, cancel := context.WithTimeout(context.TODO(), time.Duration(ctx.Timeout)*time.Second)
	defer cancel()
	resp, err := whatsapp.SendText(nctx, ctx.http, req)
	if err != nil {
		return fmt.Errorf("error sending text: %s", err.Error())
	}

	return printResponse(ctx.logger, resp, OutputFormat(ctx.Output))
}

func configurations(ctx *Context) error {
	configInFile, err := ctx.loader(ctx.ConfigPath)
	if err != nil {
		if ctx.Debug {
			ctx.logger.Write([]byte(fmt.Sprintf("error looking for config in current dir: %v", err)))
		}
	}

	if configInFile != nil {
		if ctx.BaseURL == "" {
			ctx.BaseURL = configInFile.BaseURL
		}

		if ctx.AccessToken == "" {
			ctx.AccessToken = configInFile.AccessToken
		}

		if ctx.PhoneID == "" {
			ctx.PhoneID = configInFile.PhoneID
		}

		if ctx.ApiVersion == "" {
			ctx.ApiVersion = configInFile.Version
		}

		if ctx.WabaID == "" {
			ctx.WabaID = configInFile.BusinessAccountID
		}
	}
	return err
}

func (cmd *SendLocationCommand) Run(ctx *Context) error {
	configurations(ctx)

	req := &whatsapp.SendLocationRequest{
		Recipient:     cmd.Recipient,
		Latitude:      cmd.Latitude,
		Longitude:     cmd.Longitude,
		Name:          cmd.Name,
		Address:       cmd.Address,
		ApiVersion:    ctx.ApiVersion,
		BaseURL:       ctx.BaseURL,
		PhoneNumberID: ctx.PhoneID,
		AccessToken:   ctx.AccessToken,
	}
	nctx, cancel := context.WithTimeout(context.TODO(), time.Duration(ctx.Timeout)*time.Second)
	defer cancel()
	resp, err := whatsapp.SendLocation(nctx, ctx.http, req)
	if err != nil {
		return fmt.Errorf("error sending location: %w", err)
	}

	return printResponse(ctx.logger, resp, OutputFormat(ctx.Output))
}

func (cmd *SendTemplateCommand) Run(ctx *Context) error {
	configurations(ctx)

	req := &whatsapp.SendTemplateRequest{
		Recipient:              cmd.Recipient,
		TemplateName:           cmd.TemplateName,
		TemplateLanguageCode:   cmd.LanguageCode,
		TemplateLanguagePolicy: cmd.LanguagePolicy,
		ApiVersion:             ctx.ApiVersion,
		BaseURL:                ctx.BaseURL,
		PhoneNumberID:          ctx.PhoneID,
		AccessToken:            ctx.AccessToken,
	}
	nctx, cancel := context.WithTimeout(context.TODO(), time.Duration(ctx.Timeout)*time.Second)
	defer cancel()
	resp, err := whatsapp.SendTemplate(nctx, ctx.http, req)
	if err != nil {
		return fmt.Errorf("error sending template: %w", err)
	}

	return printResponse(ctx.logger, resp, OutputFormat(ctx.Output))
}
