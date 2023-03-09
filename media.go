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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	whttp "github.com/piusalfred/whatsapp/http"
)

const (
	MediaUpload      MediaOp = "POST"
	MediaURLRetrieve MediaOp = "GET"
	MediaDelete      MediaOp = "DELETE"
	MediaDownload    MediaOp = "GET"
)

type (
	Media struct {
		MessagingProduct string `json:"messaging_product"`
		URL              string `json:"url"`
		MimeType         string `json:"mime_type"`
		Sha256           string `json:"sha256"`
		FileSize         string `json:"file_size"`
		ID               string `json:"id"`
	}

	// MediaOp represents the operations that can be performed on media. There are 4 different
	// operations that can be performed on media:
	// 	- POST /PHONE_NUMBER_ID/media Upload media.
	//	- GET /MEDIA_ID Retrieve the URL for a specific media.
	//	- DELETE /MEDIA_ID Delete a specific media.
	//	- GET /MEDIA_URL  Download media from a media URL.
	MediaOp string

	MediaType            string
	MediaOperationParams struct {
		Token         string
		PhoneNumberID string
		BaseURL       string
		ApiVersion    string
		Endpoint      string
		Method        string
	}

	// UploadMediaRequest contains the information needed to upload a media file.
	// File Path to the file stored in your local directory. For example: "@/local/path/file.jpg".
	// Type - type of media file being uploaded. See Supported Media Types for more information.
	// Product Messaging service used for the request. In this case, use whatsapp.
	// MediaID - ID of the media file. This is the ID that you will use to send the media file.
	UploadMediaRequest struct {
		MediaID  string
		FilePath string
		Type     MediaType
		Product  string
	}

	MediaRequestParams struct {
		Token   string
		MediaID string
	}

	UploadMediaResponse struct {
		ID string `json:"id"`
	}

	DeleteMediaResponse struct {
		Success bool `json:"success"`
	}
)

// UploadMedia uploads a media file to the WhatsApp Server.To upload media, A POST call to /PHONE_NUMBER_ID/media is made.
// All media files sent through this endpoint are encrypted and persist for 30 days, unless they are deleted earlier.
//
// # Media ID
//
// To complete some of the following API calls, you need to have a media ID. There are two ways to get this ID:
//
//   - From the API call: Once you have successfully uploaded media files to the API, the media ID is included
//     in the response to your call.
//
//   - From Webhooks: When a business account receives a media message, it downloads the media and uploads it to
//     the Cloud API automatically. That event triggers the Webhooks and sends you a notification that includes
//     the media ID.
//
// # Media Upload
//
// To upload media, make a POST call to /PHONE_NUMBER_ID/media and include the parameters listed below. All media files
// sent through this endpoint are encrypted and persist for 30 days, unless they are deleted earlier
//
//		curl -X POST 'https://graph.facebook.com/v16.0/<MEDIA_ID>/media' \
//		-H 'Authorization: Bearer <ACCESS_TOKEN>' \
//		-F 'file=@"2jC60Vdjn/cross-trainers-summer-sale.jpg"' \
//		-F 'type="image/jpeg"' \
//		-F 'messaging_product="whatsapp"'
//
//	 A successful response returns an object with the uploaded media's ID:
//
//				{
//	 		   "id":"<MEDIA_ID>"
//				}
//
// # Retrieve Media URL
//
// To retrieve your media’s URL, send a GET request to /MEDIA_ID. Use the returned URL to download the media file.
// Note that clicking this URL (i.e. performing a generic GET) will not return the media; you must include an
// access token.
//
// Parameters:
//
//   - phone_number_id. Optional. Business phone number ID. If included, the operation will only be processed if the
//     ID matches the ID of the business phone number that the media was uploaded on.
//
//     Example:
//     curl -X GET 'https://graph.facebook.com/v16.0/<MEDIA_ID>/' \
//     -H 'Authorization: Bearer <ACCESS_TOKEN>'
//
// A successful response includes an object with a media url. The URL is only valid for 5 minutes. To use this URL, see Download Media.
//
//		{
//	 	"messaging_product": "whatsapp",
//	 	"url": "<URL>",
//	 	"mime_type": "<MIME_TYPE>",
//	 	"sha256": "<HASH>",
//	 	"file_size": "<FILE_SIZE>",
//	 	"id": "<MEDIA_ID>"
//		}
//
// # Media Download
//
// To download media, make a GET call to your media’s URL. All media URLs expire after 5 minutes —you need to retrieve the
// media URL again if it expires. If you directly click on the URL you get from a /MEDIA_ID GET call, you get an access error.
//
// # Delete Media
//
// To delete media, make a DELETE call to the ID of the media you want to delete.
//
// # Example
//
// Sample request:
//
//	curl -X DELETE 'https://graph.facebook.com/v16.0/<MEDIA_ID>' \
//	-H 'Authorization: Bearer <ACCESS_TOKEN>'
//
// Sample response:
//
//		{
//	 	"success": true
//		}
//
// Supported Media Types
//
//   - Audio can have a max size of 16MB
//
//   - audio/aac, audio/mp4, audio/mpeg, audio/amr, audio/ogg (only opus codecs, base audio/ogg is not supported)
//
//   - Documents can have a max size of 100MB
//     Formats: text/plain, application/pdf, application/vnd.ms-powerpoint, application/msword, application/vnd.ms-excel,
//     application/vnd.openxmlformats-officedocument.wordprocessingml.document,
//     application/vnd.openxmlformats-officedocument.presentationml.presentation,
//     application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
//
//   - Images must be 8-bit, RGB or RGBA with a max size of 5MB
//     Formats: image/jpeg, image/png
//
//   - Videos can have a max size of 16MB. Only H.264 video codec and AAC audio codec is supported. We support videos
//     with a single audio stream or no audio stream.
//     Formats: video/mp4, video/3gp
//
//   - Stickers can have a max size of 100KB for static stickers and 500KB for animated stickers.
//
// The maximum supported file size for media messages on Cloud API is 100MB. In the event the customer sends a file that
// is greater than 100MB, you will receive a webhook with error code 131052 and title: "Media file size too big. Max file
// size we currently support: 100MB. Please communicate with your customer to send a media file that is smaller than 100MB"_.
// We advise that you send customers a warning message that their media file exceeds the maximum file size when this webhook
// event is triggered.
//func UploadMedia(ctx context.Context, client *http.Client, params *whttp.Request, options *UploadMediaRequest) (*whttp.Response, error) {
//	return nil, nil
//}

