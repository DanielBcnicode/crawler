package internal

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"
)

/*
This file implements a queue to process the crawling of one url [in the command]
This implementation only follows urls with the same subdomain.
It has one queue of links (type PoolItem) and a map[string]PoolItem with the visited
urls.
*/

// ErrorRootNotFound root link is not found in the visited map.
var ErrorRootNotFound = errors.New("root visited not found")

// HtmlCrawlingPendingAddress type used in the pending channel
type HtmlCrawlingPendingAddress struct {
	Link     string
	Ancestor string
}

// PoolItem holds the crawled url data
type PoolItem struct {
	link      string
	ancestor  string
	children  []string
	visited   bool
	visitDate time.Time
}

// CrawlerProcessor Interface of the main crawler processor
type CrawlerProcessor interface {
	Run(wg *sync.WaitGroup, url string, ancestor string)
	VisitedUrls() map[string]PoolItem //This function is only for testing purpose
	Response() interface{}
}

// htmlCrawlerProcessor Current crawler processor implementation
type htmlCrawlerProcessor struct {
	sync.Mutex
	pendingChannel     chan HtmlCrawlingPendingAddress
	stopPendingChannel chan int
	processChannel     chan PoolItem
	waitGroup          *sync.WaitGroup
	baseHost           string
	pending            []PoolItem
	visited            map[string]PoolItem
	crawler            Crawler
}

// NewHtmlCrawler constructor
func NewHtmlCrawler(
	Crawler Crawler,
	PendingChannel chan HtmlCrawlingPendingAddress,
	WaitGroup *sync.WaitGroup,
	BaseHost string,
) *htmlCrawlerProcessor {
	return &htmlCrawlerProcessor{
		crawler:            Crawler,
		pendingChannel:     PendingChannel,
		waitGroup:          WaitGroup,
		baseHost:           BaseHost,
		pending:            []PoolItem{},
		visited:            map[string]PoolItem{},
		processChannel:     make(chan PoolItem),
		stopPendingChannel: make(chan int),
	}
}

// Run main crawler processor function
func (p *htmlCrawlerProcessor) Run(wg *sync.WaitGroup, url string, ancestor string) {
	lwg := sync.WaitGroup{}
	lwg.Add(2)
	go p.processPendingItem(&lwg)
	go p.processPendingItemsQueue(&lwg)
	for i := 0; i < 300; i++ { // This number can be set as parameter
		go func() {
			for {
				select {
				case pc := <-p.processChannel:
					p.Crawl(pc.link, pc.ancestor, &lwg)
				}
			}
		}()
	}

	p.pendingChannel <- HtmlCrawlingPendingAddress{Link: url, Ancestor: ancestor}

	lwg.Wait()
	wg.Done()
}

// processPendingItemsQueue is a local implementation, it can be an external service
// and use a queue service as RabbitMQ, Kafka or others.
func (p *htmlCrawlerProcessor) processPendingItemsQueue(wg *sync.WaitGroup) {
	var first PoolItem
	toNotify := false
	lastProcessedTime := time.Now()
	for {
		time.Sleep(10 * time.Millisecond)
		if p.lenPending() > 0 {
			p.Lock()
			first, p.pending = p.pending[0], p.pending[1:]
			toNotify = true
			p.Unlock()
		}

		if toNotify {
			toNotify = false
			p.processChannel <- first
			lastProcessedTime = time.Now()
		}

		if time.Now().Sub(lastProcessedTime) > 5*time.Second {
			p.stopPendingChannel <- 1
			wg.Done()
			return
		}
	}
}

// processPendingItem Add pending url to the queue if the url is not processed
func (p *htmlCrawlerProcessor) processPendingItem(wg *sync.WaitGroup) {
	for {
		select {
		case _ = <-p.stopPendingChannel:
			wg.Done()
			return
		case pendingUrl, ok := <-p.pendingChannel:
			if !ok {
				fmt.Println("Error pending channel closed")
				wg.Done()
				return
			}
			u, err := url.Parse(pendingUrl.Link)
			if err != nil {
				fmt.Println(err)
				break
			}
			if u.Host != p.baseHost {
				break
			}

			if p.isUrlProcessed(pendingUrl.Link) {
				break
			}

			p.Lock()
			p.pending = append(p.pending, PoolItem{
				link:     pendingUrl.Link,
				ancestor: pendingUrl.Ancestor,
				visited:  false,
			})
			p.Unlock()
		}
	}
}

