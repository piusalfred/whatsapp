//  Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
//
//  Permission is hereby granted, free of charge, to any person obtaining a copy of this software
//  and associated documentation files (the "Software"), to deal in the Software without restriction,
//  including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
//  and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
//  subject to the following conditions:
//
//  The above copyright notice and this permission notice shall be included in all copies or substantial
//  portions of the Software.
//
//  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
//  LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
//  IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
//  WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
//  SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package webhooks

import (
	"context"
	"fmt"

	"github.com/piusalfred/whatsapp/message"
	"github.com/piusalfred/whatsapp/message/media"
)

// MassageTypeHandler is the single entry point for all WhatsApp message dispatch.
// Each field is a handler for one message type or sub-type. Leave a field nil
// to silently skip that type during dispatch.
//
// Media types delegate to [MediaHandler]; interactive messages delegate to
// [InteractiveHandler]. Unknown types fall through to [Fallback].
//
// Usage:
//
//	mh := &MessageHandlerX{}
//	mh.Media = &MediaHandler{}          // enable media dispatch
//	mh.Interactive = &InteractiveHandler{}
//	mh.Text = myTextHandler
//	mh.Fallback = catchAll              // handle anything not explicitly registered
//	h := webhooks.NewHandler()
//	h.SetMessagesHandler(mh)            // wire into Handler (future API)
type MassageTypeHandler struct {
	Media            *MediaHandler
	Interactive      *InteractiveHandler
	Text             MessageHandler[Text]
	Order            MessageHandler[Order]
	Button           MessageHandler[Button]
	System           MessageHandler[System]
	Reaction         MessageHandler[media.Reaction]
	Location         MessageHandler[media.Location]
	Contacts         MessageHandler[message.Contacts]
	Revoke           MessageHandler[Revoke]
	Edit             MessageHandler[Edit]
	RequestWelcome   MessageHandler[Message]
	Referral         MessageHandler[ReferralNotification]
	CustomerIDChange MessageHandler[Identity]
	Unknown          MessageErrorsHandler
	Unsupported      MessageErrorsHandler
	ProductInquiry   MessageHandler[Text]
	Fallback         MessageHandler[Message]
}

//nolint:cyclop,funlen,gocognit,gocyclo,wrapcheck // dispatch switch; user handlers own error context
func (mh *MassageTypeHandler) Handle(
	ctx context.Context,
	nctx *MessageNotificationContext,
	message *Message,
) error {
	info := &MessageInfo{
		From:             message.From,
		MessageID:        message.ID,
		Timestamp:        message.Timestamp,
		Type:             message.Type,
		GroupID:          message.GroupID,
		Context:          message.Context,
		IsAReply:         message.IsAReply(),
		IsForwarded:      message.IsForwarded(),
		IsProductInquiry: message.IsProductInquiry(),
		IsReferral:       message.IsReferral(),
	}

	msgType := ParseMessageType(message.Type)
	switch msgType {
	case MessageTypeAudio, MessageTypeVideo, MessageTypeImage, MessageTypeDocument, MessageTypeSticker:
		return mh.HandleMediaMessage(ctx, nctx, info, message)
	case MessageTypeInteractive:
		return mh.HandleInteractiveMessage(ctx, nctx, info, message)
	case MessageTypeText:
		return mh.handleText(ctx, nctx, info, message)
	case MessageTypeSystem:
		return mh.handleSystem(ctx, nctx, info, message)
	case MessageTypeOrder:
		if mh.Order == nil {
			return nil
		}
		if err := mh.Order.Handle(ctx, nctx, info, message.Order); err != nil {
			return fmt.Errorf("handle order message: %w", err)
		}
		return nil
	case MessageTypeButton:
		if mh.Button == nil {
			return nil
		}
		if err := mh.Button.Handle(ctx, nctx, info, message.Button); err != nil {
			return fmt.Errorf("handle button message: %w", err)
		}
		return nil
	case MessageTypeReaction:
		if mh.Reaction == nil {
			return nil
		}
		if err := mh.Reaction.Handle(ctx, nctx, info, message.Reaction); err != nil {
			return fmt.Errorf("handle reaction message: %w", err)
		}
		return nil
	case MessageTypeLocation:
		if mh.Location == nil {
			return nil
		}
		if err := mh.Location.Handle(ctx, nctx, info, message.Location); err != nil {
			return fmt.Errorf("handle location message: %w", err)
		}
		return nil
	case MessageTypeContacts:
		if mh.Contacts == nil {
			return nil
		}
		if err := mh.Contacts.Handle(ctx, nctx, info, message.Contacts); err != nil {
			return fmt.Errorf("handle contacts message: %w", err)
		}
		return nil
	case MessageTypeRevoke:
		if mh.Revoke == nil {
			return nil
		}
		if err := mh.Revoke.Handle(ctx, nctx, info, message.Revoke); err != nil {
			return fmt.Errorf("handle revoke message: %w", err)
		}
		return nil
	case MessageTypeEdit:
		if mh.Edit == nil {
			return nil
		}
		if err := mh.Edit.Handle(ctx, nctx, info, message.Edit); err != nil {
			return fmt.Errorf("handle edit message: %w", err)
		}
		return nil
	case MessageTypeRequestWelcome:
		if mh.RequestWelcome == nil {
			return nil
		}
		if err := mh.RequestWelcome.Handle(ctx, nctx, info, message); err != nil {
			return fmt.Errorf("handle request welcome: %w", err)
		}
		return nil
	case MessageTypeUnknown:
		if mh.Unknown == nil {
			return nil
		}
		if err := mh.Unknown.Handle(ctx, nctx, info, message.Errors); err != nil {
			return fmt.Errorf("handle error message: %w", err)
		}
		return nil
	case MessageTypeUnsupported:
		if mh.Unsupported == nil {
			return nil
		}
		if err := mh.Unsupported.Handle(ctx, nctx, info, message.Errors); err != nil {
			return fmt.Errorf("handle unsupported message: %w", err)
		}
		return nil
	default:
		if mh.Fallback != nil {
			return mh.Fallback.Handle(ctx, nctx, info, message)
		}
		return nil
	}
}

