package gido

import (
	"fmt"
	"strconv"
	"strings"
)

// WaitInfo represents the information about the current waiting status.
// It includes the current ticket number, which can be either an integer or a string,
// and the count of tickets that are currently waiting.
type WaitInfo struct {
	RawData             string
	CurrentTicketNumber int
	WaitingTicketsCount int
}

// convert WaitInfo to a message
func (info *WaitInfo) toMessage() string {
	currentTicket := strconv.Itoa(info.CurrentTicketNumber)
	waitingTickets := strconv.Itoa(info.WaitingTicketsCount)

	if info.CurrentTicketNumber == -1 {
		currentTicket = "----"
	}
	if info.WaitingTicketsCount == -1 {
		waitingTickets = "----"
	}

	return fmt.Sprintf("當前叫號: %v，總共等待組數: %v", currentTicket, waitingTickets)
}

func (info *WaitInfo) validateCurrentTicketNumber() bool {
	return info.CurrentTicketNumber > 0
}

// parse the wait information from the API response body
func parseWaitInfoFromResponse(info string) (WaitInfo, error) {
	var output WaitInfo = WaitInfo{RawData: info}

	parts := strings.Split(info, "|")
	if len(parts) != 3 {
		return output, fmt.Errorf("invalid format: expected 3 parts, got %d", len(parts))
	}

	// skip the first part, which is unless data

	// chunk 2, represene the current ticket number
	currentTicketNumber, err := strconv.Atoi(parts[1])
	if err != nil {
		output.CurrentTicketNumber = -1
	} else {
		output.CurrentTicketNumber = currentTicketNumber
	}

	// chunk 3, represent the count of waiting tickets
	waitingTicketsCount, err := strconv.Atoi(parts[2])
	if err != nil {
		output.WaitingTicketsCount = -1
	} else {
		output.WaitingTicketsCount = waitingTicketsCount
	}

	return output, nil
}
