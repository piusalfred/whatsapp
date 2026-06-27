//  Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
//
//  Permission is hereby granted, free of charge, to any person obtaining a copy of this software
//  and associated documentation files (the "Software"), to deal in the Software without restriction,
//  including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
//  and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
//  subject to the following conditions:
//
//  The above copyright notice and this permission notice shall be included in all copies or substantial
//  portions of the Software.
//
//  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
//  LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
//  IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
//  WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
//  SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package template_test

import (
	"testing"

	"github.com/piusalfred/whatsapp/internal/test"
	"github.com/piusalfred/whatsapp/message/media"
	"github.com/piusalfred/whatsapp/message/template"
)

func TestNewLocationHeader_Marshal(t *testing.T) {
	t.Parallel()

	test.AssertJSONMarshal(t, "location header",
		template.NewLocationHeader(),
		`{"type":"header","format":"LOCATION"}`,
	)
}

func TestNewLocationParameter_Marshal(t *testing.T) {
	t.Parallel()

	test.AssertJSONMarshal(t, "location parameter",
		template.NewLocationParameter(&media.Location{
			Latitude:  37.44211676562361,
			Longitude: -122.16155960083124,
			Name:      "Philz Coffee",
			Address:   "101 Forest Ave, Palo Alto, CA 94301",
		}),
		`{
			"type": "location",
			"location": {
				"latitude": 37.44211676562361,
				"longitude": -122.16155960083124,
				"name": "Philz Coffee",
				"address": "101 Forest Ave, Palo Alto, CA 94301"
			}
		}`,
	)
}

func TestLocationTemplate_Marshal(t *testing.T) {
	t.Parallel()

	// Simulates the send-time payload for a location template
	tmpl := &template.Template{
		Name: "order_delivery_update",
		Language: &template.Language{
			Policy: "deterministic",
			Code:   "en_US",
		},
		Components: []*template.Component{
			{
				Type: template.TemplateComponentTypeHeader,
				Parameters: []*template.Parameter{
					template.NewLocationParameter(&media.Location{
						Latitude:  37.44211676562361,
						Longitude: -122.16155960083124,
						Name:      "Philz Coffee",
						Address:   "101 Forest Ave, Palo Alto, CA 94301",
					}),
				},
			},
			{
				Type: template.TemplateComponentTypeBody,
				Parameters: []*template.Parameter{
					{Type: "text", Text: "Jane", Name: "customer_name"},
					{Type: "text", Text: "892104", Name: "order_number"},
				},
			},
		},
	}

	test.AssertJSONMarshal(t, "location template send payload", tmpl,
		`{
			"name": "order_delivery_update",
			"language": {"policy": "deterministic", "code": "en_US"},
			"components": [
				{
					"type": "header",
					"parameters": [
						{
							"type": "location",
							"location": {
								"latitude": 37.44211676562361,
								"longitude": -122.16155960083124,
								"name": "Philz Coffee",
								"address": "101 Forest Ave, Palo Alto, CA 94301"
							}
						}
					]
				},
				{
					"type": "body",
					"parameters": [
						{"type": "text", "text": "Jane", "parameter_name": "customer_name"},
						{"type": "text", "text": "892104", "parameter_name": "order_number"}
					]
				}
			]
		}`,
	)
}

func TestAuthTemplate_Marshal(t *testing.T) {
	t.Parallel()

	tmpl := template.NewAuthTemplate(&template.AuthTemplateRequest{
		Name:            "auth_otp",
		LanguageCode:    "en_US",
		LanguagePolicy:  "deterministic",
		OneTimePassword: "123456",
	})

	test.AssertJSONMarshal(t, "auth template", tmpl,
		`{
			"name": "auth_otp",
			"language": {"policy": "deterministic", "code": "en_US"},
			"components": [
				{
					"type": "body",
					"parameters": [
						{"type": "text", "text": "123456"}
					]
				},
				{
					"type": "button",
					"sub_type": "url",
					"parameters": [
						{"type": "text", "text": "123456"}
					]
				}
			]
		}`,
	)
}

func TestInteractiveTemplate_Marshal(t *testing.T) {
	t.Parallel()

	tmpl := template.NewInteractiveTemplate(
		"hello_world",
		&template.Language{Code: "en_US"},
		[]*template.Parameter{
			{Type: "text", Text: "Hello World"},
		},
		[]*template.Parameter{
			{Type: "text", Text: "Jane"},
		},
		nil,
	)

	test.AssertJSONMarshal(t, "interactive template", tmpl,
		`{
			"name": "hello_world",
			"language": {"code": "en_US"},
			"components": [
				{
					"type": "header",
					"parameters": [
						{"type": "text", "text": "Hello World"}
					]
				},
				{
					"type": "body",
					"parameters": [
						{"type": "text", "text": "Jane"}
					]
				}
			]
		}`,
	)
}

func TestLocationHeader_RoundTrip(t *testing.T) {
	t.Parallel()

	cmp := template.NewLocationHeader()
	test.AssertJSONRoundTrip(t, "location header", cmp)
}

func TestTemplate_RoundTrip(t *testing.T) {
	t.Parallel()

	tmpl := &template.Template{
		Name: "order_delivery_update",
		Language: &template.Language{
			Policy: "deterministic",
			Code:   "en_US",
		},
		Components: []*template.Component{
			template.NewLocationHeader(),
			{
				Type: template.TemplateComponentTypeBody,
				Text: "Your order is on the way!",
			},
		},
	}

	test.AssertJSONRoundTrip(t, "template", tmpl)
}
