package discord

import (
	"FamilyObserver/utils"
	"bytes"
	"context"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

// TODO: Кто угодно, с любыми правами может выполнять команду
func ConnectDiscordHandler(s *discordgo.Session, i *discordgo.InteractionCreate, opts utils.OptionMap) {
	secretKey := opts["secretkey"].StringValue()

	utils.ConnectionsMutex.Lock()
	defer utils.ConnectionsMutex.Unlock()

	for j := range utils.Connections {
		if utils.Connections[j].SecretKey == secretKey {
			utils.Connections[j].DiscordGuild = i.GuildID
			utils.Connections[j].DiscordChannel = i.ChannelID
			utils.SaveConnectionsToFile()

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Соединение с Telegram-чатом установлено!",
				},
			})

			if utils.TelegramBot != nil {
				telegramMsg := fmt.Sprintf("Соединение с Discord установлено для гильдии: %s", i.GuildID)
				_, err := utils.TelegramBot.SendMessage(context.Background(), &bot.SendMessageParams{
					ChatID: utils.Connections[j].TelegramChat,
					Text:   telegramMsg,
				})
				if err != nil {
					log.Printf("Ошибка отправки сообщения в Telegram: %v", err)
				}
			} else {
				log.Println("TelegramBot не инициализирован")
			}
			return
		}
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Неверный секретный ключ!",
		},
	})
}

// Обработка /torrent Discord
func TorrentHandler(s *discordgo.Session, i *discordgo.InteractionCreate, opts utils.OptionMap) {
	torrentName := opts["name"].StringValue()
	sortType := "none"
	if opts["sorttype"] != nil {
		sortType = opts["sorttype"].StringValue()
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		log.Printf("Error sending deferred response: %v", err)
		return
	}

	client, authErr := utils.AuthenticationManager.Auth(utils.CurrentConfig.RuTrackerLogin, utils.CurrentConfig.RuTrackerPassword)
	if authErr != nil {
		handleError(s, i, fmt.Sprintf("Authentication error: %v", authErr))
		return
	}

	torrentList, err := utils.GetTorrent(*client, torrentName, sortType)
	if err != nil {
		handleError(s, i, fmt.Sprintf("Search error: %v", err))
		return
	}

	if len(torrentList) == 0 {
		handleError(s, i, "No results found for your search")
		return
	}

	if len(torrentList) == 0 {
		handleError(s, i, "No results found for your search")
		return
	}

	state := &utils.TorrentSearchState{
		TorrentList: torrentList,
		SearchQuery: torrentName,
		SortType:    sortType,
	}

	embed := createEmbed(torrentList, torrentName, sortType, 0)
	components := createNavigationComponents(0, len(torrentList), torrentList)

	msg, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds:     &[]*discordgo.MessageEmbed{embed},
		Components: &components,
	})
	if err != nil {
		log.Printf("Error editing response: %v", err)
		return
	}
	utils.SearchManager.StoreState(msg.ID, state)
}

func handleError(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	errorContent := fmt.Sprintf("❌ %s", message)
	_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &errorContent,
	})
	if err != nil {
		log.Printf("Error sending error message: %v", err)
	}
}

func createEmbed(torrentList []utils.Torrent, torrentName string, sortType string, page int) *discordgo.MessageEmbed {
	startIdx := page * utils.ITEMS_PER_PAGE
	endIdx := min((page+1)*utils.ITEMS_PER_PAGE, len(torrentList))

	embed := &discordgo.MessageEmbed{
		Title:  fmt.Sprintf("🔍 Результаты поиска: %s", torrentName),
		Color:  0x3498db,
		Fields: []*discordgo.MessageEmbedField{},
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Страница %d/%d | Сортировка: %s | Всего результатов: %d",
				page+1,
				(len(torrentList)-1)/utils.ITEMS_PER_PAGE+1,
				utils.FormatSortType(sortType),
				len(torrentList)),
		},
	}

	// Добавляем торренты для текущей страницы
	for i := startIdx; i < endIdx; i++ {
		torrent := torrentList[i]
		if torrent.Title == "" {
			continue
		}

		// Ограничиваем длину названия
		title := utils.TruncateString(torrent.Title, 250)

		fieldValue := fmt.Sprintf(
			"📥 Загрузок: **%d**\n"+
				"👥 Раздают: **%d**\n"+
				"🗂️ Размер: **%s**\n"+
				"👤 Создатель: **%s**"+
				"📅 Дата: **%s**\n"+
				"🔗 [Скачать](https://rutracker.org/forum/%s)\n",
			torrent.Downloads,
			torrent.Seeds,
			torrent.Size,
			torrent.Creator,
			torrent.Date,
			torrent.Href,
		)

		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("%d. %s", i+1, title),
			Value:  fieldValue,
			Inline: false,
		})
	}

	return embed
}

