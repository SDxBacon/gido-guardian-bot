package gido

import (
	"fmt"
	"strconv"
	"strings"
)

type WaitInfoIntField int

func (i WaitInfoIntField) String() string {
	intValue := int(i)
	if intValue < 0 {
		return "----"
	}
	return strconv.Itoa(intValue)
}

// WaitInfo represents the information about the current waiting status.
// It includes the current ticket number, which can be either an integer or a string,
// and the count of tickets that are currently waiting.
type WaitInfo struct {
	RawData       string
	CurrentNumber WaitInfoIntField
	TotalWaiting  WaitInfoIntField
}

func (info *WaitInfo) validateCurrentTicketNumber() bool {
	return info.CurrentNumber > 0
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
	currentNumber, err := strconv.Atoi(parts[1])
	if err != nil {
		output.CurrentNumber = -1
	} else {
		output.CurrentNumber = WaitInfoIntField(currentNumber)
	}

	// chunk 3, represent the count of waiting tickets
	totalWaitingCount, err := strconv.Atoi(parts[2])
	if err != nil {
		output.TotalWaiting = -1
	} else {
		output.TotalWaiting = WaitInfoIntField(totalWaitingCount)
	}

	return output, nil
}
