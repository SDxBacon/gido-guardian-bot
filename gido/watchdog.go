package gido

import (
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
)

var watchingTasks = make(map[string]chan bool)

func WatchTicketNumber(s *discordgo.Session, i *discordgo.InteractionCreate, targetNumber int, stopChan chan bool) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rawWaitInfo, err := fetchWaitInfo()
			if err != nil {
				log.Printf("Error getting wait info: %v", err)
				continue
			}

			waitInfo, err := parseBody(rawWaitInfo)
			if err != nil {
				log.Printf("Error parsing wait info: %v", err)
				continue
			}

			var message string
			var lastCurrentNumber int

			switch currentNumber := waitInfo.CurrentTicketNumber.(type) {
			case int:
				difference := targetNumber - currentNumber
				// If difference is less than or equal to 0, the ticket has been called or passed
				if difference <= 0 {
					message = fmt.Sprintf("<@%s> 您的票號 %d 已經到達或已經過號！", i.Member.User.ID, targetNumber)
					s.ChannelMessageSend(i.ChannelID, message)
					delete(watchingTasks, i.Member.User.ID)
					return
				}

				// create message
				message = fmt.Sprintf("<@%s> 當前票號: %d，您的票號: %d，還差 %d 個號碼", i.Member.User.ID, currentNumber, targetNumber, difference)

				// If difference is less than 3, increase the check frequency from 1 minute to 30 seconds
				if difference < 3 {
					ticker.Reset(30 * time.Second)
					log.Printf("Increased check frequency for user %s (target: %d)", i.Member.User.ID, targetNumber)
				}

				// If the current number hasn't changed, don't send the message
				if lastCurrentNumber != currentNumber {
					s.ChannelMessageSend(i.ChannelID, message)
				}

				// memorize the last current number
				lastCurrentNumber = currentNumber

			case string:
				message = fmt.Sprintf("<@%s> 當前票號: %s，您的票號: %d，無法計算差距", i.Member.User.ID, currentNumber, targetNumber)
				s.ChannelMessageSend(i.ChannelID, message)

			default:
				log.Printf("Unexpected type for CurrentTicketNumber: %T", waitInfo.CurrentTicketNumber)
				continue
			}

		case <-stopChan:
			return
		}
	}
}

func StopWatchTicket(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if stopChan, exists := watchingTasks[i.Member.User.ID]; exists {
		close(stopChan)
		delete(watchingTasks, i.Member.User.ID)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "已停止監視票號",
			},
		})
	} else {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "您目前沒有正在監視的票號",
			},
		})
	}
}
