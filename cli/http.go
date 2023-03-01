// /*
// * Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
// *
// * Permission is hereby granted, free of charge, to any person obtaining a copy of this software
// * and associated documentation files (the “Software”), to deal in the Software without restriction,
// * including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
// * and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
// * subject to the following conditions:
// *
// * The above copyright notice and this permission notice shall be included in all copies or substantial
// * portions of the Software.
// *
// * THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
// * LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
// * IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
// * WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
// * SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
// */
package cli

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

//
//import (
//	"fmt"
//	"io"
//	"net"
//	"net/http"
//	"net/http/httputil"
//	"os"
//	"time"
//)
//
//type (
//	OutputFormat string
//	// HttpClient is a struct that wraps the *http.Client and hence it can be used to
//	// make http requests by the cli, with more fine grained control like logging,
//	// middlewares, retries, etc
//	HttpClient struct {
//		http   *http.Client
//		Debug  bool
//		Logger io.Writer
//		Out    OutputFormat
//	}
//
//	ClientOption func(*HttpClient)
//)
//
//func WithDebug(debug bool) ClientOption {
//	return func(client *HttpClient) {
//		client.Debug = debug
//	}
//}
//
//func WithXLogger(logger io.Writer) ClientOption {
//	return func(client *HttpClient) {
//		client.Logger = logger
//	}
//}
//
//func WithOutPutFormat(out OutputFormat) ClientOption {
//	return func(client *HttpClient) {
//		client.Out = out
//	}
//}
//
//func WithRequestTimeout(timeout time.Duration) ClientOption {
//	return func(client *HttpClient) {
//		client.http.Timeout = timeout
//	}
//}
//
//func NewHttpClient(options ...ClientOption) *HttpClient {
//	client := &HttpClient{
//		http: &http.Client{
//			Transport: &http.Transport{
//				Proxy: http.ProxyFromEnvironment,
//				DialContext: (&net.Dialer{
//					Timeout:   30 * time.Second,
//					KeepAlive: 30 * time.Second,
//					DualStack: true,
//				}).DialContext,
//				MaxIdleConns:          1,
//				IdleConnTimeout:       90 * time.Second,
//				TLSHandshakeTimeout:   10 * time.Second,
//				ExpectContinueTimeout: 1 * time.Second,
//				ForceAttemptHTTP2:     true,
//				MaxIdleConnsPerHost:   -1,
//				DisableKeepAlives:     true,
//			},
//
//			Timeout: 30 * time.Second,
//		},
//		Out:    JsonOutputFormat,
//		Debug:  false,
//		Logger: os.Stdout,
//	}
//
//	for _, option := range options {
//		option(client)
//	}
//
//	client.http.CheckRedirect = client.redirectChecker
//
//	return client
//}
//
//// HTTP returns a http.Client
//func (c *HttpClient) HTTP() *http.Client {
//	return c.http
//}
//
//func (c *HttpClient) redirectChecker(req *http.Request, via []*http.Request) error {
//	if c.Debug {
//		if len(via) == 0 {
//			return nil
//		}
//
//		// Print the path of the redirected URL
//		redirectedUrl := req.URL.String()
//		previousUrl := via[len(via)-1].URL.String()
//		_, _ = fmt.Fprintf(c.Logger, "redirected from %s to %s\n", previousUrl, redirectedUrl)
//
//		return nil
//	}
//
//	return nil
//}
//
//// LogRequest dumps the request to the logger internally uses the DumpRequestOut
//// function.
//func (c *HttpClient) LogRequest(req *http.Request) error {
//	if c.Debug {
//		dump, err := httputil.DumpRequestOut(req, true)
//		if err != nil {
//			return err
//		}
//
//		_, _ = fmt.Fprintf(c.Logger, "\nrequest dump:\n%s\n", string(dump))
//	}
//
//	return nil
//}
//
//// LogResponse dumps the response to the logger internally uses the DumpResponse
//// function.
//func (c *HttpClient) LogResponse(resp *http.Response) error {
//	if c.Debug {
//		dump, err := httputil.DumpResponse(resp, true)
//		if err != nil {
//			return err
//		}
//
//		_, _ = fmt.Fprintf(c.Logger, "\nresponse dump:\n%s\n", string(dump))
//	}
//
//	return nil
//}
//
//type Tripper struct {
//	c *HttpClient
//}
//
//func (t *Tripper) RoundTrip(req *http.Request) (*http.Response, error) {
//	err := t.c.LogRequest(req)
//	if err != nil {
//		return nil, err
//	}
//
//	resp, err := t.c.http.Do(req)
//	if err != nil {
//		return nil, err
//	}
//
//	err = t.c.LogResponse(resp)
//	if err != nil {
//		return nil, err
//	}
//
//	return resp, nil
//}

// RequestToCurlCommand returns a curl command for the given request.
func RequestToCurlCommand(req *http.Request) string {
	var curlCmd string

	curlCmd += "curl -X " + req.Method + " "

	for key, value := range req.Header {
		curlCmd += fmt.Sprintf("-H '%s: %s' ", key, value[0])
	}

	if req.Body != nil {
		bodyBytes, _ := io.ReadAll(req.Body)
		curlCmd += fmt.Sprintf("-d '%s' ", string(bodyBytes))
	}

	curlCmd += req.URL.String()

	return curlCmd
}

// CurlLogger is a middleware that logs the curl command for the request. It implements
// the RoundTripper interface.
type CurlLogger struct {
	Next http.RoundTripper
}

// RoundTrip implements the RoundTripper interface.
func (c *CurlLogger) RoundTrip(req *http.Request) (*http.Response, error) {
	fmt.Println(RequestToCurlCommand(req))

	return c.Next.RoundTrip(req)
}

// NewHttpClient returns a new http.Client with a CurlLogger middleware.
func NewHttpClient() *http.Client {
	defaultTransport := http.DefaultTransport.(*http.Transport)
	return &http.Client{
		Transport: &CurlLogger{
			Next: defaultTransport,
		},
		Jar:     nil,
		Timeout: 60 * time.Second,
	}
}
