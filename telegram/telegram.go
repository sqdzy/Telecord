package telegram

import (
	"FamilyObserver/utils"
	"context"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"log"
	"path/filepath"
	"strings"
)

func DefaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "/connect - получить секретный ключ для подключения к Discord",
	})
}

// FIXME: - Проблемы с обработкой аватаров некоторых пользователей
//   - Проблемы с отправкой видео (только видео, не GIF) от некоторых пользователей (при этом видео загружается по ссылке, но в Discord не воспроизодится)
//   - NOTE: Может быть добавить задержку рутиной на загрузку видео в вебхук
func ForwardTelegramToDiscord(ctx context.Context, b *bot.Bot, update *models.Update) {
	utils.ConnectionsMutex.Lock()
	var linkedConnection *utils.Connection
	for _, connection := range utils.Connections {
		if connection.TelegramChat == update.Message.Chat.ID {
			linkedConnection = &connection
			break
		}
	}
	utils.ConnectionsMutex.Unlock()

	if linkedConnection == nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Чат не привязан к Discord. Используйте команду /connect для привязки.",
		})
		return
	}

	discordSession, err := discordgo.New("Bot " + utils.CurrentConfig.DiscordToken)
	if err != nil {
		log.Printf("Ошибка создания сессии Discord: %v", err)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Не удалось установить соединение с Discord.",
		})
		return
	}
	defer discordSession.Close()

	webhooks, err := discordSession.ChannelWebhooks(linkedConnection.DiscordChannel)
	if err != nil {
		log.Printf("Ошибка получения вебхуков: %v", err)
		return
	}

	var webhook *discordgo.Webhook
	for _, wh := range webhooks {
		if wh.Name == "TelegramBridge" {
			webhook = wh
			break
		}
	}

	if webhook == nil {
		webhook, err = discordSession.WebhookCreate(linkedConnection.DiscordChannel, "TelegramBridge", "")
		if err != nil {
			log.Printf("Ошибка создания вебхука: %v", err)
			return
		}
	}

	telegramUser := update.Message.From
	username := telegramUser.FirstName
	if telegramUser.LastName != "" {
		username += " " + telegramUser.LastName
	}
	if telegramUser.Username != "" {
		username += " (@" + telegramUser.Username + ")"
	}

	var replyMessage string
	if update.Message.ReplyToMessage != nil {
		replyUser := update.Message.ReplyToMessage.From
		replyUsername := replyUser.FirstName
		if replyUser.LastName != "" {
			replyUsername += " " + replyUser.LastName
		}
		if replyUser.Username != "" {
			replyUsername += " (@" + replyUser.Username + ")"
		}

		replyText := update.Message.ReplyToMessage.Text
		if replyText == "" {
			if update.Message.ReplyToMessage.Caption != "" {
				replyText = update.Message.ReplyToMessage.Caption
			} else {
				replyText = "[медиа]"
			}
		}

		replyMessage = fmt.Sprintf("> **%s:** %s\n\n", replyUsername, replyText)
	}

	messageContent := update.Message.Text
	if replyMessage != "" {
		messageContent = replyMessage + messageContent
	}

	avatarURL := utils.GetTelegramAvatarURL(ctx, telegramUser)

	params := &discordgo.WebhookParams{
		Username: username,
		Content:  messageContent,
	}

	if avatarURL != "" {
		params.AvatarURL = avatarURL
	}
	if update.Message.Sticker != nil || update.Message.Photo != nil || update.Message.Document != nil || update.Message.Video != nil || update.Message.VideoNote != nil || update.Message.Voice != nil || update.Message.Poll != nil || update.Message.Location != nil {
		var fileURL string
		var err error
		var fileName string
		var fileSize int64
		var fileExt string

		switch {
		case update.Message.Sticker != nil:
			fileURL, err = utils.GetFileURL(ctx, b, update.Message.Sticker.FileID)
			fileSize = int64(update.Message.Sticker.FileSize)
			if strings.HasSuffix(fileURL, ".webm") {
				fileExt = ".webm"
			} else if strings.HasSuffix(fileURL, ".webp") {
				fileExt = ".webp"
			} else {
				fileExt = ".png"
			}
			fileName = fmt.Sprintf("sticker_%d%s", update.Message.Sticker.FileID, fileExt)

		case update.Message.Photo != nil:
			photo := update.Message.Photo[len(update.Message.Photo)-1]
			fileURL, err = utils.GetFileURL(ctx, b, photo.FileID)
			fileSize = int64(photo.FileSize)
			if strings.HasSuffix(fileURL, ".jpg") {
				fileExt = ".jpg"
			} else if strings.HasSuffix(fileURL, ".jpeg") {
				fileExt = ".jpeg"
			} else if strings.HasSuffix(fileURL, ".png") {
				fileExt = ".png"
			} else {
				fileExt = ".jpg"
			}
			fileName = fmt.Sprintf("photo_%d%s", photo.FileID, fileExt)

		case update.Message.VideoNote != nil:
			fileURL, err = utils.GetFileURL(ctx, b, update.Message.VideoNote.FileID)
			fileSize = int64(update.Message.VideoNote.FileSize)
			if strings.HasSuffix(fileURL, ".mp4") {
				fileExt = ".mp4"
			} else if strings.HasSuffix(fileURL, ".webm") {
				fileExt = ".webm"
			}
			fileName = fmt.Sprintf("video_note_%d%s", update.Message.VideoNote.FileID, fileExt)

		case update.Message.Voice != nil:
			fileURL, err = utils.GetFileURL(ctx, b, update.Message.Voice.FileID)
			fileSize = update.Message.Voice.FileSize
			fileExt = ".ogg"
			fileName = fmt.Sprintf("voice_%s%s", update.Message.Voice.FileID, fileExt)

			duration := update.Message.Voice.Duration
			if params.Content == "" {
				params.Content = fmt.Sprintf("🎤 Голосовое сообщение (%d сек.)", duration)
			} else {
				params.Content += fmt.Sprintf("\n🎤 Длительность: %d сек.", duration)
			}
		case update.Message.Video != nil:
			fileURL, err = utils.GetFileURL(ctx, b, update.Message.Video.FileID)
			fileSize = update.Message.Video.FileSize
			if strings.HasSuffix(fileURL, ".mp4") {
				fileExt = ".mp4"
			} else if strings.HasSuffix(fileURL, ".webm") {
				fileExt = ".webm"
			} else {
				fileExt = ".mp4"
			}
			fileName = fmt.Sprintf("video_%d%s", update.Message.Video.FileID, fileExt)
		case update.Message.Poll != nil:
			if params.Content == "" {
				params.Content = fmt.Sprintf("[Опрос на тему: %s]", update.Message.Poll.Question)
				_, err = discordSession.WebhookExecute(webhook.ID, webhook.Token, true, params)
				if err != nil {
					log.Printf("Ошибка отправки сообщения через вебхук: %v", err)
					return
				}
			}
			return
		case update.Message.Location != nil:
			if params.Content == "" {
				params.Content = "[Локация]"
				_, err = discordSession.WebhookExecute(webhook.ID, webhook.Token, true, params)
				if err != nil {
					log.Printf("Ошибка отправки сообщения через вебхук: %v", err)
					return
				}
			}
			return

		case update.Message.Document != nil: // Любой другой файл
			fileURL, err = utils.GetFileURL(ctx, b, update.Message.Document.FileID)
			fileSize = update.Message.Document.FileSize
			fileExt = filepath.Ext(update.Message.Document.FileName)
			fileName = update.Message.Document.FileName

		}
		if fileSize > utils.MaxFileSize {
			params.Content = fmt.Sprintf("[%s - %.2f Мб]", "Файл", float64(fileSize)/1024/1024)
		} else if err != nil {
			log.Printf("Ошибка получения данных файла: %v", err)
		} else {
			fileReader := utils.GetFileReader(fileURL)
			if fileReader == nil {
				log.Printf("Ошибка получения данных файла: %v", err)
				return
			}
			params.Files = []*discordgo.File{
				{
					Name:        fileName,
					Reader:      fileReader,
					ContentType: utils.GetContentType(filepath.Ext(fileName)),
				},
			}
		}

	}

	_, err = discordSession.WebhookExecute(webhook.ID, webhook.Token, true, params)
	if err != nil {
		log.Printf("Ошибка отправки сообщения через вебхук: %v", err)
		return
	}
}

