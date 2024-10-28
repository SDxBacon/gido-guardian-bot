package bot

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	Commands = map[string]string{
		"WaitInfo":     "wait-info",
		"Watching":     "watching",
		"StopWatching": "stop-watching",
		"CleanGido":    "clean-gido",
	}
)

var (
	AppID    string = "1292493286681870377"
	Token    string = "YOUR_BOT_TOKEN_HERE"
	GuildID  string = ""
	BotID    string
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        Commands["WaitInfo"],
			Description: "Fetch the wait info of Gido",
		},
		{
			Name:        Commands["Watching"],
			Description: "Start watching for a specific ticket number",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "number",
					Description: "The ticket number to watch for",
					Required:    true,
				},
			},
		},
		{
			Name:        Commands["StopWatching"],
			Description: "Stop watching for ticket numbers",
		},
		{
			Name:        Commands["CleanGido"],
			Description: "Delete all messages sent by the bot in this channel",
		},
	}
)

func Run() {
	// create a session
	discord, err := discordgo.New("Bot " + Token)
	if err != nil {
		log.Fatal("Error message")
	}

	// add a event handler
	discord.AddHandler(onReady)
	discord.AddHandler(handleWaitInfoInteraction)
	discord.AddHandler(handleWatchingInteraction)
	discord.AddHandler(handleStopWatchingInteraction)
	discord.AddHandler(handleCleanGidoInteraction)

	// open session
	discord.Open()
	defer discord.Close() // close session, after function termination

	// keep bot running untill there is NO os interruption (ctrl + C)
	fmt.Println("Bot running....")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	fmt.Println("Bot stopped....")
}

func onReady(s *discordgo.Session, event *discordgo.Ready) {
	log.Printf("以 %v 身份登入", s.State.User.Username)
	BotID = s.State.User.ID

	// Get all existing commands from server
	existingCommands, err := s.ApplicationCommands(s.State.User.ID, GuildID)
	if err != nil {
		log.Panicf("Getting existing commands with err:%v", err)
		return
	}

	// Map of existing commands for quick lookup
	existingCommandMap := make(map[string]*discordgo.ApplicationCommand)
	for _, cmd := range existingCommands {
		existingCommandMap[cmd.Name] = cmd
	}

	// Register or update commands based on local `commands` list
	fmt.Printf("%s %s", time.Now().Format("2006/01/02 15:04:05"), "Registering Commands...")
	for _, v := range commands {
		existingCmd, exists := existingCommandMap[v.Name]
		if !exists || existingCmd.Description != v.Description {
			// Command does not exist or description has changed; create or update
			_, err := s.ApplicationCommandCreate(BotID, GuildID, v)
			if err != nil {
				log.Panicf("Unable to register or update '%v' command: %v", v.Name, err)
			}
		}
		// Remove the command from the map, so only extra commands remain
		delete(existingCommandMap, v.Name)
	}
	fmt.Printf("[\033[32mOK\033[0m]\n")

	// Delete extra commands that are not in the local `commands` list
	fmt.Printf("%s %s", time.Now().Format("2006/01/02 15:04:05"), "Deleting Legacy Commands...")
	for _, cmd := range existingCommandMap {
		err := s.ApplicationCommandDelete(BotID, GuildID, cmd.ID)
		if err != nil {
			log.Printf("Failed to delete extra command '%v': %v", cmd.Name, err)
		} else {
			log.Printf("Deleted extra command: %v", cmd.Name)
		}
	}
	fmt.Printf("[\033[32mOK\033[0m]\n")
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
