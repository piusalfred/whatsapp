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

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type (
	Server struct {
		whatsapp           Sender
		http               *http.Server
		mcp                *mcp.Server
		config             *Config
		handler            http.Handler
		logger             *slog.Logger
		httpServerInitHook func(server *http.Server) error
		mcpServerInitHook  func(server *mcp.Server) error
		schemas            *Schemas
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

		for i := len(middlewares) - 1; i >= 0; i-- {
			finalHandler = middlewares[i](finalHandler)
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

func NewServer(config *Config, client Sender, options ...ServerOption) (*Server, error) {
	s := &Server{
		config:   config,
		whatsapp: client,
		logger:   slog.New(config.LogHandler.WithGroup("whatsapp.mcp")),
	}
	s.initSchemas()
	s.initMCP()
	s.initTools()
	for _, option := range options {
		if err := option.apply(s); err != nil {
			return nil, fmt.Errorf("init server: failed to apply option: %w", err)
		}
	}
	s.initHTTP()
	if httpHookErr := s.onHTTPServerInitHook(); httpHookErr != nil {
		return nil, fmt.Errorf("init server: %w", httpHookErr)
	}
	if mcpHookErr := s.onMCPServerInitHook(); mcpHookErr != nil {
		return nil, fmt.Errorf("init mcp server: %w", mcpHookErr)
	}
	return s, nil
}

func (s *Server) initMCP() {
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

	prompt := &mcp.Prompt{
		Description: "A prompt for interacting with a user via WhatsApp.",
		Name:        "whatsapp-interaction-prompt",
		Title:       "WhatsApp Interaction Prompt",
	}

	s.mcp.AddPrompt(prompt, func(_ context.Context, _ *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		return &mcp.GetPromptResult{
			Description: prompt.Description,
			Messages: []*mcp.PromptMessage{{
				Role: mcp.Role("user"),
				Content: &mcp.TextContent{
					Text: "",
				},
			}},
		}, nil
	})
}

func (s *Server) initTools() {
	s.addSendTextTool()
	s.addRequestLocationTool()
	s.addSendLocationTool()
	s.addSendImageTool()
	s.addSendDocumentTool()
	s.addSendAudioTool()
	s.addSendVideoTool()
	s.addSendReactionTool()
	s.addSendTemplateTool()
	s.addSendStickerTool()
	s.addSendContactsTool()
	s.addSendInteractiveTool()
}

func (s *Server) initHTTP() {
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
}

func (s *Server) Run(ctx context.Context) error {
	sigCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		s.logger.LogAttrs(ctx, slog.LevelInfo, "starting http server", slog.String("address", s.config.HTTPAddress))
		if err := s.http.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("listen and serve failed: %w", err)
		} else {
			errCh <- nil
		}
	}()

	select {
	case <-sigCtx.Done():
		s.logger.LogAttrs(ctx, slog.LevelInfo, "shutdown signal received")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), s.config.HTTPServerShutdownTimeout)
		defer cancel()
		if err := s.http.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("http server shutdown failed: %w", err)
		}
		return <-errCh
	case err := <-errCh:
		if err != nil {
			s.logger.LogAttrs(ctx, slog.LevelError, "http server error", slog.String("error", err.Error()))
		}
		return err
	}
}

func (s *Server) onHTTPServerInitHook() error {
	if s.httpServerInitHook != nil {
		if err := s.httpServerInitHook(s.http); err != nil {
			return fmt.Errorf("on http server init hook failed: %w", err)
		}
	}
	return nil
}

func (s *Server) onMCPServerInitHook() error {
	if s.mcpServerInitHook != nil {
		if err := s.mcpServerInitHook(s.mcp); err != nil {
			return fmt.Errorf("on mcp server init hook failed: %w", err)
		}
	}
	return nil
}

func (s *Server) addSendTextTool() {
	toolHandler := mcp.ToolHandlerFor[*SendTextRequest, *Response](s.HandleSendText)
	tool := &mcp.Tool{
		Description:  "Sends a text message to a WhatsApp number, with an option to preview URLs.",
		InputSchema:  s.schemas.sendTextRequest,
		Name:         "whatsapp-send-text",
		OutputSchema: s.schemas.response,
		Title:        "SendRequest WhatsApp Text Message",
	}
	mcp.AddTool(s.mcp, tool, toolHandler)
}

