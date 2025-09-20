package message

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/piusalfred/whatsapp/message"
)

type (
	Server struct {
		whatsapp           WhatsappService
		http               *http.Server
		mcp                *mcp.Server
		config             *Config
		handler            http.Handler
		logger             *slog.Logger
		httpServerInitHook func(server *http.Server) error
		mcpServerInitHook  func(server *mcp.Server) error
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

	WhatsappService interface {
		SendText(ctx context.Context, request *message.Request[message.Text]) (*message.Response, error)
	}

	Config struct {
		HTTPAddress                      string
		HTTPDisableGeneralOptionsHandler bool
		HTTPReadTimeout                  time.Duration
		HTTPReadHeaderTimeout            time.Duration
		HTTPWriteTimeout                 time.Duration
		HTTPIdleTimeout                  time.Duration
		HTTPMaxHeaderBytes               int
		HTTPServerShutdownTimeout        time.Duration
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

func NewServer(config *Config, client WhatsappService, options ...ServerOption) (*Server, error) {
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

	s.handler = mcp.NewStreamableHTTPHandler(func(_ *http.Request) *mcp.Server {
		return s.mcp
	}, &mcp.StreamableHTTPOptions{
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

	if httpHookErr := s.onHTTPServerInitHook(s.httpServerInitHook); httpHookErr != nil {
		return nil, fmt.Errorf("init server: %w", httpHookErr)
	}

	if mcpHookErr := s.onMCPServerInitHook(s.mcpServerInitHook); mcpHookErr != nil {
		return nil, fmt.Errorf("init mcp server: %w", mcpHookErr)
	}

	sendTextPrompt := &mcp.Prompt{
		Description: "send whatsapp text message to whatsapp prompt given a number and a message," +
			"with the option to specify if URL preview is allowed",
		Name:  "whatsapp-send-text",
		Title: "WhatsApp Send Text Message Prompt",
	}

	s.mcp.AddPrompt(sendTextPrompt, func(_ context.Context,
		_ *mcp.GetPromptRequest,
	) (*mcp.GetPromptResult, error) {
		result := &mcp.GetPromptResult{
			Description: sendTextPrompt.Description,
			Messages: []*mcp.PromptMessage{
				{
					Role: mcp.Role("user"),
					Content: &mcp.TextContent{
						Text:        "",
						Meta:        nil,
						Annotations: nil,
					},
				},
			},
		}

		return result, nil
	})

	s.addSendTextTool()

	return s, nil
}

func (server *Server) Run(ctx context.Context) error {
	sigCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		server.logger.LogAttrs(ctx, slog.LevelInfo, "starting http server",
			slog.String("address", server.config.HTTPAddress))
		if err := server.http.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("listen and serve: %w", err)
			return
		}
		errCh <- nil
	}()

	select {
	case <-sigCtx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), server.config.HTTPServerShutdownTimeout)
		defer cancel()

		if err := server.http.Shutdown(shutdownCtx); err != nil {
			server.logger.LogAttrs(shutdownCtx, slog.LevelError, "shutdown http server",
				slog.String("error", err.Error()))

			return fmt.Errorf("shutdown server failed: %w", err)
		}
		return <-errCh

	case err := <-errCh:
		server.logger.LogAttrs(ctx, slog.LevelError,
			"http server error", slog.String("error", err.Error()))
		return err
	}
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

func (server *Server) onHTTPServerInitHook(hook func(s *http.Server) error) error {
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

	toolHandlerFunc := mcp.ToolHandlerFor[*SendTextRequest, *SendTextResponse](server.HandleSendText)

	tool := &mcp.Tool{
		Meta:         nil,
		Annotations:  nil,
		Description:  "send text message to whatsapp number, with the option to specify previewing URL if message content contains one",
		InputSchema:  inputSchema,
		Name:         "whatsapp-send-text",
		OutputSchema: outputSchema,
		Title:        "send text message to whatsapp number",
	}

	mcp.AddTool(server.mcp, tool, toolHandlerFunc)
}

func (server *Server) HandleSendText(ctx context.Context, request *mcp.CallToolRequest, input *SendTextRequest) (
	*mcp.CallToolResult, *SendTextResponse, error,
) {
	sessionID := request.GetSession().ID()
	server.logger.LogAttrs(ctx, slog.LevelInfo, "handle send text request", slog.String("sessionID", sessionID))

	result := &mcp.CallToolResult{}
	output, err := server.SendText(ctx, input)
	if err != nil {
		result.IsError = true
		return result, nil, err
	}

	result.Content = []mcp.Content{
		&mcp.TextContent{Text: fmt.Sprintf(
			"The whatsapp message to number %s has been sent with the response being %s",
			input.Recipient,
			output.MessageID,
		)},
	}
	result.IsError = false

	return result, output, nil
}
