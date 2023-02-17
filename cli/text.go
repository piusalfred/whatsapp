package cli

import (
	"fmt"
	"io"
	"net/http"

	"github.com/piusalfred/whatsapp"
	"github.com/urfave/cli/v2"
)

func NewSendCommand(client *http.Client, logger io.Writer) *Command {
	fgs := []cli.Flag{
		&cli.StringFlag{
			Name:    "phone",
			Aliases: []string{"p"},
			Usage:   "phone number to send to",
		},
	}

	sendSubCommandParams := &CommandParams{
		Name:        "send",
		Usage:       "send a message",
		Description: "send a message to a phone number",
		Flags:       fgs,
		SubCommands: []*cli.Command{
			SendTextCommand(client, logger).C,
		},
	}

	return NewCommand(sendSubCommandParams)
}

func SendTextCommand(client *http.Client, logger io.Writer) *Command {
	fgs := []cli.Flag{
		&cli.StringFlag{
			Name:    "message",
			Aliases: []string{"m"},
			Usage:   "message to send",
		},
		&cli.BoolFlag{
			Name:    "preview-url",
			Aliases: []string{"u"},
			Usage:   "allow to preview url",
		},
	}

	sendTextActionFunc := func(c *cli.Context) error {
		// request params are passed into the context
		// so we can access them from the context
		params, ok := c.Context.Value("params").(*RequestParams)
		if !ok {
			return fmt.Errorf("could not get request params")
		}
		baseURL := c.String(flagBaseURL)
		fmt.Printf("base url: %s\n", baseURL)
		message := c.String("message")
		phone := c.String("phone")
		previewURL := c.Bool("preview-url")
		request := &whatsapp.SendTextRequest{
			Recipient:     phone,
			Message:       message,
			PreviewURL:    previewURL,
			BaseURL:       params.BaseURL,
			AccessToken:   params.AccessToken,
			PhoneNumberID: params.ID,
			ApiVersion:    params.ApiVersion,
		}
		response, err := whatsapp.SendText(c.Context, client, request)
		if err != nil {
			return err
		}

		fmt.Println(response)
		return nil
	}

	sendTextSubCommandParams := &CommandParams{
		Name:        "text",
		Usage:       "send a text message",
		Description: "send a text message to a phone number",
		Flags:       fgs,
		Action:      sendTextActionFunc,
	}

	return NewCommand(sendTextSubCommandParams)
}