func createNavigationComponents(currentPage int, totalItems int, torrentList []utils.Torrent) []discordgo.MessageComponent {
	totalPages := (totalItems-1)/utils.ITEMS_PER_PAGE + 1
	startIdx := currentPage * utils.ITEMS_PER_PAGE
	endIdx := min((currentPage+1)*utils.ITEMS_PER_PAGE, len(torrentList))

	// Создаём ряд кнопок для навигации
	navRow := discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.Button{
				Label:    "◀️ Назад",
				CustomID: "prev_page",
				Style:    discordgo.PrimaryButton,
				Disabled: currentPage == 0,
			},
			discordgo.Button{
				Label:    "Вперед ▶️",
				CustomID: "next_page",
				Style:    discordgo.PrimaryButton,
				Disabled: currentPage >= totalPages-1,
			},
		},
	}

	// Создаём ряд кнопок для скачивания
	downloadRow := discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{},
	}

	// Добавляем кнопки скачивания для каждого торрента на текущей странице
	for i := startIdx; i < endIdx; i++ {
		// Корректируем номер кнопки в соответствии с индексом на текущей странице
		buttonLabel := fmt.Sprintf("Скачать #%d", i+1)
		downloadRow.Components = append(downloadRow.Components,
			discordgo.Button{
				Label:    buttonLabel,
				CustomID: fmt.Sprintf("download_%d", i),
				Style:    discordgo.SecondaryButton,
			},
		)
	}

	// Возвращаем оба ряда кнопок
	return []discordgo.MessageComponent{navRow, downloadRow}
}

func HandleDownloadRequest(s *discordgo.Session, i *discordgo.InteractionCreate, state *utils.TorrentSearchState, customID string) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		log.Printf("Error sending deferred response: %v", err)
		return
	}

	var index int
	fmt.Sscanf(customID, "download_%d", &index)

	if index >= len(state.TorrentList) {
		handleError(s, i, "Invalid torrent index")
		return
	}

	torrent := state.TorrentList[index]

	client, authErr := utils.AuthenticationManager.Auth(utils.CurrentConfig.RuTrackerLogin, utils.CurrentConfig.RuTrackerPassword)
	if authErr != nil {
		handleError(s, i, fmt.Sprintf("Authentication error: %v", authErr))
		return
	}

	torrentURL := "https://rutracker.org/forum/" + torrent.Href
	torrentData, title, err := utils.DownloadTorrent(*client, torrentURL)
	if err != nil {
		handleError(s, i, fmt.Sprintf("Download error: %v", err))
		return
	}

	// Create a file attachment for Discord
	reader := bytes.NewReader(torrentData)
	fileName := title + ".torrent"

	// Send the torrent file as an attachment
	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: utils.StringPtr(fmt.Sprintf("✅ Скачивание торрента: %s", title)),
		Files: []*discordgo.File{
			{
				Name:   fileName,
				Reader: reader,
			},
		},
	})
	if err != nil {
		log.Printf("Error sending file: %v", err)
		handleError(s, i, "Error sending the torrent file")
		return
	}
}

func ComponentHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionMessageComponent {
		return
	}

	data := i.MessageComponentData()
	state, exists := utils.SearchManager.GetState(i.Message.ID)
	if !exists {
		handleError(s, i, "Search session expired. Please try a new search.")
		return
	}

	if strings.HasPrefix(data.CustomID, "download_") {
		HandleDownloadRequest(s, i, state, data.CustomID)
		return
	}

	currentPage := 0
	if len(i.Message.Embeds) > 0 && i.Message.Embeds[0].Footer != nil {
		fmt.Sscanf(i.Message.Embeds[0].Footer.Text, "Страница %d/", &currentPage)
		currentPage--
	}

	newPage := currentPage
	if data.CustomID == "next_page" {
		newPage++
	} else if data.CustomID == "prev_page" {
		newPage--
	}

	embed := createEmbed(state.TorrentList, state.SearchQuery, state.SortType, newPage)
	components := createNavigationComponents(newPage, len(state.TorrentList), state.TorrentList)

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
		},
	})
	if err != nil {
		log.Printf("Error updating message: %v", err)
	}
}

