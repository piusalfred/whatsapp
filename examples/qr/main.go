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

	"github.com/joho/godotenv"

	"github.com/piusalfred/whatsapp/config"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
	"github.com/piusalfred/whatsapp/qrcode"
)

func LoadConfigFromFile(filepath string) config.ReaderFunc {
	fn := func(ctx context.Context) (*config.Config, error) {
		err := godotenv.Load(filepath) // Load the .env file from the given path
		if err != nil {
			return nil, err
		}

		conf := &config.Config{
			BaseURL:           os.Getenv("WHATSAPP_CLOUD_API_BASE_URL"),
			APIVersion:        os.Getenv("WHATSAPP_CLOUD_API_API_VERSION"),
			AccessToken:       os.Getenv("WHATSAPP_CLOUD_API_ACCESS_TOKEN"),
			PhoneNumberID:     os.Getenv("WHATSAPP_CLOUD_API_PHONE_NUMBER_ID"),
			BusinessAccountID: os.Getenv("WHATSAPP_CLOUD_API_BUSINESS_ACCOUNT_ID"),
		}

		return conf, nil
	}

	return fn
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

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
		whttp.WithCoreClientMiddlewares(middlewares...),
	}

	coreClient := whttp.NewSender[any](clientOptions...)
	client := qrcode.NewBaseClient(coreClient, LoadConfigFromFile("api.env"))
	ctx := context.Background()

	createReq := &qrcode.CreateRequest{
		PrefilledMessage: " tsup  https://github.com/piusalfred/whatsapp mate",
		ImageFormat:      qrcode.ImageFormatPNG,
	}
	createResp, err := client.Create(ctx, createReq)
	if err != nil {
		fmt.Printf("Failed to create QR: %v\n", err)
		return
	} else {
		fmt.Printf("Created QR Code Link: %v\n", createResp.QRImageURL)
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
		PrefilledMessage: "please visit https://github.com/piusalfred/whatsapp to contribute",
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

	//deleteResp, err := client.Delete(ctx, qrCodeID)
	//if err != nil {
	//	fmt.Printf("Failed to delete QR code: %v\n", err)
	//} else {
	//	fmt.Printf("Deleted QR Code success: %v\n", deleteResp.Success)
	//}

	for i, datum := range listResp.Data {
		logger.LogAttrs(ctx, slog.LevelInfo, "qr code",
			slog.Int("index", i),
			slog.String("code", datum.Code),
			slog.String("messsage", datum.PrefilledMessage),
			slog.String("deep link", datum.DeepLinkURL),
		)
	}
}
