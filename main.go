package main

import (
	"FamilyObserver/discord"
	"FamilyObserver/telegram"
	"FamilyObserver/utils"
	"context"
	"github.com/bwmarrin/discordgo"
	"github.com/go-telegram/bot"
	"log"
	"os"
	"os/signal"
	"sync"
)

func main() {
	if err := utils.LoadConfig(); err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	utils.LoadConnectionsFromFile()
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()

		opts := []bot.Option{
			bot.WithDefaultHandler(telegram.DefaultHandler),
		}

		b, err := bot.New(utils.CurrentConfig.TelegramToken, opts...)
		if err != nil {
			log.Fatalf("Error creating Telegram bot: %s", err)
		}

		utils.TelegramBot = b

		b.RegisterHandler(bot.HandlerTypeMessageText, "/connect", bot.MatchTypePrefix, telegram.ConnectTelegramHandler)
		b.RegisterHandler(bot.HandlerTypeMessageText, "/unconnect", bot.MatchTypePrefix, telegram.UnconnectTelegramHandler)
		b.RegisterHandler(bot.HandlerTypeMessageText, "", bot.MatchTypeContains, telegram.ForwardTelegramToDiscord)

		log.Println("Telegram bot started")
		b.Start(ctx)
	}()

	go func() {
		defer wg.Done()

		session, err := discordgo.New("Bot " + utils.CurrentConfig.DiscordToken)
		if err != nil {
			log.Fatalf("Error creating Discord session: %s", err)
		}

		session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if i.Type != discordgo.InteractionApplicationCommand {
				return
			}

			data := i.ApplicationCommandData()
			if data.Name == "connect" {
				discord.ConnectDiscordHandler(s, i, utils.ParseOptions(data.Options))
			}
		})

		session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if i.Type != discordgo.InteractionApplicationCommand {
				return
			}

			data := i.ApplicationCommandData()
			if data.Name == "torrent" {
				discord.TorrentHandler(s, i, utils.ParseOptions(data.Options))
			}
		})

		session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
			log.Printf("Logged in as %s", r.User.String())
		})

		session.AddHandler(discord.ForwardDiscordToTelegram)
		session.AddHandler(discord.ComponentHandler)

		_, err = session.ApplicationCommandBulkOverwrite(utils.CurrentConfig.DiscordAppID, "", utils.Commands)
		if err != nil {
			log.Fatalf("could not register commands: %s", err)
		}

		err = session.Open()
		if err != nil {
			log.Fatalf("could not open session: %s", err)
		}

		log.Println("Discord bot started")

		sigch := make(chan os.Signal, 1)
		signal.Notify(sigch, os.Interrupt)
		<-sigch

		err = session.Close()
		if err != nil {
			log.Printf("could not close session gracefully: %s", err)
		}
	}()

	wg.Wait()
	log.Println("Both bots have stopped")
}
