package gido

func GetCurrentWaitInfoMessage() (string, error) {
	waitInfo, err := fetchWaitInfo()
	if err != nil {
		return "", err
	}

	return waitInfo.toMessage(), nil
}

// func StopWatchTicket(s *discordgo.Session, i *discordgo.InteractionCreate) {
// 	tasksMutex.Lock()
// 	defer tasksMutex.Unlock()

// 	if task, exists := watchingTasks[i.Member.User.ID]; exists {
// 		close(task.stopChan)
// 		delete(watchingTasks, i.Member.User.ID)
// 		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
// 			Type: discordgo.InteractionResponseChannelMessageWithSource,
// 			Data: &discordgo.InteractionResponseData{
// 				Content: fmt.Sprintf("<@%s> 已停止監視票號", i.Member.User.ID),
// 			},
// 		})
// 	} else {
// 		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
// 			Type: discordgo.InteractionResponseChannelMessageWithSource,
// 			Data: &discordgo.InteractionResponseData{
// 				Content: fmt.Sprintf("<@%s> 您目前沒有正在監視的票號", i.Member.User.ID),
// 			},
// 		})
// 	}
// }
