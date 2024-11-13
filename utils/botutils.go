package utils

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"strings"
)

// Генерация случайного секретного ключа
func GenerateSecretKey() string {
	secretKey := make([]byte, 10)
	_, err := rand.Read(secretKey)
	if err != nil {
		log.Fatalf("Не удалось сгенерировать ключ: %v", err)
	}
	return hex.EncodeToString(secretKey)
}

// Ссылка на объект в памяти
func StringPtr(s string) *string {
	return &s
}

// Сокращать текст до определенного кол-ва символов
func TruncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	return s[:maxLength-3] + "..."
}

// Выбор типа сортировки
func FormatSortType(sortType string) string {
	switch sortType {
	case "1":
		return "по загрузкам"
	case "2":
		return "по количеству скачивающих"
	case "3":
		return "по дате изменения"
	case "4":
		return "по размеру"
	default:
		return "без сортировки"
	}
}

type OptionMap = map[string]*discordgo.ApplicationCommandInteractionDataOption

// Получение данных из команды в Дискорде в словарь
func ParseOptions(options []*discordgo.ApplicationCommandInteractionDataOption) (om OptionMap) {
	om = make(OptionMap)
	for _, opt := range options {
		om[opt.Name] = opt
	}
	return
}

// Список возможных команд и их параметров
var Commands = []*discordgo.ApplicationCommand{
	{
		Name:        "connect",
		Description: "Подключение к Telegram через секретный ключ",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "secretkey",
				Description: "Секретный ключ, выданный ботом в Telegram",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    true,
			},
			{
				Name:        "channel",
				Description: "Канал к которому будет привязан бот",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    true,
			},
		},
	},
	{
		Name:        "torrent",
		Description: "Поиск торрент-файла на RuTracker и RuTor с возможность фильтрации",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "name",
				Description: "Название торрента (фильма, файла)",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    true,
			},
			{
				Name:        "sorttype",
				Description: "Скачивания - 1, сиды - 2, дата изменения - 3, размер - 4",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    false,
			},
		},
	},
}

// Сохранение данных о соединении
func SaveConnectionsToFile() {
	marshal, err := json.MarshalIndent(Connections, "", "\t")
	if err != nil {
		log.Fatalf("Ошибка сериализации данных: %v", err)
	}
	err = os.WriteFile("data.json", marshal, 0644)
	if err != nil {
		log.Fatalf("Ошибка записи в файл: %v", err)
	}
}

// Загрузка данных о соединении
func LoadConnectionsFromFile() {
	if _, err := os.Stat("data.json"); os.IsNotExist(err) {
		emptyData := []byte("[]")
		err = os.WriteFile("data.json", emptyData, 0644)
		if err != nil {
			log.Fatalf("Ошибка создания файла data.json: %v", err)
		}
		log.Println("Файл data.json был создан, так как его не существовало.")
		return
	}

	data, err := os.ReadFile("data.json")
	if err != nil {
		log.Fatalf("Ошибка чтения файла: %v", err)
	}

	err = json.Unmarshal(data, &Connections)
	if err != nil {
		log.Fatalf("Ошибка десериализации данных: %v", err)
	}
}

// Получение Reader из URL
func GetFileReader(url string) io.Reader {
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Ошибка получения файла: %v", err)
		return nil
	}
	return resp.Body
}

// Получение прямого URL файла из Telegram
func GetFileURL(ctx context.Context, b *bot.Bot, fileID string) (string, error) {
	file, err := b.GetFile(ctx, &bot.GetFileParams{FileID: fileID})
	if err != nil {
		return "", fmt.Errorf("не удалось получить файл: %v", err)
	}

	fileURL := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", CurrentConfig.TelegramToken, file.FilePath)

	return fileURL, nil
}

// Получение ссылки на аватар пользователя Telegram (Первое фото/гиф в профиле)
func GetTelegramAvatarURL(ctx context.Context, user *models.User) string {
	userAvatars, err := TelegramBot.GetUserProfilePhotos(ctx, &bot.GetUserProfilePhotosParams{UserID: user.ID})
	if err != nil {
		log.Printf("Ошибка получения фото профиля: %v", err)
		return ""
	}

	if len(userAvatars.Photos) == 0 || len(userAvatars.Photos[0]) == 0 {
		log.Printf("У пользователя нет фото профиля")
		return ""
	}

	userAvatar, err := TelegramBot.GetFile(ctx, &bot.GetFileParams{FileID: userAvatars.Photos[0][len(userAvatars.Photos[0])-1].FileID})
	if err != nil {
		log.Printf("Ошибка получения файла аватара: %v", err)
		return ""
	}

	if userAvatar.FilePath == "" {
		log.Printf("Путь к файлу аватара пуст")
		return ""
	}

	return fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", CurrentConfig.TelegramToken, userAvatar.FilePath)
}

// Вспомогательная функция для определения Content-Type
func GetContentType(ext string) string {
	switch strings.ToLower(ext) {
	case ".ogg":
		return "audio/ogg"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	case ".webm":
		return "video/webm"
	case ".mp4":
		return "video/mp4"
	default:
		return getContentTypeForElse(ext)
	}
}

// Загрузка конфигурации из файла или создает новый файл с запросом токенов
func LoadConfig() error {
	discordToken := os.Getenv("DISCORD_TOKEN")
	telegramToken := os.Getenv("TELEGRAM_TOKEN")
	discordAppID := os.Getenv("DISCORD_APP_ID")
	ruTrackerLogin := os.Getenv("RUTRACKER_LOGIN")
	ruTrackerPassword := os.Getenv("RUTRACKER_PASSWORD")

	if discordToken == "" || telegramToken == "" || discordAppID == "" {
		return fmt.Errorf("необходимые переменные окружения не заданы")
	}

	// Записываем полученные значения в структуру конфигурации
	CurrentConfig = Config{
		DiscordToken:      discordToken,
		TelegramToken:     telegramToken,
		DiscordAppID:      discordAppID,
		RuTrackerPassword: ruTrackerPassword,
		RuTrackerLogin:    ruTrackerLogin,
	}

	log.Println("Конфигурация успешно загружена из переменных окружения")
	return nil
}

// Если GetContentType дошел до default
func getContentTypeForElse(ext string) string {
	mimeType := mime.TypeByExtension(ext)
	if mimeType != "" {
		mmt := strings.Split(mimeType, ";")
		if len(mmt) > 0 {
			log.Printf(mmt[0])
			return mmt[0]
		}
		return mimeType
	}
	return "application/octet-stream"
}

// Работа с состояниями сообщений Торрента (при переключении страниц)
func NewSearchStateManager() *SearchStateManager {
	return &SearchStateManager{
		states: make(map[string]*TorrentSearchState),
	}
}

func (sm *SearchStateManager) StoreState(messageID string, state *TorrentSearchState) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.states[messageID] = state
}

func (sm *SearchStateManager) GetState(messageID string) (*TorrentSearchState, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	state, exists := sm.states[messageID]
	return state, exists
}

func (sm *SearchStateManager) RemoveState(messageID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.states, messageID)
}
