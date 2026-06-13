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

package interactive_test

import (
	"encoding/json"
	"testing"

	gcmp "github.com/google/go-cmp/cmp"

	"github.com/piusalfred/whatsapp/message/interactive"
	"github.com/piusalfred/whatsapp/message/media"
)

func TestCTAURLButton(t *testing.T) {
	t.Parallel()

	msg := interactive.CTAURLButton(&interactive.CTAURLRequest{
		DisplayText: "See Dates",
		URL:         "https://www.luckyshrub.com?clickID=kqDGWd24Q5TRwoEQTICY7W1JKoXvaZOXWAS7h1P76s0R7Paec4",
		Body:        "Tap the button below to see available dates.",
		Header: interactive.HeaderImage(&media.Image{
			Link: "https://www.luckyshrub.com/assets/lucky-shrub-banner-logo-v1.png",
		}),
		Footer: "Dates subject to change.",
	})

	got, _ := json.Marshal(msg)

	expected := &interactive.Message{}
	if err := json.Unmarshal([]byte(`{
		"type": "cta_url",
		"header": { "type": "image", "image": { "link": "https://www.luckyshrub.com/assets/lucky-shrub-banner-logo-v1.png" } },
		"body": { "text": "Tap the button below to see available dates." },
		"footer": { "text": "Dates subject to change." },
		"action": {
			"name": "cta_url",
			"parameters": {
				"display_text": "See Dates",
				"url": "https://www.luckyshrub.com?clickID=kqDGWd24Q5TRwoEQTICY7W1JKoXvaZOXWAS7h1P76s0R7Paec4"
			}
		}
	}`), expected); err != nil {
		t.Fatal(err)
	}

	want, _ := json.Marshal(expected)

	if !gcmp.Equal(got, want) {
		t.Errorf("CTA URL JSON mismatch:\n got:  %s\n want: %s", got, want)
	}
}

func TestList(t *testing.T) {
	t.Parallel()

	msg := interactive.List(&interactive.ListRequest{
		Body:   "Which shipping option do you prefer?",
		Button: "Shipping Options",
		Footer: "Lucky Shrub: Your gateway to succulents\u2122",
		Header: "Choose Shipping Option",
		Sections: []*interactive.Section{
			{
				Title: "I want it ASAP!",
				Rows: []*interactive.SectionRow{
					{ID: "priority_express", Title: "Priority Mail Express", Description: "Next Day to 2 Days"},
					{ID: "priority_mail", Title: "Priority Mail", Description: "1\u20133 Days"},
				},
			},
		},
	})

	got, _ := json.Marshal(msg)

	expected := &interactive.Message{}
	if err := json.Unmarshal([]byte(`{
		"type": "list",
		"header": { "type": "text", "text": "Choose Shipping Option" },
		"body": { "text": "Which shipping option do you prefer?" },
		"footer": { "text": "Lucky Shrub: Your gateway to succulents\u2122" },
		"action": {
			"button": "Shipping Options",
			"sections": [
				{
					"title": "I want it ASAP!",
					"rows": [
						{ "id": "priority_express", "title": "Priority Mail Express", "description": "Next Day to 2 Days" },
						{ "id": "priority_mail", "title": "Priority Mail", "description": "1\u20133 Days" }
					]
				}
			]
		}
	}`), expected); err != nil {
		t.Fatal(err)
	}

	want, _ := json.Marshal(expected)

	if !gcmp.Equal(got, want) {
		t.Errorf("List JSON mismatch:\n got:  %s\n want: %s", got, want)
	}
}

func TestReplyButtons(t *testing.T) {
	t.Parallel()

	msg := interactive.ReplyButtons(&interactive.ReplyButtonsRequest{
		Body:   "Hi Pablo! Your gardening workshop is scheduled for 9am tomorrow. Use the buttons if you need to reschedule. Thank you!",
		Footer: "Lucky Shrub: Your gateway to succulents!\u2122",
		Header: interactive.HeaderImage(&media.Image{ID: "2762702990552401"}),
		Buttons: []*interactive.ReplyButton{
			{ID: "change-button", Title: "Change"},
			{ID: "cancel-button", Title: "Cancel"},
		},
	})

	got, _ := json.Marshal(msg)

	expected := &interactive.Message{}
	if err := json.Unmarshal([]byte(`{
		"type": "button",
		"header": { "type": "image", "image": { "id": "2762702990552401" } },
		"body": { "text": "Hi Pablo! Your gardening workshop is scheduled for 9am tomorrow. Use the buttons if you need to reschedule. Thank you!" },
		"footer": { "text": "Lucky Shrub: Your gateway to succulents!\u2122" },
		"action": {
			"buttons": [
				{ "type": "reply", "reply": { "id": "change-button", "title": "Change" } },
				{ "type": "reply", "reply": { "id": "cancel-button", "title": "Cancel" } }
			]
		}
	}`), expected); err != nil {
		t.Fatal(err)
	}

	want, _ := json.Marshal(expected)

	if !gcmp.Equal(got, want) {
		t.Errorf("Reply Buttons JSON mismatch:\n got:  %s\n want: %s", got, want)
	}
}