// TODO: Кто угодно, с любыми правами может выполнять команду
func ConnectTelegramHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	secretKey := utils.GenerateSecretKey()

	for x := range utils.Connections {
		if utils.Connections[x].TelegramChat == update.Message.Chat.ID {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "Ваш чат уже прикреплен к какому-либо дискорд каналу"})
			return
		}
	}
	newConnection := utils.Connection{
		SecretKey:    secretKey,
		TelegramChat: update.Message.Chat.ID,
	}

	utils.ConnectionsMutex.Lock()
	utils.Connections = append(utils.Connections, newConnection)
	utils.SaveConnectionsToFile()
	utils.ConnectionsMutex.Unlock()

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      "Ваш секретный ключ для подключения к Discord: `" + secretKey + "`\nИспользуйте его в команде /connect в Discord.",
		ParseMode: models.ParseModeMarkdownV1,
	})
}

// TODO: Кто угодно, с любыми правами может выполнять команду
func UnconnectTelegramHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	for x := range utils.Connections {
		if utils.Connections[x].TelegramChat == update.Message.Chat.ID {
			utils.ConnectionsMutex.Lock()
			utils.Connections = append(utils.Connections[:x], utils.Connections[x+1:]...)
			utils.SaveConnectionsToFile()
			utils.ConnectionsMutex.Unlock()
		}
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      "Отвязали ваш чат от дискорд канала.",
		ParseMode: models.ParseModeMarkdownV1,
	})
}