func (mh *MassageTypeHandler) handleText(ctx context.Context, nctx *MessageNotificationContext,
	info *MessageInfo, message *Message,
) error {
	if info.IsReferral && mh.Referral != nil {
		ref := &ReferralNotification{Text: message.Text, Referral: message.Referral}
		if err := mh.Referral.Handle(ctx, nctx, info, ref); err != nil {
			return fmt.Errorf("handle referral message: %w", err)
		}
		return nil
	}
	if info.IsProductInquiry && mh.ProductInquiry != nil {
		if err := mh.ProductInquiry.Handle(ctx, nctx, info, message.Text); err != nil {
			return fmt.Errorf("handle product inquiry: %w", err)
		}
		return nil
	}
	if mh.Text == nil {
		return nil
	}
	if err := mh.Text.Handle(ctx, nctx, info, message.Text); err != nil {
		return fmt.Errorf("handle text message: %w", err)
	}
	return nil
}

func (mh *MassageTypeHandler) handleSystem(ctx context.Context, nctx *MessageNotificationContext,
	info *MessageInfo, message *Message,
) error {
	if message.System != nil && message.System.Type == "user_changed_number" && mh.CustomerIDChange != nil {
		if err := mh.CustomerIDChange.Handle(ctx, nctx, info, message.Identity); err != nil {
			return fmt.Errorf("handle customer ID change: %w", err)
		}
		return nil
	}
	if mh.System == nil {
		return nil
	}
	if err := mh.System.Handle(ctx, nctx, info, message.System); err != nil {
		return fmt.Errorf("handle system message: %w", err)
	}
	return nil
}

// HandleMediaMessage dispatches the media message to [MediaHandler] if set.
// Returns nil when mh.Media is nil.
func (mh *MassageTypeHandler) HandleMediaMessage(ctx context.Context, nctx *MessageNotificationContext,
	info *MessageInfo, msg *Message,
) error {
	if mh.Media == nil {
		return nil
	}

	if err := mh.Media.Handle(ctx, nctx, info, msg); err != nil {
		return fmt.Errorf("handle media message: %w", err)
	}

	return nil
}

// HandleInteractiveMessage dispatches the interactive message to
// [InteractiveHandler] if set. Returns nil when mh.Interactive is nil.
func (mh *MassageTypeHandler) HandleInteractiveMessage(ctx context.Context, nctx *MessageNotificationContext,
	info *MessageInfo, msg *Message,
) error {
	if mh.Interactive == nil {
		return nil
	}

	if err := mh.Interactive.Handle(ctx, nctx, info, msg); err != nil {
		return fmt.Errorf("handle interactive message: %w", err)
	}

	return nil
}