func (s *Server) addRequestLocationTool() {
	toolHandler := mcp.ToolHandlerFor[*RequestLocationRequest, *Response](s.HandleRequestLocation)
	tool := &mcp.Tool{
		Description:  "Asks a user to share their location on WhatsApp by sending a prompt.",
		InputSchema:  s.schemas.requestLocationRequest,
		Name:         "whatsapp-request-location",
		OutputSchema: s.schemas.response,
		Title:        "Request User Location via WhatsApp",
	}
	mcp.AddTool(s.mcp, tool, toolHandler)
}

func (s *Server) addSendLocationTool() {
	toolHandler := mcp.ToolHandlerFor[*SendLocationRequest, *Response](s.HandleSendLocation)
	tool := &mcp.Tool{
		Description:  "Sends a location message to a WhatsApp number with coordinates and address.",
		InputSchema:  s.schemas.sendLocationRequest,
		Name:         "whatsapp-send-location",
		OutputSchema: s.schemas.response,
		Title:        "Send WhatsApp Location Message",
	}
	mcp.AddTool(s.mcp, tool, toolHandler)
}

func (s *Server) addSendImageTool() {
	toolHandler := mcp.ToolHandlerFor[*SendImageRequest, *Response](s.HandleSendImage)
	tool := &mcp.Tool{
		Description:  "Sends an image message to a WhatsApp number with optional caption.",
		InputSchema:  s.schemas.sendImageRequest,
		Name:         "whatsapp-send-image",
		OutputSchema: s.schemas.response,
		Title:        "Send WhatsApp Image Message",
	}
	mcp.AddTool(s.mcp, tool, toolHandler)
}

func (s *Server) addSendDocumentTool() {
	toolHandler := mcp.ToolHandlerFor[*SendDocumentRequest, *Response](s.HandleSendDocument)
	tool := &mcp.Tool{
		Description:  "Sends a document message to a WhatsApp number with optional caption.",
		InputSchema:  s.schemas.sendDocumentRequest,
		Name:         "whatsapp-send-document",
		OutputSchema: s.schemas.response,
		Title:        "Send WhatsApp Document Message",
	}
	mcp.AddTool(s.mcp, tool, toolHandler)
}

func (s *Server) addSendAudioTool() {
	toolHandler := mcp.ToolHandlerFor[*SendAudioRequest, *Response](s.HandleSendAudio)
	tool := &mcp.Tool{
		Description:  "Sends an audio message to a WhatsApp number.",
		InputSchema:  s.schemas.sendAudioRequest,
		Name:         "whatsapp-send-audio",
		OutputSchema: s.schemas.response,
		Title:        "Send WhatsApp Audio Message",
	}
	mcp.AddTool(s.mcp, tool, toolHandler)
}

func (s *Server) addSendVideoTool() {
	toolHandler := mcp.ToolHandlerFor[*SendVideoRequest, *Response](s.HandleSendVideo)
	tool := &mcp.Tool{
		Description:  "Sends a video message to a WhatsApp number with optional caption.",
		InputSchema:  s.schemas.sendVideoRequest,
		Name:         "whatsapp-send-video",
		OutputSchema: s.schemas.response,
		Title:        "Send WhatsApp Video Message",
	}
	mcp.AddTool(s.mcp, tool, toolHandler)
}

func (s *Server) addSendReactionTool() {
	toolHandler := mcp.ToolHandlerFor[*SendReactionRequest, *Response](s.HandleSendReaction)
	tool := &mcp.Tool{
		Description:  "Sends a reaction (emoji) to a specific WhatsApp message.",
		InputSchema:  s.schemas.sendReactionRequest,
		Name:         "whatsapp-send-reaction",
		OutputSchema: s.schemas.response,
		Title:        "Send WhatsApp Reaction Message",
	}
	mcp.AddTool(s.mcp, tool, toolHandler)
}

func (s *Server) addSendTemplateTool() {
	toolHandler := mcp.ToolHandlerFor[*SendTemplateRequest, *Response](s.HandleSendTemplate)
	tool := &mcp.Tool{
		Description:  "Sends a template message to a WhatsApp number.",
		InputSchema:  s.schemas.sendTemplateRequest,
		Name:         "whatsapp-send-template",
		OutputSchema: s.schemas.response,
		Title:        "Send WhatsApp Template Message",
	}
	mcp.AddTool(s.mcp, tool, toolHandler)
}

func (s *Server) addSendStickerTool() {
	toolHandler := mcp.ToolHandlerFor[*SendStickerRequest, *Response](s.HandleSendSticker)
	tool := &mcp.Tool{
		Description:  "Sends a sticker message to a WhatsApp number.",
		InputSchema:  s.schemas.sendStickerRequest,
		Name:         "whatsapp-send-sticker",
		OutputSchema: s.schemas.response,
		Title:        "Send WhatsApp Sticker Message",
	}
	mcp.AddTool(s.mcp, tool, toolHandler)
}

