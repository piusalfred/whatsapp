/*
 * MIT License
 *
 * Copyright (c) 2025 Pius Alfred
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/piusalfred/whatsapp"
	"github.com/piusalfred/whatsapp/config"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
	"github.com/piusalfred/whatsapp/user"
)

func main() {
	reader := config.ReaderFunc(func(ctx context.Context) (*config.Config, error) {
		conf := &config.Config{
			BaseURL:           whatsapp.BaseURL,
			APIVersion:        "",
			AccessToken:       "",
			PhoneNumberID:     "",
			BusinessAccountID: "",
			AppSecret:         "",
			SecureRequests:    false,
		}

		return conf, nil
	})

	ctx := context.Background()

	sender := whttp.NewSender[user.BlockBaseRequest]()
	blocker := user.NewBlockClient(reader, sender)

	spammers := []string{"1234567890", "1234567891", "1234567892"}
	resp, err := blocker.Block(ctx, spammers)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(resp)

	spammers1 := []string{"1234567890", "1234567891"}
	resp1, err1 := blocker.Unblock(ctx, spammers1)
	if err1 != nil {
		panic(err1)
	}

	fmt.Println(resp1)

	resp2, err2 := blocker.ListBlocked(ctx, &user.ListBlockedUsersOptions{})
	if err2 != nil {
		panic(err2)
	}

	fmt.Println(resp2)
}
