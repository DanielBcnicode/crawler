package internal

import (
	"errors"
	"net/url"
)

type CrawlerCommand struct {
	url url.URL
}

var (
	ErrorDeepTooHigh         = errors.New("the current deep is greater than the maximum allowed")
	ErrorMaxDeepCanNotBeZero = errors.New("the current deep is greater than the maximum allowed")
)

func NewCrawlerCommand(url url.URL) (CrawlerCommand, error) {
	return CrawlerCommand{
		url: url,
	}, nil
}
