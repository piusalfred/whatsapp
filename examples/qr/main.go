/*
 *  Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy of this software
 *  and associated documentation files (the “Software”), to deal in the Software without restriction,
 *  including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
 *  and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
 *  subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all copies or substantial
 *  portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
 *  LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
 *  IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
 *  WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
 *  SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/piusalfred/libwhatsapp/config"
	whttp "github.com/piusalfred/libwhatsapp/pkg/http"
	"github.com/piusalfred/libwhatsapp/qrcode"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	configReader := config.ReaderFunc(func(ctx context.Context) (*config.Config, error) {
		return &config.Config{
			BaseURL:       "https://api.example.com",
			APIVersion:    "v1",
			PhoneNumberID: "12345",
			AccessToken:   "your-access-token",
		}, nil
	})

	var middlewares []whttp.Middleware[any]

	clientOptions := []whttp.CoreClientOption[any]{
		whttp.WithCoreClientHTTPClient[any](http.DefaultClient),
		whttp.WithCoreClientRequestInterceptor[any](
			func(ctx context.Context, req *http.Request) error {
				logger.Info("Request Intercepted", slog.String("url", req.URL.String()))
				return nil
			},
		),
		whttp.WithCoreClientResponseInterceptor[any](
			func(ctx context.Context, resp *http.Response) error {
				logger.Info("Response Intercepted", slog.Int("status", resp.StatusCode))
				return nil
			},
		),
		whttp.WithCoreClientMiddlewares(middlewares),
	}

	coreClient := whttp.NewSender[any](clientOptions...)
	client := qrcode.NewBaseClient(coreClient, configReader)
	ctx := context.Background()

	createReq := &qrcode.CreateRequest{
		PrefilledMessage: "Hello, world!",
		ImageFormat:      qrcode.ImageFormatPNG,
	}
	createResp, err := client.Create(ctx, createReq)
	if err != nil {
		fmt.Printf("Failed to create QR: %v\n", err)
	} else {
		fmt.Printf("Created QR Code: %v\n", createResp)
	}

	qrCodeID := createResp.Code
	getResp, err := client.Get(ctx, qrCodeID)
	if err != nil {
		fmt.Printf("Failed to get QR code: %v\n", err)
	} else {
		fmt.Printf("Fetched QR Code: %+v\n", getResp)
	}

	updateReq := &qrcode.UpdateRequest{
		QRCodeID:         qrCodeID,
		PrefilledMessage: "Updated message",
		ImageFormat:      qrcode.ImageFormatSVG,
	}
	updateResp, err := client.Update(ctx, updateReq)
	if err != nil {
		fmt.Printf("Failed to update QR: %v\n", err)
	} else {
		fmt.Printf("Updated QR Code success: %v\n", updateResp.Success)
	}

	listResp, err := client.List(ctx)
	if err != nil {
		fmt.Printf("Failed to list QR codes: %v\n", err)
	} else {
		fmt.Printf("List of QR Codes: %+v\n", listResp.Data)
	}

	deleteResp, err := client.Delete(ctx, qrCodeID)
	if err != nil {
		fmt.Printf("Failed to delete QR code: %v\n", err)
	} else {
		fmt.Printf("Deleted QR Code success: %v\n", deleteResp.Success)
	}
}
