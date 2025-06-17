package bot

import (
	"fmt"
	"log"

	"github.com/SDxBacon/gido-guardian-bot/gido"
	"github.com/SDxBacon/go-utils/discord/interaction"
	"github.com/bwmarrin/discordgo"
)

// handleWaitInfoInteraction handles the "WaitInfo" interaction command from Discord.
// It retrieves the current wait info message and responds to the interaction with this message.
func handleWaitInfoInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type == discordgo.InteractionApplicationCommand && i.ApplicationCommandData().Name != Commands["WaitInfo"] {
		return
	}

	// Create a new interaction responder
	responder := interaction.NewInteractionResponder(s, i.Interaction)

	// Get the current wait info struct
	waitInfo, err := gido.GetCurrentWaitInfo()
	if err != nil {
		responder.RespondWithError("Fail to GET wait info from GIDO", err)
		return
	}

	waitInfoMessage := fmt.Sprintf("當前叫號: %s，總共等待組數: %s", waitInfo.CurrentNumber.String(), waitInfo.TotalWaiting.String())

	err = responder.Respond(waitInfoMessage)
	if err != nil {
		log.Printf("error: %v", err)
	}
}

// handleWatchingInteraction processes Discord interactions for the "Watching" command
// which allows users to monitor a specified ticket number in a queue system.
//
// The function:
// 1. Validates that the interaction is an application command with the correct name
// 2. Extracts the user's ticket number from the command options
// 3. Creates a ticket tracker with appropriate event handlers that:
//   - Notifies when monitoring starts
//   - Notifies when monitoring stops
//   - Reports errors when fetching ticket information
//   - Handles cases where the ticket number is invalid
//   - Provides updates on the current ticket number and wait count
//   - Alerts the user when their ticket number is reached or passed
//
// Parameters:
//   - s: Discord session used for responding to the interaction
//   - i: The interaction data containing command information and user details
func handleWatchingInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type == discordgo.InteractionApplicationCommand && i.ApplicationCommandData().Name != Commands["Watching"] {
		return
	}

	// Get the target number from the interaction
	options := i.ApplicationCommandData().Options
	userTicketNumber := int(options[0].IntValue())

	// Create a new interaction responder
	responder := interaction.NewInteractionResponder(s, i.Interaction)

	// Create a ticket tracker instance
	ticketTracker, err := CreateUserTicketTracker(i.Member.User.ID, userTicketNumber,
		// Define the handlers for various events
		gido.WithTrackerOnStart(func(_ int) {
			responder.Respond(fmt.Sprintf("開始追蹤 Ticket: %d", userTicketNumber))
		}),
		gido.WithTrackerOnStop(func(_ int) {
			msg := fmt.Sprintf("<@%s> 已停止追蹤 Ticket: %d", i.Member.User.ID, userTicketNumber)
			s.ChannelMessageSend(i.ChannelID, msg)

			RemoveUserTicketTracker(i.Member.User.ID) // Remove the user ticket tracker when stopped
		}),
		gido.WithTrackerOnFetchError(func(err error) {
			msg := fmt.Sprintf("<@%s> 無法獲取 GIDO 伺服器回應: %v", i.Member.User.ID, err)
			s.ChannelMessageSend(i.ChannelID, msg)
		}),
		gido.WithTrackerOnFetchInvalidTicketNumber(func() {
			msg := fmt.Sprintf("<@%s> 當前票號: ----，您的票號: %d，無法計算差距", i.Member.User.ID, userTicketNumber)
			s.ChannelMessageSend(i.ChannelID, msg)
		}),
		gido.WithTrackerOnMonitorUpdate(func(currentNumber string, waitCount int) {
			msg := fmt.Sprintf("<@%s> 當前票號: %s，總共等待組數: %d", i.Member.User.ID, currentNumber, waitCount)
			s.ChannelMessageSend(i.ChannelID, msg)
		}),
		gido.WithTrackerOnTrackComplete(func() {
			msg := fmt.Sprintf("<@%s> 您的票號: %d 已經到達或已經過號！", i.Member.User.ID, userTicketNumber)
			s.ChannelMessageSend(i.ChannelID, msg)
		}))

	if err != nil {
		responder.Respond(fmt.Sprintf("無法創建 Ticket Tracker: %v", err))
		return
	}

	// start the ticket tracker
	ticketTracker.Start()
}

// handleStopWatchingInteraction handles the "StopWatching" interaction command from Discord.
// It checks if the interaction type is an application command and if the command name matches "StopWatching".
// If the conditions are met, it stops watching the target number by calling gido.StopWatchTicket.
func handleStopWatchingInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type == discordgo.InteractionApplicationCommand && i.ApplicationCommandData().Name != Commands["StopWatching"] {
		return
	}

	responder := interaction.NewInteractionResponder(s, i.Interaction)

	// Get the ticker tracker for the user
	tickerTracker := GetUserTicketTracker(i.Member.User.ID)
	if tickerTracker == nil {
		responder.Respond("您沒有正在追蹤的 Ticket")
		return
	}

	// Stop watching the target number
	responder.Respond(fmt.Sprintf("正在停止追蹤 Ticket: %d", tickerTracker.GetTrackingTicketId()))
	tickerTracker.Stop()
}

// handleCleanGidoInteraction handles the interaction for cleaning bot messages in a Discord channel.
// It responds to the interaction with a deferred message indicating that the cleaning process has started.
// Then, it attempts to delete bot messages in the specified channel and updates the interaction response
// with the result of the cleaning process.
func handleCleanGidoInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type == discordgo.InteractionApplicationCommand && i.ApplicationCommandData().Name != Commands["CleanGido"] {
		return
	}

	// Create a new interaction responder
	responder := interaction.NewInteractionResponder(s, i.Interaction)

	msg := "正在清理訊息..."
	err := responder.Respond(msg)
	if err != nil {
		log.Printf("Error responding to interaction: %v", err)
		return
	}

	deletedCount, err := cleanBotMessages(s, i.ChannelID)
	if err != nil {
		msg += fmt.Sprintf("\n❌ 清理訊息時發生錯誤: %v", err)
		responder.Respond(msg)
		return
	}

	msg += fmt.Sprintf("\n✅ 成功刪除 %d 條機器人訊息", deletedCount)
	responder.Respond(msg)
}
