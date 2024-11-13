package utils

import (
	"net/http"
	"sync"
	"time"
)

type Torrent struct {
	Title     string `json:"torrent_title"`
	Href      string `json:"torrent_href"`
	Downloads int32  `json:"torrent_downloads"`
	Seeds     int16  `json:"torrent_seeds"`
	Date      string `json:"torrent_date"`
	Size      string `json:"torrent_size"`
	Creator   string `json:"torrent_creator"`
}

type Connection struct {
	SecretKey      string `json:"secret_key"`
	TelegramChat   int64  `json:"telegram_chat"`
	DiscordGuild   string `json:"discord_guild"`
	DiscordChannel string `json:"discord_channel"`
}

type Config struct {
	DiscordToken      string `json:"discord_token"`
	TelegramToken     string `json:"telegram_token"`
	DiscordAppID      string `json:"discord_app_id"`
	RuTrackerPassword string `json:"ru_tracker_password"`
	RuTrackerLogin    string `json:"ru_tracker_login"`
}

type TorrentSearchState struct {
	TorrentList []Torrent
	SearchQuery string
	SortType    string
}

type SearchStateManager struct {
	mu     sync.RWMutex
	states map[string]*TorrentSearchState
}

type AuthManager struct {
	Client   *http.Client  // Клиент, который используется для запросов
	AuthTime time.Time     // Время последней авторизации
	Mu       sync.Mutex    // Для потокобезопасности
	Expiry   time.Duration // Время жизни сессии
}

func (a *AuthManager) Auth(username, password string) (*http.Client, error) {
	a.Mu.Lock()
	defer a.Mu.Unlock()

	// Проверяем, если сессия ещё активна
	if a.Client != nil && time.Since(a.AuthTime) < a.Expiry {
		return a.Client, nil
	}

	// Выполняем новую авторизацию
	client, err := Auth(username, password)
	if err != nil {
		return nil, err
	}

	// Сохраняем новый клиент и время авторизации
	a.Client = &client
	a.AuthTime = time.Now()

	return a.Client, nil
}
