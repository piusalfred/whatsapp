package whatsapp

import (
	"testing"
)

func TestMessageFn(t *testing.T) {
	// type args struct {
	// 	ctx     context.Context
	// 	client  *http.Client
	// 	url     string
	// 	token   string
	// 	message *Message
	// }
	// tests := []struct {
	// 	name    string
	// 	args    args
	// 	want    *MessageResponse
	// 	wantErr bool
	// }{
	// 	{
	// 		name: "TestMessageFn",
	// 		args: args{
	// 			ctx:     context.Background(),
	// 			client:  http.DefaultClient,
	// 			url:     "https://api.whatsapp.com/v1/messages",
	// 		},
	// 	},
	// }
	// for _, tt := range tests {
	// 	t.Run(tt.name, func(t *testing.T) {
	// 		got, err := MessageFn(tt.args.ctx, tt.args.client, tt.args.url, tt.args.token, tt.args.message)
	// 		if (err != nil) != tt.wantErr {
	// 			t.Errorf("MessageFn() error = %v, wantErr %v", err, tt.wantErr)
	// 			return
	// 		}
	// 		if !reflect.DeepEqual(got, tt.want) {
	// 			t.Errorf("MessageFn() = %v, want %v", got, tt.want)
	// 		}
	// 	})
	// }

	// message := &Message{
	// 	Product: "whatsapp",
	// 	To:      "255767001828",
	// 	Type:    "template",
	// 	Template: &MessageTemplate{
	// 		Name: "hello_world",
	// 		Language: TemplateLanguage{
	// 			Code: "en_US",
	// 		},
	// 	},
	// }

	// token := "EAALLrT0ok6UBAA0UvvrIZBFcfbCqGgqn6y5Lrd0arSrPNh6sDZCi1UBFPcDxLEJDeWackAHQdJXlN7FEUipZAzErMZCKGf2vnF0J0eotDH2u3PpfliCemLkxQUq8WfdrQNbqI7LdBggEBfkA7skLvtALEkviOzgGElhy3ziZAIjENgMyAZBazvAURXLvt4EObx6NEzqRZAvwQZDZD"
	// baseURL := "https://graph.facebook.com/v15.0/114425371552711/messages"

	// resp, err := MessageFn(context.TODO(), http.DefaultClient, baseURL, token, message)
	// if err != nil {
	// 	t.Errorf("MessageFn() error = %v", err)
	// 	return
	// }

	// t.Logf("MessageFn() = %v", resp)
}
