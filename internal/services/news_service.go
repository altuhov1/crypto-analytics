package services

import (
	"log/slog"
	"sort"
	"time"
	"webdev-90-days/internal/models"
	"webdev-90-days/internal/storage"

	"github.com/mmcdole/gofeed"
)

type NewsService struct {
	feeds        map[string]string
	store        storage.NewsStorage
	fetchEnabled bool
}

func NewNewsService(store storage.NewsStorage, fetchEnabled bool) *NewsService {
	service := &NewsService{
		feeds: map[string]string{
			"https://cointelegraph.com/rss":                   "cointelegraph",
			"https://www.coindesk.com/arc/outboundfeeds/rss/": "coindesk",
			"https://www.theblock.co/feed/rss":                "theblock",
		},
		store:        store,
		fetchEnabled: fetchEnabled,
	}

	// Запускаем фоновое обновление новостей
	go service.startBackgroundUpdates()

	return service
}

func (n *NewsService) startBackgroundUpdates() {
	// Сразу обновляем при старте
	n.updateNews()

	// Затем обновляем каждые 3 часа
	ticker := time.NewTicker(3 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		n.updateNews()
	}
}

// updateNews обновляет новости и сохраняет в хранилище
func (n *NewsService) updateNews() {
	if !n.fetchEnabled {
		return
	}
	slog.Info("Starting news update...")

	newsItems, err := n.fetchNewsFromFeeds()
	if err != nil {
		slog.Error("Error fetching news", "error", err)
		return
	}

	if err := n.store.UpdateNews(newsItems); err != nil {

		slog.Error("Error saving news", "error", err)
		return
	}

	slog.Info("Successfully updated news items",
		"amount", len(newsItems))
}

func (n *NewsService) fetchNewsFromFeeds() ([]models.NewsItem, error) {
	var allNews []models.NewsItem
	fp := gofeed.NewParser()

	for url, source := range n.feeds {
		feed, err := fp.ParseURL(url)
		if err != nil {
			slog.Warn("Warning: cannot parse feed from", "error", err)
			continue // Продолжаем с другими фидами при ошибке
		}

		for _, item := range feed.Items {
			newsItem := models.NewsItem{
				GUID:        item.GUID,
				Title:       item.Title,
				Description: item.Description,
				Link:        item.Link,
				PublishedAt: item.Published,
				Source:      source,
			}
			allNews = append(allNews, newsItem)
		}
	}

	return allNews, nil
}

func (n *NewsService) GetNews() ([]models.NewsItem, error) {
	news, err := n.store.GetAllNews()
	if err != nil {
		return nil, err
	}

	// Сортируем новости по дате публикации (сначала новые)
	sort.Slice(news, func(i, j int) bool {
		timeI := n.parseTimeWithFallback(news[i].PublishedAt)
		timeJ := n.parseTimeWithFallback(news[j].PublishedAt)

		return timeI.After(timeJ)
	})

	return news, nil
}

func (n *NewsService) parseTimeWithFallback(timeStr string) time.Time {
	if timeStr == "" {
		return time.Time{} // нулевое время (очень старая дата)
	}

	formats := []string{
		time.RFC1123,
		time.RFC1123Z,
		time.RFC822,
		time.RFC822Z,
		time.RFC3339,
		"Mon, 2 Jan 2006 15:04:05 MST",
		"Mon, 2 Jan 2006 15:04:05 -0700",
		"02 Jan 2006 15:04:05 MST",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return t
		}
	}

	return time.Time{}
}

func (n *NewsService) GetNewsCount() (int, error) {
	news, err := n.GetNews()
	if err != nil {
		return 0, err
	}
	return len(news), nil
}
