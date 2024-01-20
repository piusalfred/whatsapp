# whatsapp

Configurable easy to use Go wrapper for the WhatsApp Cloud API.


## set up
- Have golang installed
- Register as a Meta Developer here https://developers.facebook.com/docs/development/register 
- Create an application here https://developers.facebook.com/docs/development/create-an-app and configure
it to enable access to WhatsApp Business Cloud API and Webhooks.

- You can manage your apps here https://developers.facebook.com/apps/

- From Whatsapp Developer Dashboard you can try and send a test message to your phone number.
to be sure that everything is working fine before you start using this api. Also you need to
reply to that message to be able to send other messages.

- Go to [examples/base](examples/base) then create `.env` file that looks like [examples/base/.envrc](examples/base/.envrc)
 and add your credentials there.

- Run `make run` and wait to receive a message on your phone. Make sure you have sent the template message
first from the Whatsapp Developer Dashboard.

## usage

1. [Messages](##messages) ✅
   * [1.1 Normal Messages](###11-normal-messages) 🚧
   * [1.2 Reply Messages](###12-reply-messages) 🚧
   * [1.3 Media Messages](###13-media-messages) 🚧
   * [1.4 Interactive Messages](###14-interactive-messages) 🚧
   * [1.5 Template Messages](###15-template-messages) 🚧
     + [1.5.1 Text-based Message Templates](####151-text-based-message-templates) 🚧
     + [1.5.2 Media-based Message Templates](#####152-media-based-message-templates) 🚧
     + [1.5.3 Interactive Message Templates](#####153-interactive-message-templates) 🚧
     + [1.5.4 Location-based Message Templates](####154-location-based-message-templates) 🚧
     + [1.5.5 Authentication Templates with OTP Buttons](#####155-authentication-templates-with-otp-buttons) 🚧
     + [1.5.6 Multi-Product Message Templates](#####156-multi-product-message-templates) 🚧
2. [Webhooks](##2-webhooks) ✅
   * [2.1 Verify Requests](####21-verify-requests) 🚧
   * [2.2 Listen To Requests](####22-listen-to-requests) 🚧
3. [Health Status](##3-health-status) 🚧
4. [Templates Management](##4-templates-management) ✅
5. [PhoneNumbers](##5-phonenumbers) 🚧
   * [5.1 Register](###51-register) 🚧
   * [5.2 Delete](###52-delete) 🚧
   * [5.3 Set PIN](###53-set-pin) 🚧
6. [QR Codes Management](##6-qr-codes-management) ✅
7. [Media Management](##7-media-management) ✅
   * [7.1 Upload](###71-upload) 🚧
   * [7.2 Delete](###72-delete) 🚧
   * [7.3 List](###73-list) 🚧
   * [7.4 Download](###74-download) 🚧
   * [7.5 Retrieve Information](###75-retrieve-information) 🚧
8. [WhatsApp Business Account](##8-whatsapp-business-account) ✅
9. [WhatsApp Business Encryption](##9-whatsapp-business-encryption) ✅
10. [Flows](##10-flows) 🚧



### messages

Create a `client` instance and use it to send messages. It uses `config.Reader` to read
configuration values from a source. You can implement your own `config.Reader`.The `dotEnvReader`
is an example of a `config.Reader` that reads configuration values from a `.env` file.


```go
package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/joho/godotenv"
	"github.com/piusalfred/whatsapp"
	"github.com/piusalfred/whatsapp/pkg/config"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
	"github.com/piusalfred/whatsapp/pkg/models"
)

var _ config.Reader = (*dotEnvReader)(nil)

type dotEnvReader struct {
	filePath string
}

func (d *dotEnvReader) Read(ctx context.Context) (*config.Values, error) {
	vm, err := godotenv.Read(d.filePath)
	if err != nil {
		return nil, err
	}

	return &config.Values{
		BaseURL:           vm["BASE_URL"],
		Version:           vm["VERSION"],
		AccessToken:       vm["ACCESS_TOKEN"],
		PhoneNumberID:     vm["PHONE_NUMBER_ID"],
		BusinessAccountID: vm["BUSINESS_ACCOUNT_ID"],
	}, nil
}

func initBaseClient(ctx context.Context) (*whatsapp.Client, error) {
	reader := &dotEnvReader{filePath: ".env"}
	b, err := whatsapp.NewClient(ctx, reader,
		whatsapp.WithBaseClientOptions(
			[]whttp.BaseClientOption{
				whttp.WithHTTPClient(http.DefaultClient),
				whttp.WithRequestHooks(),
				whttp.WithResponseHooks(),
				whttp.WithSendMiddleware(),
			},
		),
		whatsapp.WithSendMiddlewares(),
	)
	if err != nil {
		return nil, err
	}

	return b, nil
}

```
### 1.1 Normal Messages
An example to send a text message
```go
func send(ctx context.Context)error{
	client, err := initBaseClient(ctx)
	response, err := client.Text(ctx, &whatsapp.RequestParams{
		ID:        "",  // Optional
		Metadata:  map[string]string{ // Optional -for stuffs like observability
			"key": "value",
			"context": "demo",
		},
		Recipient: "+2557XXXXXXX",
		ReplyID:   "",// Put the message ID here if you want to reply to that message.
	}, &models.Text{
		Body:       "Hello World From github.com/piusalfred/whatsapp",
		PreviewURL: true,
	})
	if err != nil {
		return err
	}
	
	fmt.Printf("\n%+v\n", response)
	
	return nil
}
```
There are other client methods that you can use to send other types of messages.
like `client.Location` for sending location messages, `client.Image` for sending image messages
and so on.
### 1.2 Reply Messages
### 1.3 Media Messages
### 1.4 Interactive Messages
### 1.5 Template Messages
#### 1.5.1 Text-based Message Templates
#### 1.5.2 Media-based Message Templates
#### 1.5.3 Interactive Message Templates
#### 1.5.4 Location-based Message Templates
#### 1.5.5 Authentication Templates with OTP Buttons
#### 1.5.6 Multi-Product Message Templates

## Webhooks
### 2.1 Verify Requests
### 2.2 Listen To Requests

## Health Status

## Templates Management

## PhoneNumbers
### 5.1 Register
### 5.2 Delete
### 5.3 Set PIN

## QR Codes Management

## Media Management
### 7.1 Upload
### 7.2 Delete
### 7.3 List
### 7.4 Download
### 7.5 Retrieve Information

## WhatsApp Business Account

## WhatsApp Business Encryption


