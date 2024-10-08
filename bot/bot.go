package bot

import (
	"fmt"
	"log"
	"os"
	"os/signal"

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
	watchingTasks = make(map[string]chan bool)
)

func Run() {

	// create a session
	discord, err := discordgo.New("Bot " + Token)
	if err != nil {
		log.Fatal("Error message")
	}

	// add a event handler
	discord.AddHandler(onReady)
	discord.AddHandler(interactionCreate)

	// open session
	discord.Open()
	defer discord.Close() // close session, after function termination

	// keep bot running untill there is NO os interruption (ctrl + C)
	fmt.Println("Bot running....")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	// delete all commands
	cmds, err := discord.ApplicationCommands(AppID, GuildID)
	if err != nil {
		log.Panicf("Cannot fetch commands: %v", err)
	} else {
		for _, v := range cmds {
			err := discord.ApplicationCommandDelete(discord.State.User.ID, GuildID, v.ID)
			if err != nil {
				log.Panicf("Cannot delete '%v' command: %v", v.Name, err)
			}
		}
		fmt.Println("Commands deleted....")
	}

	fmt.Println("Bot stopped....")
}

func onReady(s *discordgo.Session, event *discordgo.Ready) {
	log.Printf("以 %v 身份登入", s.State.User.Username)
	BotID = s.State.User.ID

	log.Println("正在註冊命令...")
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := s.ApplicationCommandCreate(BotID, GuildID, v)
		if err != nil {
			log.Panicf("無法創建 '%v' 命令: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}
}

func interactionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type == discordgo.InteractionApplicationCommand {
		switch i.ApplicationCommandData().Name {
		// case "wait-info":
		case Commands["WaitInfo"]:
			waitInfo, err := getAndParseWaitInfo()
			if err != nil {
				log.Printf("error: %v", err)
				return
			}

			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: formatWaitInfo(waitInfo),
				},
			})
			if err != nil {
				log.Printf("error: %v", err)
			}

		// case "watching":
		case Commands["Watching"]:
			options := i.ApplicationCommandData().Options
			number := int(options[0].IntValue())

			// Stop any existing watching task for this user
			if stopChan, exists := watchingTasks[i.Member.User.ID]; exists {
				close(stopChan)
			}

			stopChan := make(chan bool)
			watchingTasks[i.Member.User.ID] = stopChan

			go watchTicketNumber(s, i, number, stopChan)

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("開始監視票號 %d", number),
				},
			})

		// case "stop-watching":
		case Commands["StopWatching"]:
			stopWatchTicket(s, i)

		// case "clean-gido":
		case Commands["CleanGido"]:
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
	}
}