// FIXME: - Сделать более UX отправку сообщения
//   - Обработка медиа с сообщением работает неправильно
func ForwardDiscordToTelegram(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot {
		return
	}
	utils.ConnectionsMutex.Lock()
	var linkedConnection *utils.Connection
	for _, connection := range utils.Connections {
		if connection.DiscordChannel == m.ChannelID {
			linkedConnection = &connection
			break
		}
	}
	utils.ConnectionsMutex.Unlock()

	if linkedConnection == nil {
		return
	}

	escapeMarkdown := func(text string) string {
		specialChars := []string{"_", "*", "`", "[", "]", "(", ")", "~", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}
		escaped := text
		for _, char := range specialChars {
			escaped = strings.ReplaceAll(escaped, char, "\\"+char)
		}
		return escaped
	}

	senderName := m.Author.Username
	if m.Member != nil && m.Member.Nick != "" {
		senderName = m.Member.Nick
	}
	senderName = escapeMarkdown(senderName)

	messageContent := escapeMarkdown(m.Content)
	messageText := fmt.Sprintf("%s: %s", senderName, messageContent)

	if m.Message.ReferencedMessage != nil {
		var replyContent string
		referenced, err := s.ChannelMessage(m.Message.ReferencedMessage.ChannelID, m.Message.ReferencedMessage.ID)
		if err == nil {
			replyAuthor := referenced.Author.Username
			if referenced.Member != nil && referenced.Member.Nick != "" {
				replyAuthor = referenced.Member.Nick
			}
			replyAuthor = escapeMarkdown(replyAuthor)
			if len(referenced.Attachments) != 0 {
				replyContent = "[Медиа]"
			} else {
				replyContent = escapeMarkdown(referenced.Content)
			}
			messageText = fmt.Sprintf("> %s: %s\n\n%s", replyAuthor, replyContent, messageText)
		}
	}

	params := &bot.SendMessageParams{
		ChatID:    linkedConnection.TelegramChat,
		Text:      messageText,
		ParseMode: models.ParseModeMarkdown,
	}

	if m.Content != "" || m.Message.ReferencedMessage != nil {
		_, err := utils.TelegramBot.SendMessage(context.Background(), params)
		if err != nil {
			log.Printf("Ошибка отправки сообщения в Telegram: %v", err)
		}
	}

	for _, attachment := range m.Attachments {
		resp, err := http.Get(attachment.URL)
		if err != nil {
			log.Printf("Ошибка получения файла: %v", err)
			continue
		}
		defer resp.Body.Close()

		caption := escapeMarkdown(fmt.Sprintf("%s", senderName))

		fileType := strings.ToLower(filepath.Ext(attachment.Filename))
		switch fileType {
		case ".gif":
			gifParams := &bot.SendAnimationParams{
				ChatID:  linkedConnection.TelegramChat,
				Caption: caption,
				Animation: &models.InputFileUpload{
					Filename: attachment.Filename,
					Data:     resp.Body,
				},
				ParseMode: models.ParseModeMarkdown,
			}
			_, err := utils.TelegramBot.SendAnimation(context.Background(), gifParams)
			if err != nil {
				log.Printf("Ошибка отправки гифки в Telegram: %v", err)
			}

		case ".jpg", ".jpeg", ".png":
			photoParams := &bot.SendPhotoParams{
				ChatID:  linkedConnection.TelegramChat,
				Caption: caption,
				Photo: &models.InputFileUpload{
					Filename: attachment.Filename,
					Data:     resp.Body,
				},
				ParseMode: models.ParseModeMarkdown,
			}
			_, err = utils.TelegramBot.SendPhoto(context.Background(), photoParams)

		case ".mp4", ".mov", ".webm":
			videoParams := &bot.SendVideoParams{
				ChatID:  linkedConnection.TelegramChat,
				Caption: caption,
				Video: &models.InputFileUpload{
					Filename: attachment.Filename,
					Data:     resp.Body,
				},
				ParseMode: models.ParseModeMarkdown,
			}
			_, err = utils.TelegramBot.SendVideo(context.Background(), videoParams)

		case ".mp3", ".ogg", ".wav":
			audioParams := &bot.SendAudioParams{
				ChatID:  linkedConnection.TelegramChat,
				Caption: caption,
				Audio: &models.InputFileUpload{
					Filename: attachment.Filename,
					Data:     resp.Body,
				},
				ParseMode: models.ParseModeMarkdown,
			}
			_, err = utils.TelegramBot.SendAudio(context.Background(), audioParams)

		default:
			docParams := &bot.SendDocumentParams{
				ChatID:  linkedConnection.TelegramChat,
				Caption: caption,
				Document: &models.InputFileUpload{
					Filename: attachment.Filename,
					Data:     resp.Body,
				},
				ParseMode: models.ParseModeMarkdown,
			}
			_, err = utils.TelegramBot.SendDocument(context.Background(), docParams)
		}

		if err != nil {
			log.Printf("Ошибка отправки файла в Telegram: %v", err)
		}
	}

}
