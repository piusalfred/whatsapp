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

package flow

// https://developers.facebook.com/docs/whatsapp/flows/reference/flowswebhooks

import (
	"context"
	"net/http"

	"github.com/piusalfred/libwhatsapp/webhooks"
)

const (
	StatusDraft               = "DRAFT"
	StatusPublished           = "PUBLISHED"
	StatusDeprecated          = "DEPRECATED"
	StatusBlocked             = "BLOCKED"
	StatusThrottled           = "THROTTLED"
	EventFlowStatusChange     = "FLOW_STATUS_CHANGE"
	EventEndpointErrorRate    = "ENDPOINT_ERROR_RATE"
	EventEndpointLatency      = "ENDPOINT_LATENCY"
	EventEndpointAvailability = "ENDPOINT_AVAILABILITY"
	EventClientErrorRate      = "CLIENT_ERROR_RATE"
)

var _ NotificationHandler = (*Handlers)(nil)

type Handlers struct {
	FlowStatusChangeHandler     EventHandlerFunc[StatusChangeDetails]
	ClientErrorRateHandler      EventHandlerFunc[ClientErrorRateDetails]
	EndpointErrorRateHandler    EventHandlerFunc[EndpointErrorRateDetails]
	EndpointLatencyHandler      EventHandlerFunc[EndpointLatencyDetails]
	EndpointAvailabilityHandler EventHandlerFunc[EndpointAvailabilityDetails]
}

func (handlers *Handlers) HandleNotification(ctx context.Context, notification *Notification) *webhooks.Response {
	if err := handlers.dispatchNotification(ctx, notification); err != nil {
		return &webhooks.Response{StatusCode: http.StatusInternalServerError}
	}

	return &webhooks.Response{StatusCode: http.StatusOK}
}

func (handlers *Handlers) dispatchNotification(ctx context.Context, notification *Notification) error {
	for _, entry := range notification.Entry {
		for _, change := range entry.Changes {
			notificationCtx := &NotificationContext{
				NotificationObject: notification.Object,
				EntryID:            entry.ID,
				EntryTime:          entry.Time,
				ChangeField:        change.Field,
				EventName:          change.Value.Event,
				EventMessage:       change.Value.Message,
				FlowID:             change.Value.FlowID,
			}

			switch change.Value.Event {
			case EventFlowStatusChange:
				details := &StatusChangeDetails{
					OldStatus: change.Value.OldStatus,
					NewStatus: change.Value.NewStatus,
				}
				return handlers.FlowStatusChangeHandler(ctx, notificationCtx, details)
			case EventClientErrorRate:
				details := &ClientErrorRateDetails{
					ErrorRate:  change.Value.ErrorRate,
					Threshold:  change.Value.Threshold,
					AlertState: change.Value.AlertState,
					Errors:     change.Value.Errors,
				}
				return handlers.ClientErrorRateHandler(ctx, notificationCtx, details)
			case EventEndpointErrorRate:
				details := &EndpointErrorRateDetails{
					ErrorRate:  change.Value.ErrorRate,
					Threshold:  change.Value.Threshold,
					AlertState: change.Value.AlertState,
					Errors:     change.Value.Errors,
				}
				return handlers.EndpointErrorRateHandler(ctx, notificationCtx, details)
			case EventEndpointLatency:
				details := &EndpointLatencyDetails{
					P50Latency:    change.Value.P50Latency,
					P90Latency:    change.Value.P90Latency,
					RequestsCount: change.Value.RequestsCount,
					Threshold:     change.Value.Threshold,
					AlertState:    change.Value.AlertState,
				}
				return handlers.EndpointLatencyHandler(ctx, notificationCtx, details)
			case EventEndpointAvailability:
				details := &EndpointAvailabilityDetails{
					Availability: change.Value.Availability,
					Threshold:    change.Value.Threshold,
					AlertState:   change.Value.AlertState,
				}
				return handlers.EndpointAvailabilityHandler(ctx, notificationCtx, details)
			}
		}
	}
	return nil
}

type (
	NotificationHandler     webhooks.NotificationHandler[Notification]
	NotificationHandlerFunc webhooks.NotificationHandlerFunc[Notification]
)

