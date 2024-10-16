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
	CurrentTicketNumber interface{} // 可以是 int 或 string
	WaitingTicketsCount interface{}
}

// convert WaitInfo to a message
func toMessage(info WaitInfo) string {
	return fmt.Sprintf("當前叫號: %v，總共等待組數: %v", info.CurrentTicketNumber, info.WaitingTicketsCount)
}

// parse the wait information from the API response body
func parseBody(info string) (WaitInfo, error) {
	var output WaitInfo

	parts := strings.Split(info, "|")
	if len(parts) != 3 {
		return output, fmt.Errorf("invalid format: expected 3 parts, got %d", len(parts))
	}

	// skip the first part, which is unless data

	// 第二部分是當前叫號
	CurrentTicketNumber, err := strconv.Atoi(parts[1])
	if err != nil {
		// 如果無法轉換為整數，保留為字符串
		output.CurrentTicketNumber = parts[1]
	} else {
		output.CurrentTicketNumber = CurrentTicketNumber
	}

	// 第三部分，解析總共等待組數
	waitingTicketsCount, err := strconv.Atoi(parts[2])
	if err != nil {
		// 如果無法轉換為整數，保留為字符串
		output.WaitingTicketsCount = parts[2]
	} else {
		output.WaitingTicketsCount = waitingTicketsCount
	}

	return output, nil
}
