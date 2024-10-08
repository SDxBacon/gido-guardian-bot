package bot

import (
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
)

func watchTicketNumber(s *discordgo.Session, i *discordgo.InteractionCreate, targetNumber int, stopChan chan bool) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			waitInfo, err := getAndParseWaitInfo()
			if err != nil {
				log.Printf("Error getting wait info: %v", err)
				continue
			}

			var message string
			switch currentNumber := waitInfo.CurrentTicketNumber.(type) {
			case int:
				difference := targetNumber - currentNumber
				if difference <= 0 {
					message = fmt.Sprintf("<@%s> 您的票號 %d 已經到達或已經過號！", i.Member.User.ID, targetNumber)
					s.ChannelMessageSend(i.ChannelID, message)
					delete(watchingTasks, i.Member.User.ID)
					return
				}
				message = fmt.Sprintf("<@%s> 當前票號: %d，您的票號: %d，還差 %d 個號碼", i.Member.User.ID, currentNumber, targetNumber, difference)
			case string:
				message = fmt.Sprintf("<@%s> 當前票號: %s，您的票號: %d，無法計算差距", i.Member.User.ID, currentNumber, targetNumber)
			default:
				log.Printf("Unexpected type for CurrentTicketNumber: %T", waitInfo.CurrentTicketNumber)
				continue
			}

			s.ChannelMessageSend(i.ChannelID, message)

		case <-stopChan:
			return
		}
	}
}

func stopWatchTicket(s *discordgo.Session, i *discordgo.InteractionCreate) {
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
