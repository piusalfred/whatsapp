package message

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/piusalfred/whatsapp/message"
)

type (
	Server struct {
		whatsapp           message.Service
		http               *http.Server
		mcp                *mcp.Server
		config             *Config
		handler            http.Handler
		logger             *slog.Logger
		httpServerInitHook func(server *http.Server) error
		mcpServerInitHook  func(server *mcp.Server) error
		sessionIDGetter    func() string
	}

	SendTextRequest struct {
		Text       string `json:"text"`
		PreviewURL bool   `json:"preview_url"`
		Recipient  string `json:"recipient"`
		ReplyTo    string `json:"reply_to"`
	}

	SendTextResponse struct {
		Product       string `json:"product"`
		Input         string `json:"input"`
		WhatsappID    string `json:"whatsapp_id"`
		MessageID     string `json:"message_id"`
		MessageStatus string `json:"message_status"`
	}

	Service interface {
		SendText(ctx context.Context, input *SendTextRequest) (*SendTextResponse, error)
	}

	Config struct {
		HTTPAddress                      string
		HTTPDisableGeneralOptionsHandler bool
		HTTPReadTimeout                  time.Duration
		HTTPReadHeaderTimeout            time.Duration
		HTTPWriteTimeout                 time.Duration
		HTTPIdleTimeout                  time.Duration
		HTTPMaxHeaderBytes               int
		LogLevel                         slog.Level
		LogHandler                       slog.Handler
		MCPServerName                    string
		MCPServerTitle                   string
		MCPServerVersion                 string
		MCPOptions                       *mcp.ServerOptions
	}

	ServerOption interface {
		apply(server *Server) error
	}

	serverOption func(server *Server) error
)

func (fn serverOption) apply(server *Server) error {
	return fn(server)
}

func WithServerHTTPMiddlewares(middlewares ...func(http.Handler) http.Handler) ServerOption {
	return serverOption(func(server *Server) error {
		finalHandler := server.handler
		for _, m := range middlewares {
			finalHandler = m(finalHandler)
		}

		server.handler = finalHandler

		return nil
	})
}

func WithServerMCPServerInitHook(hook func(server *mcp.Server) error) ServerOption {
	return serverOption(func(server *Server) error {
		server.mcpServerInitHook = hook
		return nil
	})
}

func WithServerHTTPServerInitHook(hook func(server *http.Server) error) ServerOption {
	return serverOption(func(server *Server) error {
		server.httpServerInitHook = hook
		return nil
	})
}

func NewServer(config *Config, client message.Service, options ...ServerOption) (*Server, error) {
	s := &Server{
		config:   config,
		whatsapp: client,
		logger:   slog.New(config.LogHandler.WithGroup("whatsapp.mcp")),
	}
	s.mcp = mcp.NewServer(&mcp.Implementation{
		Name:    s.config.MCPServerName,
		Title:   s.config.MCPServerTitle,
		Version: s.config.MCPServerVersion,
	}, s.config.MCPOptions)

	s.handler = mcp.NewStreamableHTTPHandler(func(req *http.Request) *mcp.Server {
		return s.mcp
	}, &mcp.StreamableHTTPOptions{
		GetSessionID: s.sessionIDGetter,
		Stateless:    true,
		JSONResponse: true,
	})

	for _, option := range options {
		if err := option.apply(s); err != nil {
			return nil, fmt.Errorf("init server apply option: %w", err)
		}
	}

	s.http = &http.Server{
		Addr:                         s.config.HTTPAddress,
		Handler:                      s.handler,
		DisableGeneralOptionsHandler: s.config.HTTPDisableGeneralOptionsHandler,
		ReadTimeout:                  s.config.HTTPReadTimeout,
		ReadHeaderTimeout:            s.config.HTTPReadHeaderTimeout,
		WriteTimeout:                 s.config.HTTPWriteTimeout,
		IdleTimeout:                  s.config.HTTPIdleTimeout,
		MaxHeaderBytes:               s.config.HTTPMaxHeaderBytes,
		ErrorLog:                     slog.NewLogLogger(s.config.LogHandler, s.config.LogLevel),
	}

	if httpHookErr := s.onHttpServerInitHook(s.httpServerInitHook); httpHookErr != nil {
		return nil, fmt.Errorf("init server: %w", httpHookErr)
	}

	if mcpHookErr := s.onMCPServerInitHook(s.mcpServerInitHook); mcpHookErr != nil {
		return nil, fmt.Errorf("init mcp server: %w", mcpHookErr)
	}

	s.addSendTextTool()

	return s, nil
}