func (s *Server) addSendContactsTool() {
	toolHandler := mcp.ToolHandlerFor[*SendContactsRequest, *Response](s.HandleSendContacts)
	tool := &mcp.Tool{
		Description:  "Sends contact information to a WhatsApp number.",
		InputSchema:  s.schemas.sendContactsRequest,
		Name:         "whatsapp-send-contacts",
		OutputSchema: s.schemas.response,
		Title:        "Send WhatsApp Contacts Message",
	}
	mcp.AddTool(s.mcp, tool, toolHandler)
}

func (s *Server) addSendInteractiveTool() {
	toolHandler := mcp.ToolHandlerFor[*SendInteractiveRequest, *Response](s.HandleSendInteractive)
	tool := &mcp.Tool{
		Description:  "Sends an interactive message to a WhatsApp number (buttons, CTAs, etc.).",
		InputSchema:  s.schemas.sendInteractiveRequest,
		Name:         "whatsapp-send-interactive",
		OutputSchema: s.schemas.response,
		Title:        "Send WhatsApp Interactive Message",
	}
	mcp.AddTool(s.mcp, tool, toolHandler)
}

func (s *Server) HandleSendText(ctx context.Context, request *mcp.CallToolRequest, input *SendTextRequest) (*mcp.CallToolResult, *Response, error) {
	sessionID := request.GetSession().ID()
	s.logger.LogAttrs(ctx, slog.LevelInfo, "handling send text request", slog.String("sessionID", sessionID))

	output, err := s.SendText(ctx, input)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, nil, err
	}

	result := &mcp.CallToolResult{
		IsError: false,
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf(
				"Message sent to %s. The message ID is %s.",
				input.Recipient,
				output.MessageID,
			)},
		},
	}
	return result, output, nil
}

func (s *Server) HandleRequestLocation(ctx context.Context, request *mcp.CallToolRequest, input *RequestLocationRequest) (*mcp.CallToolResult, *Response, error) {
	sessionID := request.GetSession().ID()
	s.logger.LogAttrs(ctx, slog.LevelInfo, "handling request location request", slog.String("sessionID", sessionID))

	output, err := s.RequestLocation(ctx, input)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, nil, err
	}

	result := &mcp.CallToolResult{
		IsError: false,
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf(
				"Location request sent to %s. The message ID is %s.",
				input.Recipient,
				output.MessageID,
			)},
		},
	}
	return result, output, nil
}

func (s *Server) HandleSendLocation(ctx context.Context, request *mcp.CallToolRequest, input *SendLocationRequest) (*mcp.CallToolResult, *Response, error) {
	sessionID := request.GetSession().ID()
	s.logger.LogAttrs(ctx, slog.LevelInfo, "handling send location request", slog.String("sessionID", sessionID))

	output, err := s.SendLocation(ctx, input)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, nil, err
	}

	result := &mcp.CallToolResult{
		IsError: false,
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf(
				"Location sent to %s. The message ID is %s.",
				input.Recipient,
				output.MessageID,
			)},
		},
	}
	return result, output, nil
}

func (s *Server) HandleSendImage(ctx context.Context, request *mcp.CallToolRequest, input *SendImageRequest) (*mcp.CallToolResult, *Response, error) {
	sessionID := request.GetSession().ID()
	s.logger.LogAttrs(ctx, slog.LevelInfo, "handling send image request", slog.String("sessionID", sessionID))

	output, err := s.SendImage(ctx, input)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, nil, err
	}

	result := &mcp.CallToolResult{
		IsError: false,
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf(
				"Image sent to %s. The message ID is %s.",
				input.Recipient,
				output.MessageID,
			)},
		},
	}
	return result, output, nil
}

func (s *Server) HandleSendDocument(ctx context.Context, request *mcp.CallToolRequest, input *SendDocumentRequest) (*mcp.CallToolResult, *Response, error) {
	sessionID := request.GetSession().ID()
	s.logger.LogAttrs(ctx, slog.LevelInfo, "handling send document request", slog.String("sessionID", sessionID))

	output, err := s.SendDocument(ctx, input)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, nil, err
	}

	result := &mcp.CallToolResult{
		IsError: false,
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf(
				"Document sent to %s. The message ID is %s.",
				input.Recipient,
				output.MessageID,
			)},
		},
	}
	return result, output, nil
}

