package utils

import (
	"bytes"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/text/encoding/charmap"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strconv"
	"strings"
)

func Auth(login string, password string) (http.Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return http.Client{}, fmt.Errorf("Error creating cookie jar: %v", err)
	}

	client := &http.Client{
		Jar: jar,
	}

	// Данные для авторизации
	loginURL := "https://rutracker.org/forum/login.php"
	data := fmt.Sprintf("login_username=%s&login_password=%s&login=Вход", login, password)

	// Создаем запрос на авторизацию
	req, err := http.NewRequest("POST", loginURL, bytes.NewBuffer([]byte(data)))
	if err != nil {
		return http.Client{}, fmt.Errorf("Ошибка при создании запроса: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Выполняем запрос
	resp, err := client.Do(req)
	if err != nil {
		return http.Client{}, fmt.Errorf("Ошибка при выполнении запроса: %v", err)
	}
	defer resp.Body.Close()

	// Проверяем успешность авторизации
	if resp.StatusCode != http.StatusOK {
		return http.Client{}, fmt.Errorf("Не удалось авторизоваться, статус код: %d", resp.StatusCode)
	}
	return *client, nil
}

func GetTorrent(client http.Client, req string, sorttype ...string) ([]Torrent, error) {
	targetURL := fmt.Sprintf("https://rutracker.org/forum/tracker.php?nm=%s", req)
	if len(sorttype) > 0 {
		switch sorttype[0] {
		case "1":
			targetURL = fmt.Sprintf("https://rutracker.org/forum/tracker.php?nm=%s&o=4&s=2", req)
		case "2":
			targetURL = fmt.Sprintf("https://rutracker.org/forum/tracker.php?nm=%s&o=10&s=2", req)
		case "3":
			targetURL = fmt.Sprintf("https://rutracker.org/forum/tracker.php?nm=%s&o=1&s=2", req)
		case "4":
			targetURL = fmt.Sprintf("https://rutracker.org/forum/tracker.php?nm=%s&o=7&s=2", req)
		}
	}
	resp, err := client.Get(targetURL)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при запросе целевой страницы: %v", err)
	}
	defer resp.Body.Close()
	decoder := charmap.Windows1251.NewDecoder()

	// Читаем и декодируем содержимое
	utf8Reader := decoder.Reader(resp.Body)
	body, err := io.ReadAll(utf8Reader)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при чтении тела ответа: %v", err)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("Ошибка при парсинге страницы: %v", err)
	}

	torrentData := doc.Find("#search-results tbody tr.tCenter")
	torrentList := make([]Torrent, 10)
	index := 0

	torrentData.EachWithBreak(func(i int, s *goquery.Selection) bool {
		if index >= 10 {
			return false
		}

		// Получаем ссылку и название торрента
		titleElement := s.Find("td.t-title-col a.tt-text")
		href, isHrefExists := titleElement.Attr("href")
		if !isHrefExists {
			err = fmt.Errorf("не удалось получить ссылку на торрент")
			return false
		}
		title := titleElement.Text()

		// Получаем количество личей
		seedsStr := s.Find("b.seedmed").Text()
		seeds, leechErr := strconv.Atoi(strings.TrimSpace(seedsStr))
		if leechErr != nil {
			err = fmt.Errorf("не удалось получить количество личей: %v", leechErr)
			return false
		}
		torrentSize := s.Find("td.tor-size").Text()
		if torrentSize == "" {
			err = fmt.Errorf("не удалось получить размер торрента: %v", torrentSize)
			return false
		}
		torrentCreator := s.Find("td.u-name-col").Text()
		if torrentCreator == "" {
			err = fmt.Errorf("не удалось получить создателя торрента: %v", torrentCreator)
			return false
		}
		torrentDate := s.Find("td").Last().Text()
		if torrentDate == "" {
			err = fmt.Errorf("не удалось получить дату изменения торрента: %v", torrentDate)
			return false
		}
		// Получаем количество загрузок
		downloadsStr := s.Find("td.number-format").Text()
		downloads, downloadErr := strconv.Atoi(strings.TrimSpace(downloadsStr))
		if downloadErr != nil {
			err = fmt.Errorf("не удалось получить количество загрузок: %v", downloadErr)
			return false
		}

		torrentList[index] = Torrent{
			Title:     title,
			Href:      href,
			Downloads: int32(downloads),
			Seeds:     int16(seeds),
			Size:      torrentSize,
			Creator:   torrentCreator,
			Date:      torrentDate,
		}

		index++
		return true
	})

	if err != nil {
		return nil, err
	}
	return torrentList, nil
}

func DownloadTorrent(client http.Client, url string) ([]byte, string, error) {
	// Request the torrent page
	resp, err := client.Get(url)
	if err != nil {
		return nil, "", fmt.Errorf("error requesting torrent page: %v", err)
	}
	defer resp.Body.Close()

	// Read and decode the page content from Windows-1251 to UTF-8
	decoder := charmap.Windows1251.NewDecoder()
	utf8Reader := decoder.Reader(resp.Body)
	body, err := io.ReadAll(utf8Reader)
	if err != nil {
		return nil, "", fmt.Errorf("error reading page content: %v", err)
	}

	// Parse the HTML document
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil, "", fmt.Errorf("error parsing HTML: %v", err)
	}

	title := strings.TrimSpace(doc.Find("a#topic-title").Text())
	if title == "" {
		title = "download" // default filename if title not found
	}

	// Clean the title to make it suitable for a filename
	title = strings.Map(func(r rune) rune {
		if strings.ContainsRune(`<>:"/\|?*`, r) {
			return '_'
		}
		return r
	}, title)

	// Find the download link
	downloadLink := ""
	doc.Find("a.dl-stub.dl-link.dl-topic").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists && strings.Contains(href, "dl.php?t=") {
			downloadLink = "https://rutracker.org/forum/" + href
		}
	})

	if downloadLink == "" {
		return nil, "", fmt.Errorf("could not find torrent download link")
	}

	// Download the .torrent file
	torrentResp, err := client.Get(downloadLink)
	if err != nil {
		return nil, "", fmt.Errorf("error downloading torrent file: %v", err)
	}
	defer torrentResp.Body.Close()

	// Read the torrent file content
	torrentData, err := io.ReadAll(torrentResp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("error reading torrent file: %v", err)
	}

	return torrentData, title, nil
}