func TestMediaCarousel_URLButtons(t *testing.T) {
	t.Parallel()

	msg := interactive.MediaCarousel("Here are our latest arrivals:", []*interactive.MediaCarouselCard{
		{
			HeaderType:     "image",
			HeaderLink:     "https://www.luckyshrub.com/assets/blue-echeveria.jpeg",
			BodyText:       "*Blue Echeveria*\n\nA rosette-shaped succulent with powdery blue leaves.",
			URLButtonLabel: "Buy now",
			URLButtonURL:   "https://shop.luckyshrub.com/latest/blue-echeveria",
		},
		{
			HeaderType:     "image",
			HeaderLink:     "https://www.luckyshrub.com/assets/zebra-haworthia.jpeg",
			BodyText:       "*Zebra Haworthia*\n\nStriking white stripes on deep green leaves.",
			URLButtonLabel: "Buy now",
			URLButtonURL:   "https://shop.luckyshrub.com/latest/zebra-haworthia",
		},
	})

	got, _ := json.Marshal(msg)

	expected := &interactive.Message{}
	if err := json.Unmarshal([]byte(`{
		"type": "carousel",
		"body": { "text": "Here are our latest arrivals:" },
		"action": {
			"cards": [
				{
					"card_index": 0,
					"type": "cta_url",
					"header": { "type": "image", "image": { "link": "https://www.luckyshrub.com/assets/blue-echeveria.jpeg" } },
					"body": { "text": "*Blue Echeveria*\n\nA rosette-shaped succulent with powdery blue leaves." },
					"action": { "name": "cta_url", "parameters": { "display_text": "Buy now", "url": "https://shop.luckyshrub.com/latest/blue-echeveria" } }
				},
				{
					"card_index": 1,
					"type": "cta_url",
					"header": { "type": "image", "image": { "link": "https://www.luckyshrub.com/assets/zebra-haworthia.jpeg" } },
					"body": { "text": "*Zebra Haworthia*\n\nStriking white stripes on deep green leaves." },
					"action": { "name": "cta_url", "parameters": { "display_text": "Buy now", "url": "https://shop.luckyshrub.com/latest/zebra-haworthia" } }
				}
			]
		}
	}`), expected); err != nil {
		t.Fatal(err)
	}

	want, _ := json.Marshal(expected)

	if !gcmp.Equal(got, want) {
		t.Errorf("Media Carousel (URL buttons) JSON mismatch:\n got:  %s\n want: %s", got, want)
	}
}

func TestMediaCarousel_QuickReplyButtons(t *testing.T) {
	t.Parallel()

	msg := interactive.MediaCarousel("Check out our new arrivals:", []*interactive.MediaCarouselCard{
		{
			HeaderType: "image",
			HeaderLink: "https://www.luckyshrub.com/assets/blue-echeveria.jpeg",
			BodyText:   "*Blue Echeveria*\n\nA rosette-shaped succulent.",
			QuickReplyButtons: []interactive.MediaCarouselButton{
				{ID: "learn-blue-echeveria", Title: "Learn more"},
				{ID: "fav-blue-echeveria", Title: "Add to favorites"},
			},
		},
		{
			HeaderType: "video",
			HeaderLink: "https://www.luckyshrub.com/assets/panda-plant-preview.mp4",
			BodyText:   "*Panda Plant*\n\nSoft, fuzzy leaves.",
			QuickReplyButtons: []interactive.MediaCarouselButton{
				{ID: "learn-panda-plant", Title: "Learn more"},
				{ID: "fav-panda-plant", Title: "Add to favorites"},
			},
		},
	})

	got, _ := json.Marshal(msg)

	expected := &interactive.Message{}
	if err := json.Unmarshal([]byte(`{
		"type": "carousel",
		"body": { "text": "Check out our new arrivals:" },
		"action": {
			"cards": [
				{
					"card_index": 0,
					"type": "cta_url",
					"header": { "type": "image", "image": { "link": "https://www.luckyshrub.com/assets/blue-echeveria.jpeg" } },
					"body": { "text": "*Blue Echeveria*\n\nA rosette-shaped succulent." },
					"action": {
						"buttons": [
							{ "type": "quick_reply", "quick_reply": { "id": "learn-blue-echeveria", "title": "Learn more" } },
							{ "type": "quick_reply", "quick_reply": { "id": "fav-blue-echeveria", "title": "Add to favorites" } }
						]
					}
				},
				{
					"card_index": 1,
					"type": "cta_url",
					"header": { "type": "video", "video": { "link": "https://www.luckyshrub.com/assets/panda-plant-preview.mp4" } },
					"body": { "text": "*Panda Plant*\n\nSoft, fuzzy leaves." },
					"action": {
						"buttons": [
							{ "type": "quick_reply", "quick_reply": { "id": "learn-panda-plant", "title": "Learn more" } },
							{ "type": "quick_reply", "quick_reply": { "id": "fav-panda-plant", "title": "Add to favorites" } }
						]
					}
				}
			]
		}
	}`), expected); err != nil {
		t.Fatal(err)
	}

	want, _ := json.Marshal(expected)

	if !gcmp.Equal(got, want) {
		t.Errorf("Media Carousel (quick-reply buttons) JSON mismatch:\n got:  %s\n want: %s", got, want)
	}
}
