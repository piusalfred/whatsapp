package errors

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type (

	// Error represents a WhatsApp error returned by the API when a request fails.
	// It implements the error interface.
	//
	// Message represent a human-readable description of the error.
	// Code: An error code. Common values are listed below, along with common recovery tactics.
	// Data (optional): Additional information about the error.
	// Subcode: Additional information about the error. Common values are listed below.
	// UserMessage: The message to display to the user. The language of the message is based on the locale of the API request.
	// UserTitle: The title of the dialog, if shown. The language of the message is based on the locale of the API request.
	// FbTraceID: Internal support identifier. When reporting a bug related to a Graph API call, include the fbtrace_id to help us find log data for debugging.
	// Example of error response
	//
	//	"error": {
	//	        "message": "(#131030) Recipient phone number not in allowed list",
	//	        "type": "OAuthException",
	//	        "code": 131030,
	//	        "error_data": {
	//	             "messaging_product": "whatsapp",
	//	            "details": "Recipient phone number not in allowed list: Add recipient phone number to recipient list and try again."
	//	        },
	//	        "error_subcode": 2655007,
	//	 	   "error_user_title": "Recipient phone number not in allowed list",
	//	        "error_user_msg": "Add recipient phone number to recipient list and try again.",
	//	        "fbtrace_id": "AI5Ob2z72R0JAUB5zOF-nao"
	//	}
	Error struct {
		Message   string     `json:"message,omitempty"`
		Type      string     `json:"type,omitempty"`
		Code      int        `json:"code,omitempty"`
		Data      *ErrorData `json:"error_data,omitempty"`
		Subcode   int        `json:"error_subcode,omitempty"`
		UserTitle string     `json:"error_user_title,omitempty"`
		UserMsg   string     `json:"error_user_msg,omitempty"`
		FBTraceID string     `json:"fbtrace_id,omitempty"`
	}

	// ErrorData represents additional information about the error.
	ErrorData struct {
		MessagingProduct string `json:"messaging_product,omitempty"`
		Details          string `json:"details,omitempty"`
	}

	// ErrorResponse struct {
	// 	Err  *Error `json:"error,omitempty"`
	// 	Code int    `json:"code,omitempty"` // http status code
	// }
)

// IsError checks if the error is a WhatsApp error.
func IsError(err error) bool {
	var e *Error
	return errors.As(err, &e)
}

func (e *ErrorData) String() string {
	if e.MessagingProduct == "" && e.Details == "" {
		return "<nil>"
	} else {
		var b strings.Builder
		if e.MessagingProduct != "" {
			b.WriteString("Messaging Product: " + e.MessagingProduct)
		}

		if e.Details != "" {
			b.WriteString(", Details: " + e.Details)
		}

		return b.String()
	}
}

func (e *Error) String() string {
	if e == nil {
		return "<nil>"
	} else {
		var b strings.Builder
		if e.Message != "" {
			b.WriteString("Message: " + e.Message)
		}

		if e.Type != "" {
			b.WriteString(", Type: " + e.Type)
		}

		if e.Code != 0 {
			b.WriteString(", Code: " + strconv.Itoa(e.Code))
		}

		if e.Data != nil {
			b.WriteString(", Data: " + e.Data.String())
		}

		if e.Subcode != 0 {
			b.WriteString(", Subcode: " + strconv.Itoa(e.Subcode))
		}

		if e.UserTitle != "" {
			b.WriteString(", UserTitle: " + e.UserTitle)
		}

		if e.UserMsg != "" {
			b.WriteString(", UserMsg: " + e.UserMsg)
		}

		if e.FBTraceID != "" {
			b.WriteString(", FBTraceID: " + e.FBTraceID)
		}

		return b.String()
	}
}

func (e *Error) Error() string {
	return fmt.Sprintf("whatsapp: %s", strings.ToLower(e.String()))
}
