package internal

import (
	"errors"
	"net/url"
)

type CrawlerCommand struct {
	url     url.URL
	maxDeep uint64
	deep    uint64
}

var (
	ErrorDeepTooHigh         = errors.New("the current deep is greater than the maximum allowed")
	ErrorMaxDeepCanNotBeZero = errors.New("the current deep is greater than the maximum allowed")
)

func NewCrawlerCommand(url url.URL, maxDeep uint64, deep uint64) (CrawlerCommand, error) {
	if maxDeep == 0 {
		return CrawlerCommand{}, ErrorMaxDeepCanNotBeZero
	}

	if deep > maxDeep {
		return CrawlerCommand{}, ErrorDeepTooHigh
	}

	return CrawlerCommand{
		url:     url,
		maxDeep: maxDeep,
		deep:    deep,
	}, nil
}
