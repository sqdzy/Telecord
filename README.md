# Telecord

![GitHub license](https://img.shields.io/github/license/sqdzy/Telecord)
![Go Version](https://img.shields.io/github/go-mod/go-version/sqdzy/Telecord)

## Описание (RU)

**Telecord** — это инструмент для пересылки сообщений между Discord и Telegram с поддержкой медиа-контента.  
Также включает парсер торрентов с возможностью сортировки по различным критериям (загрузки, сиды, размер, дата обновления).

### Технологии
- Язык программирования: **Golang**
- Библиотеки:
  - [go-telegram/bot](https://github.com/go-telegram/bot)
  - [discordgo](https://github.com/bwmarrin/discordgo)

### Переменные окружения
Для работы скрипта необходимо задать следующие переменные окружения:

- `discordToken := os.Getenv("DISCORD_TOKEN")`
- `telegramToken := os.Getenv("TELEGRAM_TOKEN")`
- `discordAppID := os.Getenv("DISCORD_APP_ID")`
- `ruTrackerLogin := os.Getenv("RUTRACKER_LOGIN")`
- `ruTrackerPassword := os.Getenv("RUTRACKER_PASSWORD")`

**Примечание:** Аутентификация на RuTracker сохраняется на 30 минут.

### Локальное хранилище
Все данные о подключении сохраняются локально в файле `data.js`.

### Зависимости
- [PuerkitoBio/goquery v1.10.0](https://github.com/PuerkitoBio/goquery)
- [andybalholm/cascadia v1.3.2](https://github.com/andybalholm/cascadia)
- [bwmarrin/discordgo v0.28.1](https://github.com/bwmarrin/discordgo)
- [go-telegram/bot v1.9.1](https://github.com/go-telegram/bot)
- [gorilla/websocket v1.4.2](https://github.com/gorilla/websocket)
- [golang.org/x/crypto v0.28.0](https://pkg.go.dev/golang.org/x/crypto)
- [golang.org/x/net v0.30.0](https://pkg.go.dev/golang.org/x/net)
- [golang.org/x/sys v0.26.0](https://pkg.go.dev/golang.org/x/sys)
- [golang.org/x/text v0.20.0](https://pkg.go.dev/golang.org/x/text)

---

## Description (EN)

**Telecord** is a tool for forwarding messages between Discord and Telegram with media content support.  
It also includes a torrent parser with optional sorting capabilities (downloads, seeds, size, update date).

### Technologies
- Programming Language: **Golang**
- Libraries:
  - [go-telegram/bot](https://github.com/go-telegram/bot)
  - [discordgo](https://github.com/bwmarrin/discordgo)

### Environment Variables
The following environment variables need to be set for the script to function:

- `discordToken := os.Getenv("DISCORD_TOKEN")`
- `telegramToken := os.Getenv("TELEGRAM_TOKEN")`
- `discordAppID := os.Getenv("DISCORD_APP_ID")`
- `ruTrackerLogin := os.Getenv("RUTRACKER_LOGIN")`
- `ruTrackerPassword := os.Getenv("RUTRACKER_PASSWORD")`

**Note:** RuTracker authentication is saved for 30 minutes.

### Local Storage
All connection data is stored locally in `data.js`.

### Dependencies
- [PuerkitoBio/goquery v1.10.0](https://github.com/PuerkitoBio/goquery)
- [andybalholm/cascadia v1.3.2](https://github.com/andybalholm/cascadia)
- [bwmarrin/discordgo v0.28.1](https://github.com/bwmarrin/discordgo)
- [go-telegram/bot v1.9.1](https://github.com/go-telegram/bot)
- [gorilla/websocket v1.4.2](https://github.com/gorilla/websocket)
- [golang.org/x/crypto v0.28.0](https://pkg.go.dev/golang.org/x/crypto)
- [golang.org/x/net v0.30.0](https://pkg.go.dev/golang.org/x/net)
- [golang.org/x/sys v0.26.0](https://pkg.go.dev/golang.org/x/sys)
- [golang.org/x/text v0.20.0](https://pkg.go.dev/golang.org/x/text)

---

## Установка и запуск

```bash
# Клонируйте репозиторий
git clone https://github.com/sqdzy/Telecord.git
```
```bash
# Перейдите в папку проекта
cd Telecord
```
```bash
# Установите зависимости
go mod tidy
```
```bash
# Установите переменные окружения

Windows:
Откройте командную строку и введите:

set DISCORD_TOKEN=your_discord_token
set TELEGRAM_TOKEN=your_telegram_token
set DISCORD_APP_ID=your_discord_app_id
set RUTRACKER_LOGIN=your_rutracker_login
set RUTRACKER_PASSWORD=your_rutracker_password

LINUX
Откройте командную строку и введите:

export DISCORD_TOKEN=your_discord_token
export TELEGRAM_TOKEN=your_telegram_token
export DISCORD_APP_ID=your_discord_app_id
export RUTRACKER_LOGIN=your_rutracker_login
export RUTRACKER_PASSWORD=your_rutracker_password
```
```bash
# Запустите приложение
go run main.go
```
