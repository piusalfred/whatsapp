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
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"path/filepath"

	whttp "github.com/piusalfred/whatsapp/http"
)

type (
	MediaInformation struct {
		MessagingProduct string `json:"messaging_product"`
		URL              string `json:"url"`
		MimeType         string `json:"mime_type"`
		Sha256           string `json:"sha256"`
		FileSize         int64  `json:"file_size"`
		ID               string `json:"id"`
	}

	MediaType string

	// UploadMediaRequest contains the information needed to upload a media file.
	// File Path to the file stored in your local directory. For example: "@/local/path/file.jpg".
	// Type - type of media file being uploaded. See Supported MediaInformation Types for more information.
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

// GetMediaInformation retrieve the media object by using its corresponding media ID.
func (client *Client) GetMediaInformation(ctx context.Context, mediaID string) (*MediaInformation, error) {
	reqCtx := &whttp.RequestContext{
		Name:       "get media",
		BaseURL:    client.Config.BaseURL,
		ApiVersion: client.Config.Version,
		Endpoints:  []string{mediaID},
	}

	params := &whttp.Request{
		Context: reqCtx,
		Method:  http.MethodGet,
		Bearer:  client.Config.AccessToken,
		Payload: nil,
	}

	var media MediaInformation

	err := client.Base.Do(ctx, params, &media)
	if err != nil {
		return nil, fmt.Errorf("get media: %w", err)
	}

	return &media, nil
}

// DeleteMedia delete the media by using its corresponding media ID.
func (client *Client) DeleteMedia(ctx context.Context, mediaID string) (*DeleteMediaResponse, error) {
	reqCtx := &whttp.RequestContext{
		Name:       "delete media",
		BaseURL:    client.Config.BaseURL,
		ApiVersion: client.Config.Version,
		Endpoints:  []string{mediaID},
	}

	params := &whttp.Request{
		Context: reqCtx,
		Method:  http.MethodDelete,
		Headers: map[string]string{"Content-Type": "application/json"},
		Bearer:  client.Config.AccessToken,
		Payload: nil,
	}

	resp := new(DeleteMediaResponse)
	err := client.Base.Do(ctx, params, &resp)
	if err != nil {
		return nil, fmt.Errorf("delete media: %w", err)
	}

	return resp, nil
}

func (client *Client) UploadMedia(ctx context.Context, mediaType MediaType, filename string,
	fr io.Reader,
) (*UploadMediaResponse, error) {
	payload, contentType, err := uploadMediaPayload(mediaType, filename, fr)
	if err != nil {
		return nil, err
	}

	reqCtx := &whttp.RequestContext{
		Name:       "upload media",
		BaseURL:    client.Config.BaseURL,
		ApiVersion: client.Config.Version,
		Endpoints:  []string{client.Config.PhoneNumberID, "media"},
	}

	params := &whttp.Request{
		Context: reqCtx,
		Method:  http.MethodPost,
		Headers: map[string]string{"Content-Type": contentType},
		Bearer:  client.Config.AccessToken,
		Payload: payload,
	}

	resp := new(UploadMediaResponse)
	err = client.Base.Do(ctx, params, &resp)
	if err != nil {
		return nil, fmt.Errorf("upload media: %w", err)
	}

	return resp, nil
}

var ErrMediaDownload = fmt.Errorf("failed to download media")

type DownloadMediaResponse struct {
	Headers    http.Header
	Body       io.Reader
	StatusCode int
}

type DownloadResponseDecoder struct {
	Resp     *DownloadMediaResponse
	response *http.Response
}

func (d *DownloadResponseDecoder) Decode(response *http.Response) error {
	d.Resp.Headers = response.Header
	d.Resp.Body = response.Body

	return nil
}

// DownloadMedia download the media by using its corresponding media ID. It uses the media ID to retrieve
// the media URL. All media URLs expire after 5 minutes —you need to retrieve the media URL again if it
// expires.
// If successful, *DownloadMediaResponse will be returned. It contains headers and io.Reader. From the headers
// you  can check a content-type header to indicate the mime type of returned data.
//
// If media fails to download, Facebook returns a 404 http status code. It is recommended to try to retrieve
// a new media URL and download it again. This will go on for an n retries. If doing so doesn't resolve the issue,
// please try to renew the access token, then retry downloading the media.
func (client *Client) DownloadMedia(ctx context.Context, mediaID string, retries int) (*DownloadMediaResponse, error) {
	// create a for loop to retry the download if it fails with a 404 http status code.
	for i := 0; i <= retries; i++ {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("media download: %w", ctx.Err())
		default:
		}
		media, err := client.GetMediaInformation(ctx, mediaID)
		if err != nil {
			return nil, err
		}

		request := whttp.MakeRequest(
			whttp.WithRequestContext(&whttp.RequestContext{
				Name:              "download media",
				BaseURL:           media.URL,
				ApiVersion:        client.Config.Version,
				PhoneNumberID:     client.Config.PhoneNumberID,
				Bearer:            client.Config.AccessToken,
				BusinessAccountID: "",
				Endpoints:         nil,
			}),
			whttp.WithRequestName("download media"),
			whttp.WithMethod(http.MethodGet),
			whttp.WithBearer(client.Config.AccessToken))
		decoder := &DownloadResponseDecoder{}
		if err := client.Base.DoWithDecoder(
			ctx,
			request,
			whttp.RawResponseDecoder(decoder.Decode),
			nil); err != nil {
			return nil, fmt.Errorf("media download: %w", err)
		}

		// retry
		resp := decoder.response
		if resp.StatusCode == http.StatusNotFound {
			_ = resp.Body.Close()

			continue
		}

		if resp.StatusCode != http.StatusOK {
			_ = resp.Body.Close()

			return nil, fmt.Errorf("%w: status %d", ErrMediaDownload, resp.StatusCode)
		}

		var buf bytes.Buffer
		_, err = io.CopyN(&buf, resp.Body, MaxDocSize)
		if err != nil && !errors.Is(err, io.EOF) {
			_ = resp.Body.Close()

			return nil, fmt.Errorf("media download: %w", err)
		}

		_ = resp.Body.Close()

		return &DownloadMediaResponse{
			Headers: resp.Header,
			Body:    &buf,
		}, nil
	}

	return nil, fmt.Errorf("%w: retries exceeded", ErrMediaDownload)
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
		return nil, "", fmt.Errorf("media upload: %w", err)
	}

	_, err = io.Copy(part, fr)
	if err != nil {
		return nil, "", fmt.Errorf("media upload: %w", err)
	}

	err = writer.WriteField("type", string(mediaType))
	if err != nil {
		return nil, "", fmt.Errorf("media upload: %w", err)
	}

	err = writer.WriteField("messaging_product", "whatsapp")
	if err != nil {
		return nil, "", fmt.Errorf("media upload: %w", err)
	}

	_ = writer.Close()

	return payload.Bytes(), writer.FormDataContentType(), nil
}
