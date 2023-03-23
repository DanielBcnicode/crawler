package internal

import (
	"golang.org/x/net/html"
	"net/url"
	"strings"
)

const (
	ROOT = "ROOT"
)

type Crawler interface {
	Run(command CrawlerCommand) (map[string]int, string, error)
}

type HttpCrawler struct {
	Extractor HtmlContentExtractor
}

func NewHttpCrawler(extractor HtmlContentExtractor) *HttpCrawler { // error is not necessary by now
	return &HttpCrawler{
		Extractor: extractor,
	}
}

// Run the crawler in the CrawlerCommand url, returns a map[string]int with the url crawled,
// the realUrl visited (can be redirected) and error
func (c *HttpCrawler) Run(command CrawlerCommand) (map[string]int, string, error) {
	returnData := map[string]int{}
	realURL, body, err := c.Extractor.Run(command.url.String())
	if err != nil {
		return nil, "", err
	}
	//Tokenizer
	tokenizer := html.NewTokenizer(body)
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
