/*
 *  Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy of this software
 *  and associated documentation files (the “Software”), to deal in the Software without restriction,
 *  including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
 *  and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
 *  subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all copies or substantial
 *  portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
 *  LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
 *  IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
 *  WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
 *  SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package errors

import (
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
	// UserMsg: The message to display to the user. The language of the message is based on the
	// locale of the API request.
	// UserTitle: The title of the dialog, if shown. The language of the message is based on the locale
	// of the API request.
	// FBTraceID: Internal support identifier. When reporting a bug related to a Graph API call, include
	// the fbtrace_id to help us find log data for debugging.
	// Example of error response
	//
	//	"error": {
	//	        "message": "(#131030) Recipient phone number not in allowed list",
	//	        "type": "OAuthException",
	//	        "code": 131030,
	//	        "error_data": {
	//	             "messaging_product": "whatsapp",
	//	            "details": "Recipient phone number not in allowed list: Add recipient phone number to recipient
	//	                        list and try again."
	//	        },
	//	        "error_subcode": 2655007,
	//	 	   "error_user_title": "Recipient phone number not in allowed list",
	//	        "error_user_msg": "Add recipient phone number to recipient list and try again.",
	//	        "fbtrace_id": "AI5Ob2z72R0JAUB5zOF-nao"
	//	}
	Error struct {
		Message   string     `json:"message,omitempty"`
		Type      string     `json:"type,omitempty"`
		Details   string     `json:"details,omitempty"`
		Code      int        `json:"code,omitempty"`
		Data      *ErrorData `json:"error_data,omitempty"`
		Subcode   int        `json:"error_subcode,omitempty"`
		UserTitle string     `json:"error_user_title,omitempty"`
		UserMsg   string     `json:"error_user_msg,omitempty"`
		FBTraceID string     `json:"fbtrace_id,omitempty"`
		Href      string     `json:"href,omitempty"`
	}

	// ErrorData represents additional information about the error.
	ErrorData struct {
		MessagingProduct string `json:"messaging_product,omitempty"`
		Details          string `json:"details,omitempty"`
	}
)

func (e *ErrorData) String() string {
	if e.MessagingProduct == "" && e.Details == "" {
		return "<nil>"
	}
	var b strings.Builder
	if e.MessagingProduct != "" {
		b.WriteString("Messaging Product: " + e.MessagingProduct)
	}
	if e.Details != "" {
		b.WriteString(", Details: " + e.Details)
	}

	return b.String()
}

func (e *Error) String() string {
	if e == nil {
		return "<nil>"
	}
	var b strings.Builder
	if e.Message != "" {
		b.WriteString("Message: " + e.Message)
	}
	if e.Type != "" {
		b.WriteString(", Type: " + e.Type)
	}
	if e.Code != 0 {
		b.WriteString(", StatusCode: " + strconv.Itoa(e.Code))
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

	if e.Href != "" {
		b.WriteString(", Href: " + e.Href)
	}

	if e.Details != "" {
		b.WriteString(", Details: " + e.Details)
	}

	return b.String()
}

func (e *Error) Error() string {
	return "whatsapp error: " + strings.ToLower(e.String())
}

type ValidationError struct {
	Err         string `json:"error"`
	ErrorType   string `json:"error_type"`
	Message     string `json:"message"`
	LineStart   int    `json:"line_start"`
	LineEnd     int    `json:"line_end"`
	ColumnStart int    `json:"column_start"`
	ColumnEnd   int    `json:"column_end"`
}

func (e ValidationError) Error() string {
	return e.Err
}