func (server *Server) SendText(ctx context.Context, request *SendTextRequest) (*SendTextResponse, error) {
	text := message.NewRequest(request.Recipient, &message.Text{
		Body:       request.Text,
		PreviewURL: request.PreviewURL,
	}, request.ReplyTo)

	output, err := server.whatsapp.SendText(ctx, text)
	if err != nil {
		return nil, err
	}

	return &SendTextResponse{
		Product:       output.Product,
		Input:         output.Contacts[0].Input,
		WhatsappID:    output.Contacts[0].WhatsappID,
		MessageStatus: output.Messages[0].MessageStatus,
		MessageID:     output.Messages[0].ID,
	}, nil
}

func (server *Server) onHttpServerInitHook(hook func(s *http.Server) error) error {
	if hook != nil {
		if hookErr := hook(server.http); hookErr != nil {
			return fmt.Errorf("on http server init hook: %w", hookErr)
		}
	}

	return nil
}

func (server *Server) onMCPServerInitHook(hook func(s *mcp.Server) error) error {
	if hook != nil {
		if hookErr := hook(server.mcp); hookErr != nil {
			return fmt.Errorf("on mcp server init hook: %w", hookErr)
		}
	}

	return nil
}

func (server *Server) addSendTextTool() {
	inputSchema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"text": {
				Type:        "string",
				Description: "Text body to send.",
			},
			"recipient": {
				Type:        "string",
				Description: "Recipient phone number in international format.",
			},
			"preview_url": {
				Type:        "boolean",
				Description: "Whether to allow URL preview if a link is present in the text.",
			},
			"reply_to": {
				Type:        "string",
				Description: "Optional message ID this text is replying to.",
			},
		},
		Required:    []string{"text", "recipient"},
		Description: "Input for sending a WhatsApp text message.",
	}

	outputSchema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"product": {
				Type:        "string",
				Description: "API product type returned by WhatsApp (e.g., whatsapp).",
			},
			"input": {
				Type:        "string",
				Description: "Normalized recipient input returned by WhatsApp.",
			},
			"whatsapp_id": {
				Type:        "string",
				Description: "WhatsApp ID associated with the recipient.",
			},
			"message_status": {
				Type:        "string",
				Description: "Delivery status of the message (e.g., accepted).",
			},
			"message_id": {
				Type:        "string",
				Description: "Unique ID of the created message.",
			},
		},
		Required:    []string{"product", "input", "whatsapp_id", "message_status", "message_id"},
		Description: "Response returned after sending a WhatsApp text message.",
	}

	toolHandlerFunc := mcp.ToolHandlerFor[*SendTextRequest, *SendTextResponse](
		func(ctx context.Context, request *mcp.CallToolRequest, input *SendTextRequest) (
			result *mcp.CallToolResult, output *SendTextResponse, err error,
		) {
			output, err = server.SendText(ctx, input)
			if err != nil {
				return nil, nil, err
			}

			result = &mcp.CallToolResult{
				IsError: false,
			}

			return result, output, nil
		})

	tool := &mcp.Tool{
		Meta:         nil,
		Annotations:  nil,
		Description:  "send text to whatsapp number",
		InputSchema:  inputSchema,
		Name:         "whatsapp-send-text",
		OutputSchema: outputSchema,
		Title:        "send text to whatsapp number",
	}

	mcp.AddTool(server.mcp, tool, toolHandlerFunc)
}
