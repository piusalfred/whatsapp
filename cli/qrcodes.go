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
		context.TODO(), ctx.http,
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
		context.TODO(), ctx.http, ctx.BaseURL,
		ctx.PhoneID, ctx.AccessToken,
		cmd.ImageID,
	)
	if err != nil {
		return err
	}

	fmt.Printf("qr code retrieved successfully: %+v\n", resp)
	return nil
}

type QrCodeListCommand struct{}

func (cmd *QrCodeListCommand) Run(ctx *Context) error {
	resp, err := qrcodes.List(
		context.TODO(), ctx.http, ctx.BaseURL,
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
		context.TODO(), ctx.http, ctx.BaseURL,
		ctx.PhoneID, ctx.AccessToken,
		cmd.ImageID,
	)
	if err != nil {
		return err
	}

	fmt.Printf("qr code deleted successfully: %+v\n", resp)
	return nil
}
