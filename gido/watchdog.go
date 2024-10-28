package gido

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

type TicketResult struct {
	Message                string
	IsTicketNumberExceeded bool
}

type WatchingTask struct {
	stopChan     chan struct{}
	targetNumber int
}

var (
	watchingTasks = make(map[string]WatchingTask)
	tasksMutex    sync.Mutex
)

func checkTicketStatus(i *discordgo.InteractionCreate, targetNumber int) TicketResult {
	waitInfo, err := fetchWaitInfo()
	if err != nil {
		log.Printf("Error getting wait info: %v", err)
		return TicketResult{
			Message:                fmt.Sprintf("<@%s> 無法獲取等待信息: %v", i.Member.User.ID, err),
			IsTicketNumberExceeded: false,
		}
	}

	// if the current ticket number is invalid, we can't calculate the difference
	if !waitInfo.validateCurrentTicketNumber() {
		return TicketResult{
			Message:                fmt.Sprintf("<@%s> 當前票號: ----，您的票號: %d，無法計算差距", i.Member.User.ID, targetNumber),
			IsTicketNumberExceeded: false,
		}
	}

	difference := targetNumber - waitInfo.CurrentTicketNumber
	if difference <= 0 {
		return TicketResult{
			Message:                fmt.Sprintf("<@%s> 您的票號 %d 已經到達或已經過號！", i.Member.User.ID, targetNumber),
			IsTicketNumberExceeded: true,
		}
	}
	return TicketResult{
		Message:                fmt.Sprintf("<@%s> 當前票號: %d，您的票號: %d，還差 %d 個號碼", i.Member.User.ID, waitInfo.CurrentTicketNumber, targetNumber, difference),
		IsTicketNumberExceeded: false,
	}
}

func watchTicketNumberRoutine(s *discordgo.Session, i *discordgo.InteractionCreate, targetNumber int, stopChan <-chan struct{}) {
	// First check immediately
	checkResult := checkTicketStatus(i, targetNumber)
	s.ChannelMessageSend(i.ChannelID, checkResult.Message)
	// If the ticket number is exceeded, return immediately
	if checkResult.IsTicketNumberExceeded {
		return
	}

	// Then check every minute
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			checkResult := checkTicketStatus(i, targetNumber)
			s.ChannelMessageSend(i.ChannelID, checkResult.Message)
			if checkResult.IsTicketNumberExceeded {
				return
			}

		case <-stopChan:
			return
		}
	}
}
