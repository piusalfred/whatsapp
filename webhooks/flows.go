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

package webhooks

import (
	"context"
	"fmt"
)

type (
	FlowNotificationContext struct {
		NotificationObject string // Corresponds to the 'object' field
		EntryID            string // Corresponds to the 'id' field in Entry
		EntryTime          int64  // Corresponds to the 'time' field in Entry
		ChangeField        string // Corresponds to the 'field' in Changes
		EventName          string // Corresponds to 'event' field in Value
		EventMessage       string // Corresponds to 'message' field in Value
		FlowID             string // Corresponds to 'flow_id' field in Value
	}

	FlowStatusHandler               EventHandler[FlowNotificationContext, StatusChangeDetails]
	FlowClientErrorRateHandler      EventHandler[FlowNotificationContext, ClientErrorRateDetails]
	FlowEndpointErrorRateHandler    EventHandler[FlowNotificationContext, EndpointErrorRateDetails]
	FlowEndpointLatencyHandler      EventHandler[FlowNotificationContext, EndpointLatencyDetails]
	FlowEndpointAvailabilityHandler EventHandler[FlowNotificationContext, EndpointAvailabilityDetails]
)

func (value *Value) FlowStatusChange() *StatusChangeDetails {
	return &StatusChangeDetails{
		OldStatus: value.OldStatus,
		NewStatus: value.NewStatus,
	}
}

func (value *Value) FlowClientErrorRate() *ClientErrorRateDetails {
	return &ClientErrorRateDetails{
		ErrorRate:  value.ErrorRate,
		Threshold:  value.Threshold,
		AlertState: value.AlertState,
		Errors:     value.Errors,
	}
}

func (value *Value) FlowEndpointErrorRate() *EndpointErrorRateDetails {
	return &EndpointErrorRateDetails{
		ErrorRate:  value.ErrorRate,
		Threshold:  value.Threshold,
		AlertState: value.AlertState,
		Errors:     value.Errors,
	}
}

func (value *Value) FlowEndpointLatency() *EndpointLatencyDetails {
	return &EndpointLatencyDetails{
		P50Latency:    value.P50Latency,
		P90Latency:    value.P90Latency,
		RequestsCount: value.RequestsCount,
		Threshold:     value.Threshold,
		AlertState:    value.AlertState,
	}
}

func (value *Value) FlowEndpointAvailability() *EndpointAvailabilityDetails {
	return &EndpointAvailabilityDetails{
		Availability: value.Availability,
		Threshold:    value.Threshold,
		AlertState:   value.AlertState,
	}
}

func (handler *Handler) handleFlowNotification(
	ctx context.Context,
	notificationContext *FlowNotificationContext,
	value *Value,
) error {
	switch value.Event {
	case EventFlowStatusChange:
		details := value.FlowStatusChange()
		if err := handler.flowStatus.HandleEvent(ctx, notificationContext, details); err != nil {
			return fmt.Errorf("handle flow status change event: %w", err)
		}

	case EventClientErrorRate:
		details := value.FlowClientErrorRate()
		if err := handler.flowClientErrorRate.HandleEvent(ctx, notificationContext, details); err != nil {
			return fmt.Errorf("handle client error rate event: %w", err)
		}

	case EventEndpointErrorRate:
		details := value.FlowEndpointErrorRate()
		if err := handler.flowEndpointErrorRate.HandleEvent(ctx, notificationContext, details); err != nil {
			return fmt.Errorf("handle endpoint error rate event: %w", err)
		}

	case EventEndpointLatency:
		details := value.FlowEndpointLatency()
		if err := handler.flowEndpointLatency.HandleEvent(ctx, notificationContext, details); err != nil {
			return fmt.Errorf("handle endpoint latency event: %w", err)
		}

	case EventEndpointAvailability:
		details := value.FlowEndpointAvailability()
		if err := handler.flowEndpointAvailability.HandleEvent(ctx, notificationContext, details); err != nil {
			return fmt.Errorf("handle endpoint availability event: %w", err)
		}
	}

	return nil
}

func (handler *Handler) OnFlowStatusChange(
	fn func(ctx context.Context, notificationContext *FlowNotificationContext, details *StatusChangeDetails) error,
) {
	handler.flowStatus = EventHandlerFunc[FlowNotificationContext, StatusChangeDetails](fn)
}

func (handler *Handler) SetFlowStatusChangeHandler(fn FlowStatusHandler) {
	handler.flowStatus = fn
}

func (handler *Handler) OnFlowClientErrorRate(
	fn func(ctx context.Context, notificationContext *FlowNotificationContext, details *ClientErrorRateDetails) error,
) {
	handler.flowClientErrorRate = EventHandlerFunc[FlowNotificationContext, ClientErrorRateDetails](fn)
}

func (handler *Handler) SetFlowClientErrorRateHandler(
	fn FlowClientErrorRateHandler,
) {
	handler.flowClientErrorRate = fn
}

func (handler *Handler) OnFlowEndpointErrorRate(
	fn func(ctx context.Context, notificationContext *FlowNotificationContext, details *EndpointErrorRateDetails) error,
) {
	handler.flowEndpointErrorRate = EventHandlerFunc[FlowNotificationContext, EndpointErrorRateDetails](fn)
}

func (handler *Handler) SetFlowEndpointErrorRateHandler(
	fn FlowEndpointErrorRateHandler,
) {
	handler.flowEndpointErrorRate = fn
}

func (handler *Handler) OnFlowEndpointLatency(
	fn func(ctx context.Context, notificationContext *FlowNotificationContext, details *EndpointLatencyDetails) error,
) {
	handler.flowEndpointLatency = EventHandlerFunc[FlowNotificationContext, EndpointLatencyDetails](fn)
}

func (handler *Handler) SetFlowEndpointLatencyHandler(
	fn FlowEndpointLatencyHandler,
) {
	handler.flowEndpointLatency = fn
}

func (handler *Handler) OnFlowEndpointAvailability(
	fn func(ctx context.Context, notificationContext *FlowNotificationContext, details *EndpointAvailabilityDetails) error,
) {
	handler.flowEndpointAvailability = EventHandlerFunc[FlowNotificationContext, EndpointAvailabilityDetails](fn)
}

func (handler *Handler) SetFlowEndpointAvailabilityHandler(
	fn FlowEndpointAvailabilityHandler,
) {
	handler.flowEndpointAvailability = fn
}
