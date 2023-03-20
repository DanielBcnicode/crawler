package internal

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/net/html"
	"net/http"
	"strings"
	"time"
)

var (
	ErrorInRequest = errors.New("error in the request")
)

type Crawler interface {
	Run(command CrawlerCommand) (interface{}, error)
}

type HttpCrawler struct{}

func NewHttpCrawler() (*HttpCrawler, error) { // error is not necessary by now
	return &HttpCrawler{}, nil
}

func (c *HttpCrawler) Run(command CrawlerCommand) (interface{}, error) {
	var returnData map[string]int = map[string]int{}

	ctx, closeFunc := context.WithTimeout(context.Background(), 20*time.Second)
	defer closeFunc()
	req, err := http.NewRequestWithContext(ctx, "GET", command.url.String(), nil)
	if err != nil {
		return nil, nil
	}
	client := http.DefaultClient
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, err
	}
	realURL := res.Request.URL.String()
	println("Real url: " + realURL)
	defer func() { _ = res.Body.Close() }()
	//body, err := io.ReadAll(res.Body)

	//fmt.Print(string(body))

	//Tokenizer
	tokenizer := html.NewTokenizer(res.Body)
	for {
		tokenType := tokenizer.Next()
		switch {
		case tokenType == html.ErrorToken:
			fmt.Printf("Urls a Parsear : %d\n", len(returnData))
			return returnData, tokenizer.Err()
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
							fmt.Printf("Relative Anchor ..... %s\n", u)
							u = command.url.String() + u
						}
						if !strings.HasPrefix(u, "http") {
							continue
						}
						// Check if the url has been crawled
						returnData[u] = 0
						fmt.Printf("Anchor ..... %+v\n", u)
					}
				}
			}
		}

	}
}