// MediaHandler groups handlers for media message types (audio, video, image,
// document, sticker). All fields are [MessageHandler[media.Info]]. Leave a
// field nil to silently skip that media type during dispatch.
//
// MediaHandler has no fallback — unhandled media types return an error.
// A catch-all for unrecognized message types belongs in the message-level
// handler, not here.
//
// Usage:
//
//	mh := &MediaHandler{}
//	mh.OnAudio(myAudioHandler)
type MediaHandler struct {
	Audio    MessageHandler[media.Info]
	Video    MessageHandler[media.Info]
	Image    MessageHandler[media.Info]
	Document MessageHandler[media.Info]
	Sticker  MessageHandler[media.Info]
}

// OnAudio sets the handler for audio messages.
func (mh *MediaHandler) OnAudio(
	fn func(ctx context.Context, nctx *MessageNotificationContext, info *MessageInfo, data *media.Info) error,
) {
	mh.Audio = MessageHandlerFunc[media.Info](fn)
}

// OnVideo sets the handler for video messages.
func (mh *MediaHandler) OnVideo(
	fn func(ctx context.Context, nctx *MessageNotificationContext, info *MessageInfo, data *media.Info) error,
) {
	mh.Video = MessageHandlerFunc[media.Info](fn)
}

// OnImage sets the handler for image messages.
func (mh *MediaHandler) OnImage(
	fn func(ctx context.Context, nctx *MessageNotificationContext, info *MessageInfo, data *media.Info) error,
) {
	mh.Image = MessageHandlerFunc[media.Info](fn)
}

// OnDocument sets the handler for document messages.
func (mh *MediaHandler) OnDocument(
	fn func(ctx context.Context, nctx *MessageNotificationContext, info *MessageInfo, data *media.Info) error,
) {
	mh.Document = MessageHandlerFunc[media.Info](fn)
}

// OnSticker sets the handler for sticker messages.
func (mh *MediaHandler) OnSticker(
	fn func(ctx context.Context, nctx *MessageNotificationContext, info *MessageInfo, data *media.Info) error,
) {
	mh.Sticker = MessageHandlerFunc[media.Info](fn)
}

// Handle dispatches the media message to the correct handler based on msg.Type.
//
// If the dedicated handler is nil, the message is silently skipped. Unknown
// media types return an error — they should have been caught earlier by the
// message dispatch switch.
//
//nolint:gocognit // typed dispatch switch
func (mh *MediaHandler) Handle(
	ctx context.Context,
	nctx *MessageNotificationContext,
	info *MessageInfo,
	msg *Message,
) error {
	mt := ParseMessageType(msg.Type)
	switch mt {
	case MessageTypeAudio:
		if mh.Audio == nil {
			return nil
		}
		if err := mh.Audio.Handle(ctx, nctx, info, msg.Audio); err != nil {
			return fmt.Errorf("handle audio message: %w", err)
		}

	case MessageTypeVideo:
		if mh.Video == nil {
			return nil
		}
		if err := mh.Video.Handle(ctx, nctx, info, msg.Video); err != nil {
			return fmt.Errorf("handle video message: %w", err)
		}

	case MessageTypeImage:
		if mh.Image == nil {
			return nil
		}
		if err := mh.Image.Handle(ctx, nctx, info, msg.Image); err != nil {
			return fmt.Errorf("handle image message: %w", err)
		}

	case MessageTypeDocument:
		if mh.Document == nil {
			return nil
		}
		if err := mh.Document.Handle(ctx, nctx, info, msg.Document); err != nil {
			return fmt.Errorf("handle document message: %w", err)
		}

	case MessageTypeSticker:
		if mh.Sticker == nil {
			return nil
		}
		if err := mh.Sticker.Handle(ctx, nctx, info, msg.Sticker); err != nil {
			return fmt.Errorf("handle sticker message: %w", err)
		}

	default:
		return fmt.Errorf("media message handler: %w: %s", ErrUnrecognizedMessageType, mt)
	}

	return nil
}

