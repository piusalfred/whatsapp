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

package media

type (
	// Info contains metadata about a media attachment received via webhooks.
	// URL is a direct download link (use with access token). ID can be used
	// with the Media API to retrieve the asset.
	Info struct {
		ID       string `json:"id,omitempty"`
		Caption  string `json:"caption,omitempty"`
		MimeType string `json:"mime_type,omitempty"`
		Sha256   string `json:"sha256,omitempty"`
		Filename string `json:"filename,omitempty"`
		Animated bool   `json:"animated,omitempty"` // used with stickers true if animated
		URL      string `json:"url,omitempty"`      // direct download URL (v2025.11+)
	}

	Media struct {
		ID       string `json:"id,omitempty"`
		Link     string `json:"link,omitempty"`
		Caption  string `json:"caption,omitempty"`
		Filename string `json:"filename,omitempty"`
		Provider string `json:"provider,omitempty"`
	}

	Document struct {
		ID       string `json:"id,omitempty"`
		Link     string `json:"link,omitempty"`
		Caption  string `json:"caption,omitempty"`
		Filename string `json:"filename,omitempty"`
	}

	Video struct {
		ID      string `json:"id,omitempty"`
		Link    string `json:"link,omitempty"`
		Caption string `json:"caption,omitempty"`
	}

	Image struct {
		ID       string `json:"id,omitempty"`
		Link     string `json:"link,omitempty"`
		Caption  string `json:"caption,omitempty"`
		Filename string `json:"filename,omitempty"`
	}

	Sticker struct {
		ID string `json:"id,omitempty"`
	}

	Audio struct {
		ID    string `json:"id,omitempty"`
		Link  string `json:"link,omitempty"`
		Voice bool   `json:"voice,omitempty"`
	}

	Location struct {
		Longitude float64 `json:"longitude"`
		Latitude  float64 `json:"latitude"`
		Name      string  `json:"name"`
		Address   string  `json:"address"`
		URL       string  `json:"url,omitempty"`
	}

	Reaction struct {
		MessageID string `json:"message_id"`
		Emoji     string `json:"emoji"`
	}
)
