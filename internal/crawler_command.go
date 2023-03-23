package internal

import (
	"net/url"
)

type CrawlerCommand struct {
	url url.URL
}

func NewCrawlerCommand(url url.URL) (CrawlerCommand, error) {
	return CrawlerCommand{
		url: url,
	}, nil
}