func (s *Server) HandleSendAudio(ctx context.Context, request *mcp.CallToolRequest, input *SendAudioRequest) (*mcp.CallToolResult, *Response, error) {
	sessionID := request.GetSession().ID()
	s.logger.LogAttrs(ctx, slog.LevelInfo, "handling send audio request", slog.String("sessionID", sessionID))

	output, err := s.SendAudio(ctx, input)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, nil, err
	}

	result := &mcp.CallToolResult{
		IsError: false,
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf(
				"Audio sent to %s. The message ID is %s.",
				input.Recipient,
				output.MessageID,
			)},
		},
	}
	return result, output, nil
}

func (s *Server) HandleSendVideo(ctx context.Context, request *mcp.CallToolRequest, input *SendVideoRequest) (*mcp.CallToolResult, *Response, error) {
	sessionID := request.GetSession().ID()
	s.logger.LogAttrs(ctx, slog.LevelInfo, "handling send video request", slog.String("sessionID", sessionID))

	output, err := s.SendVideo(ctx, input)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, nil, err
	}

	result := &mcp.CallToolResult{
		IsError: false,
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf(
				"Video sent to %s. The message ID is %s.",
				input.Recipient,
				output.MessageID,
			)},
		},
	}
	return result, output, nil
}

func (s *Server) HandleSendReaction(ctx context.Context, request *mcp.CallToolRequest, input *SendReactionRequest) (*mcp.CallToolResult, *Response, error) {
	sessionID := request.GetSession().ID()
	s.logger.LogAttrs(ctx, slog.LevelInfo, "handling send reaction request", slog.String("sessionID", sessionID))

	output, err := s.SendReaction(ctx, input)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, nil, err
	}

	result := &mcp.CallToolResult{
		IsError: false,
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf(
				"Reaction sent to %s. The message ID is %s.",
				input.Recipient,
				output.MessageID,
			)},
		},
	}
	return result, output, nil
}

func (s *Server) HandleSendTemplate(ctx context.Context, request *mcp.CallToolRequest, input *SendTemplateRequest) (*mcp.CallToolResult, *Response, error) {
	sessionID := request.GetSession().ID()
	s.logger.LogAttrs(ctx, slog.LevelInfo, "handling send template request", slog.String("sessionID", sessionID))

	output, err := s.SendTemplate(ctx, input)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, nil, err
	}

	result := &mcp.CallToolResult{
		IsError: false,
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf(
				"Template sent to %s. The message ID is %s.",
				input.Recipient,
				output.MessageID,
			)},
		},
	}
	return result, output, nil
}

func (s *Server) HandleSendSticker(ctx context.Context, request *mcp.CallToolRequest, input *SendStickerRequest) (*mcp.CallToolResult, *Response, error) {
	sessionID := request.GetSession().ID()
	s.logger.LogAttrs(ctx, slog.LevelInfo, "handling send sticker request", slog.String("sessionID", sessionID))

	output, err := s.SendSticker(ctx, input)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, nil, err
	}

	result := &mcp.CallToolResult{
		IsError: false,
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf(
				"Sticker sent to %s. The message ID is %s.",
				input.Recipient,
				output.MessageID,
			)},
		},
	}
	return result, output, nil
}

func (s *Server) HandleSendContacts(ctx context.Context, request *mcp.CallToolRequest, input *SendContactsRequest) (*mcp.CallToolResult, *Response, error) {
	sessionID := request.GetSession().ID()
	s.logger.LogAttrs(ctx, slog.LevelInfo, "handling send contacts request", slog.String("sessionID", sessionID))

	output, err := s.SendContacts(ctx, input)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, nil, err
	}

	result := &mcp.CallToolResult{
		IsError: false,
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf(
				"Contacts sent to %s. The message ID is %s.",
				input.Recipient,
				output.MessageID,
			)},
		},
	}
	return result, output, nil
}

func (s *Server) HandleSendInteractive(ctx context.Context, request *mcp.CallToolRequest, input *SendInteractiveRequest) (*mcp.CallToolResult, *Response, error) {
	sessionID := request.GetSession().ID()
	s.logger.LogAttrs(ctx, slog.LevelInfo, "handling send interactive request", slog.String("sessionID", sessionID))

	output, err := s.SendInteractiveMessage(ctx, input)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, nil, err
	}

	result := &mcp.CallToolResult{
		IsError: false,
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf(
				"Interactive message sent to %s. The message ID is %s.",
				input.Recipient,
				output.MessageID,
			)},
		},
	}
	return result, output, nil
}
