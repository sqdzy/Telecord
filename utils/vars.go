package utils

import (
	"github.com/go-telegram/bot"
	"sync"
	"time"
)

var ConnectionsMutex sync.Mutex
var TelegramBot *bot.Bot
var CurrentConfig Config
var SearchManager = NewSearchStateManager()
var Connections []Connection
var AuthenticationManager = &AuthManager{
	Expiry: 30 * time.Minute,
}

const MaxFileSize int64 = 10*1024*1024 - 1 // 10 мб
const ITEMS_PER_PAGE = 5
