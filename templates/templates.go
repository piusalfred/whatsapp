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

// Package templates provides a set of functions for creating and managing templates.
//
// LINK: https://developers.facebook.com/docs/whatsapp/business-management-api/message-templates
package templates // import "github.com/piusalfred/whatsapp/templates"

const (
	StatusApproved Status = "APPROVED"
	StatusPending  Status = "PENDING"
	StatusRejected Status = "REJECTED"
)

const (
	CategoryAuthentication Category = "AUTHENTICATION"
	CategoryMarketing      Category = "MARKETING"
	CategoryUtility        Category = "UTILITY"
)

const (
	TemplateEndpoint = "message_templates"
)

type (
	// Status is the status of the template. There are 3 possible values:
	//
	// APPROVED — The template has passed template review and been approved, and can now be sent
	// in template messages.
	//
	// PENDING — The template passed category validation and is undergoing template review.
	//
	// REJECTED — The template failed category validation or template review. You can request the
	// rejected_reason field on the template to get the reason.
	Status string

	// Category is the category of the template. Templates must be categorized as one of the following
	// categories. Categories factor into pricing and the category you designate will be validated at the
	// time of template creation.
	//
	// - AUTHENTICATION
	//
	// - MARKETING
	//
	// - UTILITY
	//
	// For more LINK: https://developers.facebook.com/docs/whatsapp/updates-to-pricing/new-template-guidelines
	Category string

	// CreateRequest is the request body for creating a template.
	//
	// SUPPORTED LANGUAGES:
	// https://developers.facebook.com/docs/whatsapp/business-management-api/message-templates/supported-languages
	CreateRequest struct {
		Name                string       `json:"name,omitempty"`
		Language            string       `json:"language,omitempty"`
		Category            Category     `json:"category,omitempty"`
		AllowCategoryChange bool         `json:"allow_category_change"`
		Components          []*Component `json:"components,omitempty"`
	}

	// CreateResponse is the response body for creating a template.
	CreateResponse struct {
		ID       string   `json:"id,omitempty"`
		Status   Status   `json:"status,omitempty"`
		Category Category `json:"category,omitempty"`
	}

	Component struct{}
)
