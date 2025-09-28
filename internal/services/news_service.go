package services

import (
	"fmt"
	"webdev-90-days/internal/models"

	"github.com/mmcdole/gofeed"
)

type NewsService struct {
	Feeds map[string]string
}

func NewNewsService() *NewsService {
	return &NewsService{
		Feeds: map[string]string{
			"https://www.coindesk.com/feed/": "coindesk",      // убраны лишние пробелы
			"https://cointelegraph.com/rss":  "cointelegraph", // убраны лишние пробелы
			"https://www.theblock.co/rss":    "theblock",      // убраны лишние пробелы
		},
	}
}

func (n *NewsService) FetchAllNews() ([]models.NewsItem, error) {
	store := make([]models.NewsItem, 0)

	for k, v := range n.Feeds {
		fp := gofeed.NewParser()
		feed, err := fp.ParseURL(k)
		if err != nil {
			return nil, fmt.Errorf("cannot parse feed from %s: %v", v, err)
		}
		for _, item := range feed.Items {
			newNewsItem := models.NewsItem{
				ID:          0, // Пока ставим 0, потом можно будет генерировать или брать из базы
				Title:       item.Title,
				Description: item.Description,
				Link:        item.Link,
				PublishedAt: item.Published, // обрати внимание: Published, а не PublishedAt
				Source:      v,              // v — это имя источника (например, "coindesk")
			}
			store = append(store, newNewsItem)
		}
	}

	return store, nil // Не забудь вернуть результат!
}
