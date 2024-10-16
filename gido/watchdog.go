package gido

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

type WatchingTask struct {
	stopChan     chan struct{}
	targetNumber int
}

var (
	watchingTasks = make(map[string]WatchingTask)
	tasksMutex    sync.Mutex
)

func checkTicketNumber(s *discordgo.Session, i *discordgo.InteractionCreate, targetNumber int) bool {
	rawWaitInfo, err := fetchWaitInfo()
	if err != nil {
		log.Printf("Error getting wait info: %v", err)
		return false
	}

	waitInfo, err := parseBody(rawWaitInfo)
	if err != nil {
		log.Printf("Error parsing wait info: %v", err)
		return false
	}

	switch currentNumber := waitInfo.CurrentTicketNumber.(type) {
	case int:
		difference := targetNumber - currentNumber
		if difference <= 0 {
			message := fmt.Sprintf("<@%s> 您的票號 %d 已經到達或已經過號！", i.Member.User.ID, targetNumber)
			s.ChannelMessageSend(i.ChannelID, message)
			return true
		}

		message := fmt.Sprintf("<@%s> 當前票號: %d，您的票號: %d，還差 %d 個號碼", i.Member.User.ID, currentNumber, targetNumber, difference)
		s.ChannelMessageSend(i.ChannelID, message)

	case string:
		message := fmt.Sprintf("<@%s> 當前票號: %s，您的票號: %d，無法計算差距", i.Member.User.ID, currentNumber, targetNumber)
		s.ChannelMessageSend(i.ChannelID, message)

	default:
		log.Printf("Unexpected type for CurrentTicketNumber: %T", waitInfo.CurrentTicketNumber)
	}

	return false
}

func watchTicketNumberRoutine(s *discordgo.Session, i *discordgo.InteractionCreate, targetNumber int, stopChan <-chan struct{}) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	// Immediately check once before entering the loop
	if checkTicketNumber(s, i, targetNumber) {
		return
	}

	for {
		select {
		case <-ticker.C:
			if checkTicketNumber(s, i, targetNumber) {
				return
			}

		case <-stopChan:
			return
		}
	}
}

func WatchTicketNumber(s *discordgo.Session, i *discordgo.InteractionCreate, number int) {
	tasksMutex.Lock()
	defer tasksMutex.Unlock()

	if task, exists := watchingTasks[i.Member.User.ID]; exists {
		close(task.stopChan)
		delete(watchingTasks, i.Member.User.ID)
	}

	stopChan := make(chan struct{})
	watchingTasks[i.Member.User.ID] = WatchingTask{stopChan: stopChan, targetNumber: number}

	go watchTicketNumberRoutine(s, i, number, stopChan)

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("已開始監視 Ticket: %d", number),
		},
	})
}

func StopWatchTicket(s *discordgo.Session, i *discordgo.InteractionCreate) {
	tasksMutex.Lock()
	defer tasksMutex.Unlock()

	if task, exists := watchingTasks[i.Member.User.ID]; exists {
		close(task.stopChan)
		delete(watchingTasks, i.Member.User.ID)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("<@%s> 已停止監視票號", i.Member.User.ID),
			},
		})
	} else {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("<@%s> 您目前沒有正在監視的票號", i.Member.User.ID),
			},
		})
	}
}
