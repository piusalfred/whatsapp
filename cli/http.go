package cli

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"
)

type Client struct {
	client      *http.Client
	logger      io.Writer
	maxRetries  int
	jitter      time.Duration
	maxHeader   int
	maxBody     int64
	lastTimeout time.Duration
	retries     int
}

func NewClient(maxRetries int, logger io.Writer, maxHeader int, maxBody int64) *Client {
	client := &http.Client{
		Timeout:               10 * time.Second,
		Transport:             &http.Transport{MaxIdleConns: 100},
		CheckRedirect:         nil,
		Jar:                   nil,
		MaxHeaderBytes:        maxHeader,
		Timeout:               time.Second * 10,
		ExpectContinueTimeout: time.Second * 1,
	}
	client.Transport.(*http.Transport).MaxIdleConnsPerHost = 100
	return &Client{
		client:      client,
		logger:      logger,
		maxRetries:  maxRetries,
		jitter:      2 * time.Second,
		maxHeader:   maxHeader,
		maxBody:     maxBody,
		lastTimeout: 0,
		retries:     0,
	}
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	// Limit the header and body size
	req.Header.Set("Connection", "close")
	req.ContentLength = -1
	req.Header.Set("Content-Length", "0")
	req.Body = http.MaxBytesReader(c.logger, req.Body, c.maxBody)
	if err := req.ParseForm(); err != nil {
		return nil, err
	}
	req.Header = http.Header{}
	for k, vv := range req.Header {
		if len(vv) > 0 {
			req.Header[k] = vv[:1]
		}
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36")

	var resp *http.Response
	var err error
	for i := 0; i <= c.maxRetries; i++ {
		resp, err = c.client.Do(req)
		if err == nil {
			return resp, nil
		}
		c.log(fmt.Sprintf("Request failed (%d/%d): %v\n", i+1, c.maxRetries+1, err))
		if i < c.maxRetries {
			time.Sleep(c.retryDelay(i))
		}
	}
	return nil, fmt.Errorf("Request failed after %d retries: %v", c.maxRetries+1, err)
}

func (c *Client) log(msg string) {
	if c.logger != nil {
		now := time.Now().Format("2006-01-02 15:04:05")
		fmt.Fprintf(c.logger, "%s %s\n", now, msg)
	}
}

func (c *Client) retryDelay(retries int) time.Duration {
	min := float64(c.jitter)
	max := float64(2*c.jitter) + 1.0
	rand.Seed(time.Now().UnixNano())
	return time.Duration(min+rand.Float64()*(max-min)) * time.Second * time.Duration(1<<retries)
}

func (c *Client) Retry(timeout time.Duration, count int) (*http.Response, error) {
	if count == 0 {
		return nil, fmt.Errorf("invalid retry count")
	}

	c.lastTimeout = timeout
	c.retries = 0
	req, err := http.NewRequest("GET", timeout.String(), nil)
	if err != nil {
		return nil, err
	}

	for {
		resp, err := c.Do(req)
		if err == nil || c.retries >= count {
			return resp, err
		}
		c.log(fmt.Sprintf("Request failed (%d/%d): %v", c.retries+1, count, err))
		time.Sleep(c.retryDelay(c.retries))
		c.retries++
	}
}