// GetMedia retrieve the media object by using its corresponding media ID.
func (client *Client) GetMedia(ctx context.Context, id string) (*Media, error) {
	reqCtx := &whttp.RequestContext{
		Name:       "get media",
		BaseURL:    client.baseURL,
		ApiVersion: client.apiVersion,
		SenderID:   client.phoneNumberID,
		Endpoints:  []string{id, "media"},
	}

	params := &whttp.Request{
		Context: reqCtx,
		Method:  http.MethodGet,
		Headers: map[string]string{"Content-Type": "application/json"},
		Bearer:  client.accessToken,
		Payload: nil,
	}

	media := new(Media)
	err := whttp.Send(ctx, client.http, params, &media)
	if err != nil {
		return nil, fmt.Errorf("get media: %w", err)
	}

	return media, nil
}

// DeleteMedia delete the media by using its corresponding media ID.
func (client *Client) DeleteMedia(ctx context.Context, id string) (*DeleteMediaResponse, error) {
	reqCtx := &whttp.RequestContext{
		Name:       "delete media",
		BaseURL:    client.baseURL,
		ApiVersion: client.apiVersion,
		SenderID:   client.phoneNumberID,
		Endpoints:  []string{id, "media"},
	}

	params := &whttp.Request{
		Context: reqCtx,
		Method:  http.MethodDelete,
		Headers: map[string]string{"Content-Type": "application/json"},
		Bearer:  client.accessToken,
		Payload: nil,
	}

	resp := new(DeleteMediaResponse)
	err := whttp.Send(ctx, client.http, params, &resp)
	if err != nil {
		return nil, fmt.Errorf("delete media: %w", err)
	}

	return resp, nil
}

// DownloadMedia downloads a media file from the given URL.
// It accepts a media url and returns a byte array and an error.
func (client *Client) DownloadMedia(ctx context.Context, url string) (io.Reader, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", client.accessToken))

	resp, err := client.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResponse whttp.ResponseError
		err = json.NewDecoder(resp.Body).Decode(&errResponse)
		if err != nil {
			return nil, fmt.Errorf("failed to download media: %w", err)
		}

		return nil, fmt.Errorf("failed to download media: %w", errResponse.Err)
	}

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, resp.Body)
	if err != nil && err != io.EOF {
		return nil, err
	}

	return buf, nil
}
