# Telecord

**RU**:
Пересылка сообщений Discord в телеграмм Telegram и обратно (поддержка медиа-контента)

Golang, github.com/go-telegram/bot, github.com/bwmarrin/discordgo
Парсер торрентов с некоторыми сортировщиками (загрузки, сиды, размер, дата обновления) или без них

Параметры окружения:
	discordToken := os.Getenv(«DISCORD_TOKEN»)
	telegramToken := os.Getenv(«TELEGRAM_TOKEN»)
	discordAppID := os.Getenv(«DISCORD_APP_ID»)
	ruTrackerLogin := os.Getenv(«RUTRACKER_LOGIN»)
	ruTrackerPassword := os.Getenv(«RUTRACKER_PASSWORD»)

Аутентификация RuTracker сохраняется в течение 30 минут

Данные о подключении хранятся локально в файле data.js

Модули:
	github.com/PuerkitoBio/goquery v1.10.0 
	github.com/andybalholm/cascadia v1.3.2 
	github.com/bwmarrin/discordgo v0.28.1 
	github.com/go-telegram/bot v1.9.1 
	github.com/gorilla/websocket v1.4.2 
	golang.org/x/crypto v0.28.0 
	golang.org/x/net v0.30.0 
	golang.org/x/sys v0.26.0 
	golang.org/x/text v0.20.0 


**EN**:
Forwarding Discord messages to Telegram and vice versa (media content support)

Golang, github.com/go-telegram/bot, github.com/bwmarrin/discordgo
Torrent parser with or without some sorters (downloads, seeds, size, update date)

Enviroments vars:
	discordToken := os.Getenv("DISCORD_TOKEN")
	telegramToken := os.Getenv("TELEGRAM_TOKEN")
	discordAppID := os.Getenv("DISCORD_APP_ID")
	ruTrackerLogin := os.Getenv("RUTRACKER_LOGIN")
	ruTrackerPassword := os.Getenv("RUTRACKER_PASSWORD")

RuTracker auth saved for 30 mins

Connection data is stored locally in data.js

Modules:
	github.com/PuerkitoBio/goquery v1.10.0 
	github.com/andybalholm/cascadia v1.3.2 
	github.com/bwmarrin/discordgo v0.28.1 
	github.com/go-telegram/bot v1.9.1 
	github.com/gorilla/websocket v1.4.2 
	golang.org/x/crypto v0.28.0 
	golang.org/x/net v0.30.0 
	golang.org/x/sys v0.26.0 
	golang.org/x/text v0.20.0 