func (e NotificationHandlerFunc) HandleNotification(ctx context.Context, notification *Notification) *webhooks.Response {
	return e(ctx, notification)
}

type (
	EvenHandler[T any] interface {
		HandleEvent(ctx context.Context, ntx *NotificationContext, notification *T) error
	}

	EventHandlerFunc[T any] func(ctx context.Context, ntx *NotificationContext, notification *T) error
)

func (fn EventHandlerFunc[T]) HandleEvent(ctx context.Context, ntx *NotificationContext, notification *T) error {
	return fn(ctx, ntx, notification)
}

type (
	NotificationContext struct {
		NotificationObject string // Corresponds to the 'object' field
		EntryID            string // Corresponds to the 'id' field in Entry
		EntryTime          int64  // Corresponds to the 'time' field in Entry
		ChangeField        string // Corresponds to the 'field' in Changes
		EventName          string // Corresponds to 'event' field in Value
		EventMessage       string // Corresponds to 'message' field in Value
		FlowID             string // Corresponds to 'flow_id' field in Value
	}

	StatusChangeDetails struct {
		OldStatus string `json:"old_status,omitempty"`
		NewStatus string `json:"new_status,omitempty"`
	}

	ClientErrorRateDetails struct {
		ErrorRate  float64     `json:"error_rate,omitempty"`
		Threshold  int         `json:"threshold,omitempty"`
		AlertState string      `json:"alert_state,omitempty"`
		Errors     []ErrorInfo `json:"errors,omitempty"`
	}

	EndpointErrorRateDetails struct {
		ErrorRate  float64     `json:"error_rate,omitempty"`
		Threshold  int         `json:"threshold,omitempty"`
		AlertState string      `json:"alert_state,omitempty"`
		Errors     []ErrorInfo `json:"errors,omitempty"`
	}

	EndpointLatencyDetails struct {
		P50Latency    int    `json:"p50_latency,omitempty"`
		P90Latency    int    `json:"p90_latency,omitempty"`
		RequestsCount int    `json:"requests_count,omitempty"`
		Threshold     int    `json:"threshold,omitempty"`
		AlertState    string `json:"alert_state,omitempty"`
	}

	EndpointAvailabilityDetails struct {
		Availability int    `json:"availability,omitempty"`
		Threshold    int    `json:"threshold,omitempty"`
		AlertState   string `json:"alert_state,omitempty"`
	}

	Notification struct {
		Object string   `json:"object"`
		Entry  []*Entry `json:"entry"`
	}

	Entry struct {
		ID      string     `json:"id"`
		Time    int64      `json:"time"`
		Changes []*Changes `json:"changes"`
	}

	Changes struct {
		Value *Value `json:"value"`
		Field string `json:"field"`
	}

	Value struct {
		Event         string      `json:"event"`                    // Event type, e.g., "CLIENT_ERROR_RATE"
		Message       string      `json:"message"`                  // Descriptive message of the event
		FlowID        string      `json:"flow_id"`                  // ID of the flow
		OldStatus     string      `json:"old_status,omitempty"`     // Previous status of the flow (optional)
		NewStatus     string      `json:"new_status,omitempty"`     // New status of the flow (optional)
		ErrorRate     float64     `json:"error_rate,omitempty"`     // Overall error rate for the alert (optional)
		Threshold     int         `json:"threshold,omitempty"`      // Alert threshold that was reached or recovered from
		AlertState    string      `json:"alert_state,omitempty"`    // Status of the alert, e.g., "ACTIVATED" or "DEACTIVATED"
		Errors        []ErrorInfo `json:"errors,omitempty"`         // List of errors describing the alert (optional)
		P50Latency    int         `json:"p50_latency,omitempty"`    // P50 latency of the endpoint requests (optional)
		P90Latency    int         `json:"p90_latency,omitempty"`    // P90 latency of the endpoint requests (optional)
		RequestsCount int         `json:"requests_count,omitempty"` // Number of requests used to calculate metric (optional)
		Availability  int         `json:"availability"`
	}

	ErrorInfo struct {
		ErrorType  string  `json:"error_type,omitempty"`
		ErrorRate  float64 `json:"error_rate,omitempty"`
		ErrorCount int64   `json:"error_count,omitempty"`
	}
)
