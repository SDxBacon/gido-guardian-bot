package gido

const FALLBACK_CURRENT_TICKET_NUMBER = "----"

func GetCurrentWaitInfoMessage() (string, error) {
	rawInfo, err := fetchWaitInfo()
	if err != nil {
		return "", err
	}

	waitInfo, err := parseBody(rawInfo)
	if err != nil {
		return "", err
	}
	return toMessage(waitInfo), nil
}
