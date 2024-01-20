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

// Package health describes how to determine whether or not you can send messages successfully
// using a given API resource.
// The following nodes have a health_status field:
//
//   - WhatsApp Business Account
//
//   - WhatsApp Business Phone Number
//
//   - WhatsApp Message Template
//
// If you request the health_status field on any of these nodes, the API will return a summary
// of the messaging health of all the nodes involved in messaging requests if using the targeted
// node. This summary indicates if you will be able to use the API to send messages successfully,
// if you will have limited success due to some limitation on one or more nodes, or if you will be
// prevented from messaging entirely.
//
//	When you attempt to send a message, multiple nodes are involved, including the app,
//	the business that owns or has claimed it, a WABA, a business phone number, and a template (if sending a template message).
//
//	Each of these nodes can have one of the following health statuses assigned to the can_send_message property:
//
//	AVAILABLE: Indicates that the node meets all messaging requirements.
//	LIMITED: Indicates that the node meets messaging requirements, but has some limitations.
//	If a given node has this value, additional info will be included.
//	BLOCKED: Indicates that the node does not meet one or more messaging requirements. If a given node
//	has this value, the errors property will be included which describes the error and a possible solution.
//	Overall Status
//
//	The overall health status property (health_status.can_send_message) will be set as follows:
//
//	If one or more nodes is blocked, it will be set to BLOCKED.
//	If no nodes are blocked, but one or more nodes is limited, it will be set to LIMITED.
//	If all nodes are available, it will be set to AVAILABLE.
package health

type Status struct {
	CanSendMessage string    `json:"can_send_message,omitempty"`
	Entities       []*Entity `json:"entities,omitempty"`
}

type Entity struct {
	EntityType     string   `json:"entity_type,omitempty"`
	ID             string   `json:"id,omitempty"`
	CanSendMessage string   `json:"can_send_message,omitempty"`
	AdditionalInfo []string `json:"additional_info,omitempty"` // Optional field
	Errors         []Error  `json:"errors,omitempty"`          // Optional field
}

type Error struct {
	ErrorCode        int    `json:"error_code"`
	ErrorDescription string `json:"error_description"`
	PossibleSolution string `json:"possible_solution"`
}

type Response struct {
	HealthStatus *Status `json:"health_status,omitempty"`
	ID           string  `json:"id,omitempty"`
}
