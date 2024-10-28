package gido

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// fetchWaitInfo retrieves the wait information for "吉哆火鍋百匯" from the specified URL.
// It constructs the URL using the current date in YYYYMMDD format and the current timestamp in milliseconds.
// The function sends an HTTP GET request to the constructed URL and returns the response body as a string.
// If any error occurs during the process, it returns an error.
//
// Returns:
//   - string: The response body from the HTTP GET request.
//   - error: An error if the HTTP request fails, the status code is not OK, or reading the response body fails.
func fetchWaitInfo() (WaitInfo, error) {
	// get current date in YYYYMMDD format
	currentDate := time.Now().Format("20060102")
	// get current timestamp in milliseconds
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	// construct the URL
	url := fmt.Sprintf("http://vpn.weshine.com.tw:8088/WaitInfoWeb/WaitInfo_GIDOHandler.ashx?act=WaitInfo&DEP_CODE=吉哆火鍋百匯&Kind=a1&date=%s&_=%d", currentDate, timestamp)

	// create an HTTP client with a timeout of 2 seconds
	client := &http.Client{
		Timeout: 2 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		return WaitInfo{}, fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	// check if the status code is OK
	if resp.StatusCode != http.StatusOK {
		return WaitInfo{}, fmt.Errorf("HTTP request returned status code %d", resp.StatusCode)
	}

	// read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return WaitInfo{}, fmt.Errorf("failed to read response body: %v", err)
	}

	// parse the response body to WaitInfo
	return parseWaitInfoFromResponse(string(body))
}