// InteractiveHandler groups handlers for interactive message sub-types
// (list reply, button reply, flow completion, address submission) plus a
// general handler for unrecognized interactive types.
//
// Leave a field nil to skip that subtype. The general [Interactive] handler
// catches any interactive type without a dedicated handler.
//
// Usage:
//
//	ih := &InteractiveHandler{}
//	ih.OnButtonReply(myButtonReplyHandler)
//	ih.OnFallback(myGeneralHandler) // catch-all for unrecognized types
type InteractiveHandler struct {
	ListReply         MessageHandler[ListReply]
	ButtonReply       MessageHandler[ButtonReply]
	FlowCompletion    MessageHandler[NFMReply]
	AddressSubmission MessageHandler[NFMReply]
	Fallback          MessageHandler[Interactive]
}

// OnListReply sets the handler for list reply interactive messages.
func (ih *InteractiveHandler) OnListReply(
	fn func(ctx context.Context, nctx *MessageNotificationContext, info *MessageInfo, data *ListReply) error,
) {
	ih.ListReply = MessageHandlerFunc[ListReply](fn)
}

// OnButtonReply sets the handler for button reply interactive messages.
func (ih *InteractiveHandler) OnButtonReply(
	fn func(ctx context.Context, nctx *MessageNotificationContext, info *MessageInfo, data *ButtonReply) error,
) {
	ih.ButtonReply = MessageHandlerFunc[ButtonReply](fn)
}

// OnFlowCompletion sets the handler for flow completion (nfm_reply) messages.
func (ih *InteractiveHandler) OnFlowCompletion(
	fn func(ctx context.Context, nctx *MessageNotificationContext, info *MessageInfo, data *NFMReply) error,
) {
	ih.FlowCompletion = MessageHandlerFunc[NFMReply](fn)
}

// OnAddressSubmission sets the handler for address submission messages.
func (ih *InteractiveHandler) OnAddressSubmission(
	fn func(ctx context.Context, nctx *MessageNotificationContext, info *MessageInfo, data *NFMReply) error,
) {
	ih.AddressSubmission = MessageHandlerFunc[NFMReply](fn)
}

// OnFallback sets the catch-all handler for interactive messages without
// a dedicated subtype handler.
func (ih *InteractiveHandler) OnFallback(
	fn func(ctx context.Context, nctx *MessageNotificationContext, info *MessageInfo, data *Interactive) error,
) {
	ih.Fallback = MessageHandlerFunc[Interactive](fn)
}

// Handle dispatches an interactive message to the correct handler based on
// msg.Interactive.Type. Subtypes without a dedicated handler fall through to
// the general [Interactive] handler. Messages with a nil Interactive payload
// are silently skipped.
//
//nolint:gocognit // typed dispatch switch
func (ih *InteractiveHandler) Handle(
	ctx context.Context,
	nctx *MessageNotificationContext,
	info *MessageInfo,
	msg *Message,
) error {
	if msg.Interactive == nil {
		return nil
	}

	switch msg.Interactive.Type {
	case InteractiveTypeListReply:
		if ih.ListReply != nil {
			if err := ih.ListReply.Handle(ctx, nctx, info, msg.Interactive.ListReply); err != nil {
				return fmt.Errorf("handle list reply: %w", err)
			}
			return nil
		}

	case InteractiveTypeButtonReply:
		if ih.ButtonReply != nil {
			if err := ih.ButtonReply.Handle(ctx, nctx, info, msg.Interactive.ButtonReply); err != nil {
				return fmt.Errorf("handle button reply: %w", err)
			}
			return nil
		}

	case InteractiveTypeNFMReply:
		if ih.FlowCompletion != nil {
			if err := ih.FlowCompletion.Handle(ctx, nctx, info, msg.Interactive.NFMReply); err != nil {
				return fmt.Errorf("handle flow completion: %w", err)
			}
			return nil
		}

	case InteractiveAddressSubmission:
		if ih.AddressSubmission != nil {
			if err := ih.AddressSubmission.Handle(ctx, nctx, info, msg.Interactive.NFMReply); err != nil {
				return fmt.Errorf("handle address submission: %w", err)
			}
			return nil
		}
	}

	if ih.Fallback != nil {
		if err := ih.Fallback.Handle(ctx, nctx, info, msg.Interactive); err != nil {
			return fmt.Errorf("handle interactive message: %w", err)
		}
	}

	return nil
}
