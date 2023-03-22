package internal

import (
	"context"
	"errors"
	"golang.org/x/net/html"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	ROOT = "ROOT"
)

var (
	ErrorInRequest = errors.New("error in the request")
)

type Crawler interface {
	Run(command CrawlerCommand) (map[string]int, string, error)
}

type HttpCrawler struct{}

func NewHttpCrawler() (*HttpCrawler, error) { // error is not necessary by now
	return &HttpCrawler{}, nil
}

// Run the crawler in the CrawlerCommand url, returns a map[string]int with the url crawled,
// the realUrl visited (can be redirected) and error
func (c *HttpCrawler) Run(command CrawlerCommand) (map[string]int, string, error) {
	returnData := map[string]int{}

	ctx, closeFunc := context.WithTimeout(context.Background(), 20*time.Second)
	defer closeFunc()
	req, err := http.NewRequestWithContext(ctx, "GET", command.url.String(), nil)
	if err != nil {
		return nil, "", nil
	}
	client := http.DefaultClient
	res, err := client.Do(req)
	if err != nil {
		println(err.Error())
		return nil, "", err
	}
	if res.StatusCode != http.StatusOK {
		return nil, "", err
	}
	realURL := res.Request.URL.String()
	defer func() { _ = res.Body.Close() }()

	//Tokenizer
	tokenizer := html.NewTokenizer(res.Body)
	for {
		tokenType := tokenizer.Next()
		switch {
		case tokenType == html.ErrorToken:
			if tokenizer.Err().Error() == "EOF" {
				return returnData, realURL, nil
			}

			return returnData, realURL, tokenizer.Err()
		case tokenType == html.StartTagToken:
			token := tokenizer.Token()
			if token.Data == "a" {
				for _, attribute := range token.Attr {
					if attribute.Key == "href" {
						u := attribute.Val
						if strings.HasPrefix(u, "#") {
							continue
						}
						if strings.HasPrefix(u, "/") {
							u, _ = url.JoinPath(realURL, u)
						}
						if !strings.HasPrefix(u, "http") {
							continue
						}
						returnData[u] = 0
					}
				}
			}
		}
	}
}
