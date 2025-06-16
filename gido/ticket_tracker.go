package gido

import (
	"context"
	"strconv"
	"time"
)

type TicketTracker struct {
	ctx                        context.Context
	cancel                     context.CancelFunc
	trackingTicketId           int
	onStart                    func(ticketID int)
	onStop                     func(ticketID int)
	onFetchError               func(err error)
	onFetchInvalidTicketNumber func()
	onMonitorUpdate            func(currentNumber string, waitCount int)
	onTrackComplete            func()
}

type TicketTrackerOption func(*TicketTracker)

func WithTrackerOnStart(fn func(ticketID int)) TicketTrackerOption {
	return func(tt *TicketTracker) {
		tt.onStart = fn
	}
}

func WithTrackerOnStop(fn func(ticketID int)) TicketTrackerOption {
	return func(tt *TicketTracker) {
		tt.onStop = fn
	}
}

func WithTrackerOnFetchError(fn func(err error)) TicketTrackerOption {
	return func(tt *TicketTracker) {
		tt.onFetchError = fn
	}
}

func WithTrackerOnMonitorUpdate(fn func(currentNumber string, waitCount int)) TicketTrackerOption {
	return func(tt *TicketTracker) {
		tt.onMonitorUpdate = fn
	}
}

func WithTrackerOnFetchInvalidTicketNumber(fn func()) TicketTrackerOption {
	return func(tt *TicketTracker) {
		tt.onFetchInvalidTicketNumber = fn
	}
}

func WithTrackerOnTrackComplete(fn func()) TicketTrackerOption {
	return func(tt *TicketTracker) {
		tt.onTrackComplete = fn
	}
}

func NewTicketTracker(ticketID int, opts ...TicketTrackerOption) *TicketTracker {
	ctx, cancel := context.WithCancel(context.Background())

	tt := &TicketTracker{
		ctx:                        ctx,
		cancel:                     cancel,
		trackingTicketId:           ticketID,
		onStart:                    func(ticketID int) {},
		onStop:                     func(ticketID int) {},
		onFetchError:               func(err error) {},
		onFetchInvalidTicketNumber: func() {},
		onMonitorUpdate:            func(currentNumber string, waitCount int) {},
		onTrackComplete:            func() {},
	}

	// Apply options
	for _, opt := range opts {
		opt(tt)
	}
	return tt
}

func (tt *TicketTracker) Start() {
	go func() {
		// Ensure that the onStop is called when the goroutine exits
		defer func() {
			tt.onStop(tt.trackingTicketId)
		}()

		// if onStart is set, call it with the target ticket number
		tt.onStart(tt.trackingTicketId)

		for {
			select {
			case <-time.After(1 * time.Minute):
				// Fetch the current wait info
				currentWaitInfo, err := fetchWaitInfo()
				if err != nil {
					tt.onFetchError(err)
					continue
				}
				//
				if !currentWaitInfo.validateCurrentTicketNumber() {
					tt.onFetchInvalidTicketNumber()
					continue
				}

				currentTicketNumber := currentWaitInfo.CurrentTicketNumber

				// Calculate the wait count
				waitCount := tt.trackingTicketId - currentTicketNumber
				// If the wait count is greater than zero, it means the ticket is still waiting
				if waitCount > 0 {
					tt.onMonitorUpdate(
						strconv.Itoa(currentTicketNumber),
						waitCount,
					)
					continue
				}
				// If the wait count is less than or equal to zero, it means the ticket has been reached or exceeded
				tt.onTrackComplete()
				// Call the Stop method to terminate the tracking
				tt.Stop()

			case <-tt.ctx.Done():
				// Context is cancelled, exit the goroutine
				return
			}
		}
	}()
}

// Stop gracefully terminates the ticket tracking process.
// It cancels the context used by the tracker, which signals any running goroutines to exit.
func (tt *TicketTracker) Stop() {
	// Cancel the context to stop the goroutine
	tt.cancel()
}

func (tt *TicketTracker) GetTrackingTicketId() int {
	return tt.trackingTicketId
}
