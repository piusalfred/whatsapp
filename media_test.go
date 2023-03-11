/*
 * Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of this software
 * and associated documentation files (the “Software”), to deal in the Software without restriction,
 * including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
 * and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
 * subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all copies or substantial
 * portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
 * LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
 * IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
 * WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
 * SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package whatsapp

import (
	"bytes"
	"testing"
)

func TestBuildPayloadForAudioMessage(t *testing.T) { //nolint:paralleltest
	request := &SendMediaRequest{
		Recipient: "2348123456789",
		Type:      "audio",
		MediaID:   "1234567890",
		MediaLink: "https://example.com/audio.mp3",
		Caption:   "Audio caption",
		Filename:  "audio.mp3",
		Provider:  "whatsapp",

		CacheOptions: nil,
	}

	payload, err := formatMediaPayload(request)
	if err != nil {
		t.Errorf("formatMediaPayload() error = %v", err)
	}

	expected := `{"messaging_product":"whatsapp","recipient_type":"individual","to":"2348123456789","type": "audio","audio":{"id":"1234567890","link":"https://example.com/audio.mp3","caption":"Audio caption","filename":"audio.mp3","provider":"whatsapp"}}` //nolint:lll

	if !bytes.Equal(payload, []byte(expected)) {
		t.Errorf("formatMediaPayload() got = %v, want %v", payload, expected)
	}

	t.Logf("audio payload: %s", payload)
}

func BenchmarkBuildPayloadForMediaMessage(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := formatMediaPayload(&SendMediaRequest{
			Recipient: "2348123456789",
			Type:      "audio",
			MediaID:   "1234567890",
			MediaLink: "https://example.com/audio.mp3",
			Caption:   "Audio caption",
			Filename:  "audio.mp3",
			Provider:  "whatsapp",

			CacheOptions: nil,
		})
		if err != nil {
			b.Errorf("formatMediaPayload() error = %v", err)

			return
		}
	}
}
