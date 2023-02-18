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
	req := &whatsapp.SendTextRequest{
		Recipient:     cmd.Recipient,
		Message:       cmd.Message,
		PreviewURL:    cmd.PreviewURL,
		ApiVersion:    ctx.ApiVersion,
		BaseURL:       ctx.BaseURL,
		PhoneNumberID: ctx.PhoneID,
		AccessToken:   ctx.AccessToken,
	}

	fmt.Printf("Access Token: %s\n", ctx.AccessToken)
	nctx, cancel := context.WithTimeout(ctx.ctx, time.Duration(ctx.Timeout)*time.Second)
	defer cancel()
	resp, err := whatsapp.SendText(nctx, ctx.http, req)
	if err != nil {
		return err
	}
	fmt.Println(resp)
	return nil
}

func (cmd *SendLocationCommand) Run(ctx *Context) error {
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
	nctx, cancel := context.WithTimeout(ctx.ctx, time.Duration(ctx.Timeout)*time.Second)
	defer cancel()
	resp, err := whatsapp.SendLocation(nctx, ctx.http, req)
	if err != nil {
		return err
	}
	fmt.Println(resp)
	return nil
}

func (cmd *SendTemplateCommand) Run(ctx *Context) error {
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
	nctx, cancel := context.WithTimeout(ctx.ctx, time.Duration(ctx.Timeout)*time.Second)
	defer cancel()
	resp, err := whatsapp.SendTemplate(nctx, ctx.http, req)
	if err != nil {
		return err
	}
	fmt.Println(resp)
	return nil
}
