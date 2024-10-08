package bot

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

const FALLBACK_CURRENT_TICKET_NUMBER = "----"

// WaitInfo represents the information about the current waiting status.
// It includes the current ticket number, which can be either an integer or a string,
// and the count of tickets that are currently waiting.
type WaitInfo struct {
	CurrentTicketNumber interface{} // 可以是 int 或 string
	WaitingTicketsCount int
}

// getWaitInfo retrieves the wait information for "吉哆火鍋百匯" from the specified URL.
// It constructs the URL using the current date in YYYYMMDD format and the current timestamp in milliseconds.
// The function sends an HTTP GET request to the constructed URL and returns the response body as a string.
// If any error occurs during the process, it returns an error.
//
// Returns:
//   - string: The response body from the HTTP GET request.
//   - error: An error if the HTTP request fails, the status code is not OK, or reading the response body fails.
func getWaitInfo() (string, error) {
	// 獲取當前日期，格式為 YYYYMMDD
	currentDate := time.Now().Format("20060102")

	// 獲取當前時間戳（毫秒）
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)

	// 構建 URL
	url := fmt.Sprintf("http://vpn.weshine.com.tw:8088/WaitInfoWeb/WaitInfo_GIDOHandler.ashx?act=WaitInfo&DEP_CODE=吉哆火鍋百匯&Kind=a1&date=%s&_=%d", currentDate, timestamp)

	// 發送 GET 請求
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	// 檢查狀態碼
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP request returned status code %d", resp.StatusCode)
	}

	// read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	return string(body), nil
}

// 修改 parseWaitInfo 函數
func parseWaitInfo(info string) (WaitInfo, error) {
	var output WaitInfo

	parts := strings.Split(info, "|")
	if len(parts) != 3 {
		return output, fmt.Errorf("invalid format: expected 3 parts, got %d", len(parts))
	}

	// 我們忽略第一個部分 (useless)
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
		return output, fmt.Errorf("failed to parse total waits: %v", err)
	}
	output.WaitingTicketsCount = waitingTicketsCount

	return output, nil
}

// 修改 formatWaitInfo 函數
func formatWaitInfo(info WaitInfo) string {
	return fmt.Sprintf("當前叫號: %v，總共等待組數: %d", info.CurrentTicketNumber, info.WaitingTicketsCount)
}

func getAndParseWaitInfo() (WaitInfo, error) {
	info, err := getWaitInfo()
	if err != nil {
		return WaitInfo{}, err
	}

	return parseWaitInfo(info)
}

func cleanBotMessages(s *discordgo.Session, channelID string) (int, error) {
	var deletedCount int
	var lastMessageID string
	for {
		messages, err := s.ChannelMessages(channelID, 100, lastMessageID, "", "")
		if err != nil {
			return deletedCount, fmt.Errorf("獲取訊息失敗: %v", err)
		}

		if len(messages) == 0 {
			break
		}

		var botMessages []string
		for _, msg := range messages {
			if msg.Author.ID == BotID {
				botMessages = append(botMessages, msg.ID)
			}
			lastMessageID = msg.ID
		}

		if len(botMessages) > 0 {
			err = s.ChannelMessagesBulkDelete(channelID, botMessages)
			if err != nil {
				return deletedCount, fmt.Errorf("批量刪除訊息失敗: %v", err)
			}
			deletedCount += len(botMessages)
		}

		if len(messages) < 100 {
			break
		}

		// 為了避免超過 Discord API 的速率限制，在每次批量刪除後稍作暫停
		time.Sleep(1 * time.Second)
	}

	return deletedCount, nil
}
