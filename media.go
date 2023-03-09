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
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"path/filepath"

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
		FileSize         int64  `json:"file_size"`
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

// GetMedia retrieve the media object by using its corresponding media ID.
func (client *Client) GetMedia(ctx context.Context, mediaID string) (*Media, error) {
	reqCtx := &whttp.RequestContext{
		Name:       "get media",
		BaseURL:    client.baseURL,
		ApiVersion: client.apiVersion,
		Endpoints:  []string{mediaID},
	}

	params := &whttp.Request{
		Context: reqCtx,
		Method:  http.MethodGet,
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
func (client *Client) DeleteMedia(ctx context.Context, mediaID string) (*DeleteMediaResponse, error) {
	reqCtx := &whttp.RequestContext{
		Name:       "delete media",
		BaseURL:    client.baseURL,
		ApiVersion: client.apiVersion,
		Endpoints:  []string{mediaID},
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

func (client *Client) UploadMedia(ctx context.Context, mediaType MediaType, filename string, fr io.Reader) (*UploadMediaResponse, error) {
	payload, contentType, err := uploadMediaPayload(mediaType, filename, fr)
	if err != nil {
		return nil, err
	}

	reqCtx := &whttp.RequestContext{
		Name:       "upload media",
		BaseURL:    client.baseURL,
		ApiVersion: client.apiVersion,
		Endpoints:  []string{client.phoneNumberID, "media"},
	}

	params := &whttp.Request{
		Context: reqCtx,
		Method:  http.MethodPost,
		Headers: map[string]string{"Content-Type": contentType},
		Bearer:  client.accessToken,
		Payload: payload,
	}

	resp := new(UploadMediaResponse)
	err = whttp.Send(ctx, client.http, params, &resp)
	if err != nil {
		return nil, fmt.Errorf("upload media: %w", err)
	}

	return resp, nil
}

// DownloadMedia downloads a media file from the given media ID.
// It accepts a media url and returns a reader and an error.
func (client *Client) DownloadMedia(ctx context.Context, mediaID string) (io.Reader, error) {
	media, err := client.GetMedia(ctx, mediaID)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, media.URL, nil)
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
		return nil, fmt.Errorf("failed to download media: status %d", resp.StatusCode)
	}

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

// uploadMediaPayload creates upload media request payload.
// If nor error, payload content and request content type is returned.
func uploadMediaPayload(mediaType MediaType, filename string, fr io.Reader) ([]byte, string, error) {
	var payload bytes.Buffer
	writer := multipart.NewWriter(&payload)

	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", fmt.Sprintf(`form-data; name=file; filename="%s"`, filename))

	contentType := mime.TypeByExtension(filepath.Ext(filename))
	header.Set("Content-Type", contentType)

	part, err := writer.CreatePart(header)
	if err != nil {
		return nil, "", err
	}

	_, err = io.Copy(part, fr)
	if err != nil {
		return nil, "", err
	}

	err = writer.WriteField("type", string(mediaType))
	if err != nil {
		return nil, "", err
	}

	err = writer.WriteField("messaging_product", "whatsapp")
	if err != nil {
		return nil, "", err
	}

	writer.Close()

	return payload.Bytes(), writer.FormDataContentType(), nil
}
