package cli

import (
	"io"
	"net/http"
	"time"
)

const (
	JsonOutPutFormat  OutPutFormat = "json"
	TextOutPutFormat  OutPutFormat = "text"
	BytesOutPutFormat OutPutFormat = "bytes"
)

type (
	OutPutFormat string
	// HttpClient is a struct that wraps the *http.Client and hence it can be used to
	// make http requests by the cli, with more fine grained control like logging,
	// middlewares, retries, etc
	HttpClient struct {
		http       *http.Client
		Debug      bool
		Logger     io.Writer
		MaxRetries int
		Jitter     time.Duration
		Timeout    time.Duration
		Out        OutPutFormat
	}
)
