# whatsapp

# SETUP

## pre requisites

To be able to test/use this api you need Access Token, Phone number ID and Business ID. For that
tou need to register as a Meta Developer. You can register here https://developers.facebook.com/docs/development/register

Then create an application here https://developers.facebook.com/docs/development/create-an-app and configre
it to enable access to WhatsApp Business Cloud API.

You can manage your apps here https://developers.facebook.com/apps/

From Whatsapp Developer Dashboard you can try and send a test message to your phone number.
to be sure that everything is working fine before you start using this api.

When all the above is done you can start using this api.


## messaging

```go

func main() {
	req := &whatsapp.SendTextRequest{
		Recipient:     "1234567890", 
		Message:       "hello world", 
		PreviewURL:    true, 
		ApiVersion:    "v16.0", 
		BaseURL:       whatsapp.BaseURL, 
		PhoneNumberID: "1234567890"",
		AccessToken:   "EAAH...ZD",
	}
	
	resp, err := whatsapp.SendText(context.TODO(), http.DefaultClient, req)
	if err != nil {
		return err
	}
	
	fmt.Println(resp)
}

```


## cli
The library has an accompanying cli tool that can be used to send messages to your phone number.
The cli tool can be used to send text, image, video, audio, document and sticker messages.
It is not sophisticated,but it can be used to test the api.

You can navigate to cmd folder and run `go build -o whatsapp whatsapp.go` to build the cli tool.

Here is the output of `./whatsapp send text --help` on how to use the cli tool to send text messages.

```text


Usage: whatsapp send text

send a text message

Flags:
  -h, --help                                      Show context-sensitive help.
      --config=".env"                             Location of client config files
  -D, --debug                                     Enable debug mode
  -L, --log-level="info"                          Set the logging level (debug|info|warn|error|fatal)
  -O, --output="text"                             Output format (json|text)
  -V, --api-version="v16.0"                       the version of Whatsapp Cloud API to use
  -b, --base-url="https://graph.facebook.com/"    the base URL of Whatsapp Cloud API to use
  -p, --phone-id=STRING                           phone ID of Whatsapp Cloud API to use
  -w, --waba-id=STRING                            whatsapp business account id
  -T, --access-token=STRING                       access token of Whatsapp Cloud API to use
  -t, --timeout=30                                http timeout for making api calls

      --to=STRING                                 the recipient of the message
  -m, --message=STRING                            the text to send
  -P, --preview-url                               preview the url
```

Some things above like BaseURL have the default values set. So you can skip them.
To send a simple text message you can run

`./whatsapp send text  -T=<your access token> -p=<your phone id> -m="hello world" --to=<recipient-phone-number"`

## webhooks example

```go

package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/piusalfred/whatsapp/models"
	"github.com/piusalfred/whatsapp/webhooks"

	// julieschmidt/httprouter
	"github.com/julienschmidt/httprouter"
)

func main() {
	var options []webhooks.ListenerOption

	// Set your very own notification error handler
	options = append(options, webhooks.WithNotificationErrorHandler(
		func(ctx context.Context, request *http.Request, err error) *webhooks.Response {
			fmt.Printf("error received in notification: %v\n", err) //nolint:forbidigo

			return &webhooks.Response{
				StatusCode: http.StatusInternalServerError,
				Headers: map[string]string{
					"Content-Type": "text/plain",
				},
				Body: []byte(err.Error()),
				Skip: false,
			}
		}))

	// Set your very own subscription verifier
	options = append(options, webhooks.WithSubscriptionVerifier(
		func(ctx context.Context, request *webhooks.VerificationRequest) error {
			fmt.Printf("subscription verification request: %+v\n", request) //nolint:forbidigo
			if request.Mode == "subscribe" && request.Challenge == "challenge" {
				return nil
			}
			return fmt.Errorf("invalid subscription verification request\n")
		}))

	// Set other Listener options here as you wish ........

	// init the listener
	listener := webhooks.NewEventListener(options...)

	// What to do when we receive a reaction?
	// all the logic goes here
	listener.OnMessageReaction(func(ctx context.Context, nctx *webhooks.NotificationContext,
		mctx *webhooks.MessageContext, reaction *models.Reaction,
	) error {
		// do something with the reaction
		fmt.Printf("reaction: %+v\n", reaction) //nolint:forbidigo

		return nil
	})

	// You want to handle media? Things like document, audio, video, image and sticker
	// Well add all your logic here.
	listener.OnMediaMessage(func(ctx context.Context, nctx *webhooks.NotificationContext,
		mctx *webhooks.MessageContext, media *models.MediaInfo,
	) error {
		// do something with the media
		fmt.Printf("media info: %+v\n", media) //nolint:forbidigo

		return nil
	})

	// What to do when we receive a notification of any type be it a media or a text you can handle it here
	// in a generic way where all are caught in one place
	// remember the notification error handler you set above? it will be called here to investigate
	// the error you return
	listener.GenericNotificationHandler(
		func(ctx context.Context, writer http.ResponseWriter,
			notification *webhooks.Notification,
		) error {
			// do something with the notification
			fmt.Printf("notification: %+v\n", notification) //nolint:forbidigo

			return nil
		})

	//// create a http server that listen to port 8080 and all notification on the path POST /webhooks
	//// will be handled by a handler created by the listener on path POST /generic_webhooks the generic
	//// handler will be used. on GET /webhooks the subscription verification will be handled

	// I have used httprouter here, but you can use any other router you wish
	router := httprouter.New()
	router.Handler(http.MethodPost, "/webhooks", listener.Handle())
	router.Handler(http.MethodPost, "/generic_webhooks", listener.GenericHandler())
	router.Handler(http.MethodGet, "/webhooks", listener.Verify())

	// start the server
	if err := http.ListenAndServe(":8080", router); err != nil {
		panic(err)
	}
}

```