//  Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
//
//  Permission is hereby granted, free of charge, to any person obtaining a copy of this software
//  and associated documentation files (the “Software”), to deal in the Software without restriction,
//  including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
//  and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
//  subject to the following conditions:
//
//  The above copyright notice and this permission notice shall be included in all copies or substantial
//  portions of the Software.
//
//  THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
//  LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
//  IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
//  WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
//  SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package examples

import (
	"context"
	"net"
	"net/http"
	"runtime"
	"time"

	"github.com/joho/godotenv"

	"github.com/piusalfred/whatsapp/config"
)

func LoadConfigFromFile(filepath string) config.ReaderFunc {
	fn := func(ctx context.Context) (*config.Config, error) {
		values, err := godotenv.Read(filepath)
		if err != nil {
			return nil, err
		}

		conf := &config.Config{
			BaseURL:           values["WHATSAPP_CLOUD_API_BASE_URL"],
			APIVersion:        values["WHATSAPP_CLOUD_API_API_VERSION"],
			AccessToken:       values["WHATSAPP_CLOUD_API_ACCESS_TOKEN"],
			PhoneNumberID:     values["WHATSAPP_CLOUD_API_PHONE_NUMBER_ID"],
			BusinessAccountID: values["WHATSAPP_CLOUD_API_BUSINESS_ACCOUNT_ID"],
			AppSecret:         values["WHATSAPP_CLOUD_API_APP_SECRET"],
			AppID:             values["WHATSAPP_CLOUD_API_APP_ID"],
			SecureRequests:    values["WHATSAPP_CLOUD_API_SECURE_REQUESTS"] == "true",
		}

		return conf, nil
	}

	return fn
}

func HttpClient() *http.Client {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ForceAttemptHTTP2:     true,
		MaxIdleConnsPerHost:   runtime.GOMAXPROCS(0) + 1,
	}

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   60 * time.Second,
	}

	return httpClient
}
