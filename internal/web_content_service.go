package internal

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"
)

var (
	ErrorHTTP = errors.New("url not reachable")
)

type HtmlContentExtractor interface {
	Run(uri string) (realURL string, data io.Reader, err error)
}

type WebContentExtract struct {
}

func NewWebContentExtrat() *WebContentExtract {
	return &WebContentExtract{}
}

func (c *WebContentExtract) Run(uri string) (realURL string, data io.Reader, err error) {
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return "", nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()
	req = req.WithContext(ctx)
	client := http.DefaultClient
	res, err := client.Do(req)
	if err != nil {
		return "", nil, err
	}
	if res.StatusCode != http.StatusOK {
		return "", nil, ErrorHTTP
	}
	realURL = res.Request.URL.String()
	defer func() { _ = res.Body.Close() }()

	content, err := io.ReadAll(res.Body)

	return realURL, strings.NewReader(string(content)), err
}
