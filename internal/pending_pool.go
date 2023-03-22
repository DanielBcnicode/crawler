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
This file implements a pull of url to be crawled,

*/

var ErrorRootNotFound = errors.New("root visited not found")

type HtmlCrawlingPendingAddress struct {
	Link     string
	Ancestor string
}
type PoolItem struct {
	Link      string
	Ancestor  string
	Children  []string
	Visited   bool
	VisitData time.Time
}
type BasicPendingPoolCrawler interface {
	Execute(wg *sync.WaitGroup, url string, ancestor string)
}
type htmlCrawlingPendingPull struct {
	sync.Mutex
	PendingChannel     chan HtmlCrawlingPendingAddress
	StopPendingChannel chan int
	StopChannel        chan int
	ProcessChannel     chan PoolItem
	WaitGroup          *sync.WaitGroup
	BaseHost           string
	pending            []PoolItem
	visited            map[string]PoolItem
	crawler            Crawler
}

func NewHtmlCrawler(
	Crawler Crawler,
	PendingChannel chan HtmlCrawlingPendingAddress,
	StopChannel chan int,
	WaitGroup *sync.WaitGroup,
	BaseHost string,
) BasicPendingPoolCrawler {
	return &htmlCrawlingPendingPull{
		crawler:            Crawler,
		PendingChannel:     PendingChannel,
		StopChannel:        StopChannel,
		WaitGroup:          WaitGroup,
		BaseHost:           BaseHost,
		pending:            []PoolItem{},
		visited:            map[string]PoolItem{},
		ProcessChannel:     make(chan PoolItem),
		StopPendingChannel: make(chan int),
	}
}
func (p *htmlCrawlingPendingPull) Execute(wg *sync.WaitGroup, url string, ancestor string) {
	lwg := sync.WaitGroup{}
	lwg.Add(2)
	go p.pendingHandler(&lwg)
	go p.processPending(&lwg)
	for i := 0; i < 300; i++ {
		// go routines to get pending url and process it
		go func() {
			for {
				select {
				case pc := <-p.ProcessChannel:
					p.Craw(pc.Link, pc.Ancestor, &lwg)
				}
			}
		}()
	}

	p.PendingChannel <- HtmlCrawlingPendingAddress{Link: url, Ancestor: ancestor}
	lwg.Wait()

	p.echoTreeResults()
	wg.Done()
}

func (p *htmlCrawlingPendingPull) processPending(wg *sync.WaitGroup) {
	var first PoolItem
	toNotify := false
	lastProcessedTime := time.Now()
	for {
		time.Sleep(10 * time.Millisecond)
		if p.lenPending() > 0 {
			p.Lock()
			first, p.pending = p.pending[0], p.pending[1:]
			//p.visited[first.Link] = first
			toNotify = true
			p.Unlock()
		}

		if toNotify {
			toNotify = false
			p.ProcessChannel <- first
			lastProcessedTime = time.Now()
		}

		if time.Now().Sub(lastProcessedTime) > time.Duration(5*time.Second) {
			p.StopPendingChannel <- 1
			wg.Done()
			return
		}
	}
}

func (p *htmlCrawlingPendingPull) pendingHandler(wg *sync.WaitGroup) {
	for {
		select {
		case _ = <-p.StopPendingChannel:
			wg.Done()
			return
		case pendingUrl, ok := <-p.PendingChannel:
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
			if u.Host != p.BaseHost {
				break
			}

			if p.isUrlProcessed(pendingUrl.Link) {
				break
			}

			p.Lock()
			p.pending = append(p.pending, PoolItem{
				Link:     pendingUrl.Link,
				Ancestor: pendingUrl.Ancestor,
				Visited:  false,
			})
			fmt.Println("Adding pending ", pendingUrl.Link)
			p.Unlock()
		}
	}
}

func (p *htmlCrawlingPendingPull) Craw(uri string, ancestor string, _ *sync.WaitGroup) {
	fmt.Println("Function: Craw() ", uri)
	u, err := url.Parse(uri)
	if err != nil { // Check error and log
		//wg.Done()
		return
	}

	cmd, err := NewCrawlerCommand(*u)
	if err != nil {
		fmt.Println(err)

	}

	urls, realUrl, err := p.crawler.Run(cmd)
	if err != nil {
		//we must to requeue
		pa := HtmlCrawlingPendingAddress{Link: uri, Ancestor: ancestor}
		p.PendingChannel <- pa
	}

	children := urlsWithSameDomain(realUrl, urls)
	//Mark url as visited
	pi := PoolItem{
		Link:      realUrl,
		Ancestor:  ancestor,
		Visited:   true,
		Children:  children,
		VisitData: time.Now(),
	}

	p.Lock()
	if ancestor == ROOT {
		ancestorUrl, _ := url.Parse(realUrl) //url confirmed previously, error never will success
		p.BaseHost = ancestorUrl.Host
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
		p.PendingChannel <- pa
	}

}

func (p *htmlCrawlingPendingPull) echoTreeResults() {
	root, err := p.rootFromVisited()
	if err != nil {
		fmt.Println("No results found")
		return
	}
	checkMap := map[string]bool{}
	p.echoResult(root.Link, 0, checkMap)
}

func (p *htmlCrawlingPendingPull) echoResult(uri string, indent int, checkMap map[string]bool) {
	current, ok := p.visited[uri]
	if !ok || checkMap[uri] == true {
		return
	}
	checkMap[uri] = true

	si := strings.Repeat("-", (indent)*2)
	fmt.Printf("%s %s :\n", si, current.Link)
	si = strings.Repeat("-", (indent+1)*2)
	for _, child := range current.Children {
		fmt.Printf("%s %s\n", si, child)
	}
	for _, child := range current.Children {
		p.echoResult(child, indent, checkMap)
	}
}

func (p *htmlCrawlingPendingPull) rootFromVisited() (PoolItem, error) {
	for _, item := range p.visited {
		if item.Ancestor == ROOT {
			return item, nil
		}
	}
	return PoolItem{}, ErrorRootNotFound
}

func (p *htmlCrawlingPendingPull) lenPending() int {
	p.Lock()
	defer p.Unlock()
	return len(p.pending)
}

func (p *htmlCrawlingPendingPull) isUrlProcessed(url string) bool {
	p.Lock()
	defer p.Unlock()

	for _, item := range p.pending {
		if item.Link == url {
			return true
		}
	}
	_, ok := p.visited[url]
	if ok {
		return true
	}

	return false
}

func urlsWithSameDomain(originalUrl string, urls map[string]int) []string {
	var childUrls []string
	if urls == nil {
		return childUrls
	}
	ou, err := url.Parse(originalUrl)
	if err != nil {
		return childUrls
	}
	for uri, _ := range urls {
		u, err := url.Parse(uri)
		if err != nil {
			continue
		}
		if ou.Host == u.Host {
			childUrls = append(childUrls, uri)
		}
	}
	return childUrls
}