// Crawl makes the url crawling and add to visited map
func (p *htmlCrawlerProcessor) Crawl(uri string, ancestor string, _ *sync.WaitGroup) {
	fmt.Print(".")
	u, err := url.Parse(uri)
	if err != nil { // Check error and log
		return
	}

	cmd, err := NewCrawlerCommand(*u)
	if err != nil {
		fmt.Println(err)

	}

	urls, realUrl, err := p.crawler.Run(cmd)
	if err != nil {
		//TODO: implement a retry system
		//pa := HtmlCrawlingPendingAddress{link: uri, ancestor: ancestor}
		//p.pendingChannel <- pa
		return
	}

	children := urlsWithSameDomain(realUrl, urls)
	//Mark url as visited
	pi := PoolItem{
		link:      realUrl,
		ancestor:  ancestor,
		visited:   true,
		children:  children,
		visitDate: time.Now(),
	}

	p.Lock()
	if ancestor == ROOT { //HACK to prevent the url redirection only in the first url ROOT
		ancestorUrl, _ := url.Parse(realUrl) //url confirmed previously, error never will success
		p.baseHost = ancestorUrl.Host
	}
	p.visited[realUrl] = pi
	p.Unlock()

	if urls == nil || len(urls) == 0 {
		return
	}
	for _, uu := range children {
		if p.isUrlProcessed(uu) {
			continue
		}
		pa := HtmlCrawlingPendingAddress{Link: uu, Ancestor: realUrl}
		p.pendingChannel <- pa
	}

}

// Response prepare the crawler response, this version is a raw text
// it can have a parameter to indicate the response format.
func (p *htmlCrawlerProcessor) Response() interface{} {
	root, err := p.visitedRoot()
	if err != nil {
		return "No results found"
	}

	checkMap := map[string]bool{}
	dataResult := ""
	p.Lock()
	p.processResponse(root.link, 0, checkMap, &dataResult)
	p.Unlock()
	return dataResult
}

// processResponse is a recursive function to get data for the response
func (p *htmlCrawlerProcessor) processResponse(uri string, indent int, checkMap map[string]bool, dataResult *string) {
	current, ok := p.visited[uri]
	if !ok || checkMap[uri] == true {
		return
	}
	checkMap[uri] = true

	si := strings.Repeat(" ", (indent)*2)
	*dataResult += fmt.Sprintf("%s %s\n", si, current.link)
	si = strings.Repeat(" ", (indent+1)*2)

	// add the children urls
	for _, child := range current.children {
		*dataResult += fmt.Sprintf("%s %s\n", si, child)
	}
	// process the children
	for _, child := range current.children {
		p.processResponse(child, indent, checkMap, dataResult)
	}
}

func (p *htmlCrawlerProcessor) visitedRoot() (PoolItem, error) {
	p.Lock()
	defer p.Unlock()
	for _, item := range p.visited {
		if item.ancestor == ROOT {
			return item, nil
		}
	}
	return PoolItem{}, ErrorRootNotFound
}

func (p *htmlCrawlerProcessor) lenPending() int {
	p.Lock()
	defer p.Unlock()
	return len(p.pending)
}

func (p *htmlCrawlerProcessor) isUrlProcessed(url string) bool {
	p.Lock()
	defer p.Unlock()

	for _, item := range p.pending {
		if item.link == url {
			return true
		}
	}
	_, ok := p.visited[url]
	if ok {
		return true
	}
	return false
}

func (p *htmlCrawlerProcessor) VisitedUrls() map[string]PoolItem {
	return p.visited
}

// urlsWithSameDomain return a slice with the url with the same HOST in the URL
func urlsWithSameDomain(originalUrl string, urls map[string]int) []string {
	var siblingUrls []string
	if urls == nil {
		return siblingUrls
	}
	ou, err := url.Parse(originalUrl)
	if err != nil {
		return siblingUrls
	}
	for uri := range urls {
		u, err := url.Parse(uri)
		if err != nil {
			continue
		}
		if ou.Host == u.Host {
			siblingUrls = append(siblingUrls, uri)
		}
	}

	return siblingUrls
}
