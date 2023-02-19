package cli

import (
	"fmt"

	"github.com/piusalfred/whatsapp/qrcodes"
)

// commands:
//   whatsapp qrcodes create
//   whatsapp qrcodes get
//   whatsapp qrcodes list
//   whatsapp qrcodes delete

type QrcodesCommand struct {
	Create QrCodeCreateCommand `cmd:"" name:"create" help:"create a new qr code"`
	Get    QrCodeGetCommand    `cmd:"" name:"get" help:"get a qr code"`
	List   QrCodeListCommand   `cmd:"" name:"list" help:"list all qr codes"`
	Delete QrCodeDeleteCommand `cmd:"" name:"delete" help:"delete a qr code"`
}

type QrCodeCreateCommand struct {
	Message     string `name:"message" short:"M" help:"message to be displayed on the qr code"`
	ImageFormat string `name:"image-format" short:"I" help:"image format of the qr code" default:"png"`
}

func (cmd *QrCodeCreateCommand) Run(ctx *Context) error {
	resp, err := qrcodes.Create(
		ctx.ctx, ctx.http,
		&qrcodes.CreateRequest{
			BaseURL:          ctx.BaseURL,
			PhoneID:          ctx.PhoneID,
			AccessToken:      ctx.AccessToken,
			PrefilledMessage: cmd.Message,
			ImageFormat:      qrcodes.ImageFormat(cmd.ImageFormat),
		},
	)
	if err != nil {
		return err
	}

	fmt.Printf("qr code created successfully: %+v\n", resp)
	return nil
}

type QrCodeGetCommand struct {
	ImageID string `name:"id" short:"i" help:"image id of the qr code"`
}

func (cmd *QrCodeGetCommand) Run(ctx *Context) error {
	resp, err := qrcodes.Get(
		ctx.ctx, ctx.http, ctx.BaseURL,
		ctx.PhoneID, ctx.AccessToken,
		cmd.ImageID,
	)
	if err != nil {
		return err
	}

	fmt.Printf("qr code retrieved successfully: %+v\n", resp)
	return nil
}

type QrCodeListCommand struct {
}

func (cmd *QrCodeListCommand) Run(ctx *Context) error {
	resp, err := qrcodes.List(
		ctx.ctx, ctx.http, ctx.BaseURL,
		ctx.PhoneID, ctx.AccessToken,
	)

	if err != nil {
		return err
	}

	fmt.Printf("qr codes retrieved successfully: %+v\n", resp)

	return nil
}

type QrCodeDeleteCommand struct {
	ImageID string `name:"id" short:"i" help:"image id of the qr code"`
}

func (cmd *QrCodeDeleteCommand) Run(ctx *Context) error {
	resp, err := qrcodes.Delete(
		ctx.ctx, ctx.http, ctx.BaseURL,
		ctx.PhoneID, ctx.AccessToken,
		cmd.ImageID,
	)

	if err != nil {
		return err
	}

	fmt.Printf("qr code deleted successfully: %+v\n", resp)
	return nil
}
