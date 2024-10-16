package bot

import (
	"fmt"
	"log"

	"github.com/SDxBacon/gido-guardian-bot/gido"
	"github.com/bwmarrin/discordgo"
)

// handleWaitInfoInteraction handles the "WaitInfo" interaction command from Discord.
// It retrieves the current wait info message and responds to the interaction with this message.
func handleWaitInfoInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type == discordgo.InteractionApplicationCommand && i.ApplicationCommandData().Name != Commands["WaitInfo"] {
		return
	}

	// Get the current wait info message
	waitInfoMessage, err := gido.GetCurrentWaitInfoMessage()
	if err != nil {
		log.Printf("error: %v", err)
		return
	}
	// Respond to the interaction with the wait info message
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: waitInfoMessage,
		},
	})
	// Log any errors
	if err != nil {
		log.Printf("error: %v", err)
	}
}

// handleWatchingInteraction handles the "Watching" interaction command from Discord.
// It checks if the interaction type is an application command and if the command name matches "Watching".
// If the conditions are met, it retrieves the target number from the interaction options and starts
// watching the target number by calling gido.WatchTicketNumber in a separate goroutine.
func handleWatchingInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type == discordgo.InteractionApplicationCommand && i.ApplicationCommandData().Name != Commands["Watching"] {
		return
	}

	// Get the target number from the interaction
	options := i.ApplicationCommandData().Options
	number := int(options[0].IntValue())

	// Start watching the target number
	gido.WatchTicketNumber(s, i, number)
}

// handleStopWatchingInteraction handles the "StopWatching" interaction command from Discord.
// It checks if the interaction type is an application command and if the command name matches "StopWatching".
// If the conditions are met, it stops watching the target number by calling gido.StopWatchTicket.
func handleStopWatchingInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type == discordgo.InteractionApplicationCommand && i.ApplicationCommandData().Name != Commands["StopWatching"] {
		return
	}
	// Stop watching the target number
	gido.StopWatchTicket(s, i)
}

// handleCleanGidoInteraction handles the interaction for cleaning bot messages in a Discord channel.
// It responds to the interaction with a deferred message indicating that the cleaning process has started.
// Then, it attempts to delete bot messages in the specified channel and updates the interaction response
// with the result of the cleaning process.
func handleCleanGidoInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type == discordgo.InteractionApplicationCommand && i.ApplicationCommandData().Name != Commands["CleanGido"] {
		return
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "正在清理訊息...",
		},
	})
	if err != nil {
		log.Printf("Error responding to interaction: %v", err)
		return
	}

	deletedCount, err := cleanBotMessages(s, i.ChannelID)
	if err != nil {
		errMsg := fmt.Sprintf("清理訊息時發生錯誤: %v", err)
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &errMsg,
		})
		return
	}

	msg := fmt.Sprintf("已成功刪除 %d 條機器人訊息", deletedCount)
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &msg,
	})
}
