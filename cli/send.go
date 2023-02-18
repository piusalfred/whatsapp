package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/piusalfred/whatsapp"
)

type (
	SendCommand struct {
		SendTextCommand SendTextCommand `cmd:"" name:"text" help:"send a text message"`
	}

	SendTextCommand struct {
		Recipient  string `name:"to" help:"the recipient of the message" type:"string"`
		Message    string `name:"message" short:"m" help:"the text to send" type:"string"`
		PreviewURL bool   `name:"preview-url" short:"P" help:"preview the url" type:"bool"`
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
	nctx, cancel := context.WithTimeout(ctx.ctx, time.Duration(ctx.Timeout)*time.Second)
	defer cancel()
	resp, err := whatsapp.SendText(nctx, ctx.http, req)
	if err != nil {
		return err
	}
	fmt.Println(resp)
	return nil
}
