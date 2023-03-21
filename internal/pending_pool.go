package internal

import (
	"fmt"
	"net/url"
	"sync"
	"time"
)

/*
This file implements a pull of url to be crawled,

*/

type HtmlCrawlingPendingAddress struct {
	Link     string
	Ancestor string
}
type PoolItem struct {
	Link      string
	Ancestor  string
	Visited   bool
	VisitData time.Time
}
type BasicPendingPoolCrawler interface {
	Execute()
}
type htmlCrawlingPendingPull struct {
	PendingChannel chan HtmlCrawlingPendingAddress
	StopChannel    chan int
	ProcessChannel chan PoolItem
	WaitGroup      *sync.WaitGroup
	BaseHost       string
	pending        []PoolItem
	visited        map[string]PoolItem
	sync.RWMutex
}

func NewHtmlCrawler(
	PendingChannel chan HtmlCrawlingPendingAddress,
	StopChannel chan int,
	WaitGroup *sync.WaitGroup,
	BaseHost string,
) BasicPendingPoolCrawler {
	return &htmlCrawlingPendingPull{
		PendingChannel: PendingChannel,
		StopChannel:    StopChannel,
		WaitGroup:      WaitGroup,
		BaseHost:       BaseHost,
		pending:        []PoolItem{},
		visited:        map[string]PoolItem{},
		ProcessChannel: make(chan PoolItem),
	}
}
func (p *htmlCrawlingPendingPull) Execute() {
	fmt.Println("Function: Execute()")
	wg := sync.WaitGroup{}
	wg.Add(2)
	go p.pendingHandler(&wg)
	go p.processPending(&wg)
	for i := 0; i < 30; i++ {
		// go routines to get pending url and process it
		go func() {
			for {
				select {
				case pc := <-p.ProcessChannel:
					p.Craw(pc.Link, &wg)
				}
			}
		}()
	}
	wg.Wait()

	for {
		select {
		case <-p.StopChannel:
			//p.WaitGroup.Done()
			return
		}
	}
}

func (p *htmlCrawlingPendingPull) lenPending() int {
	l := 0
	p.RLock()
	l = len(p.pending)
	p.RUnlock()
	return l
}

func (p *htmlCrawlingPendingPull) processPending(wg *sync.WaitGroup) {
	fmt.Println("Function: processPending()")
	var first PoolItem
	toNotify := false
	for {
		time.Sleep(10 * time.Millisecond)
		if p.lenPending() > 0 {
			p.Lock()
			first, p.pending = p.pending[0], p.pending[1:]
			p.visited[first.Link] = first
			toNotify = true
			p.Unlock()
		}

		if toNotify {
			p.ProcessChannel <- first
		}
	}
}

func (p *htmlCrawlingPendingPull) pendingHandler(wg *sync.WaitGroup) {
	fmt.Println("Function: pendingHandler()")
	for {
		select {
		case <-p.StopChannel:
			//p.WaitGroup.Done()
			return
		case _ = <-time.After(10 * time.Second): //Check if there are some pending
			if len(p.pending) == 0 {
				wg.Done()
				wg.Done()
				return
			}
		case pendingUrl, ok := <-p.PendingChannel:
			fmt.Println("Chan: PendingChannel")
			if !ok {
				fmt.Println("Error pending channel closed")
				wg.Done()
				return
			}
			u, err := url.Parse(pendingUrl.Link)
			if err != nil {
				//log err
				fmt.Println(err)
				break
			}
			if u.Host != p.BaseHost {
				//log err
				fmt.Printf("Host different!!! %s, must be %s\n", u.Host, p.BaseHost)
				break
			}

			//Check if the URL is queued or has been processed
			if p.isUrlProcessed(pendingUrl.Link) {
				break
			}
			fmt.Printf("URL: %+v\n", pendingUrl)
			p.Lock()
			//p.pending[pendingUrl.Link] = PoolItem{
			//	Link:     pendingUrl.Link,
			//	Ancestor: pendingUrl.Ancestor,
			//	Visited:  false,
			//}
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

func (p *htmlCrawlingPendingPull) isUrlProcessed(url string) bool {
	p.RLock()
	defer p.RUnlock()

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
func (p *htmlCrawlingPendingPull) Craw(uri string, _ *sync.WaitGroup) {
	fmt.Println("Function: Craw() ", uri)
	u, err := url.Parse(uri)
	if err != nil { // Check error and log
		//wg.Done()
		return
	}

	cmd, err := NewCrawlerCommand(*u, uint64(2), 1) // The 2 last parameters does not run
	if err != nil {
		fmt.Println(err)

	}
	serv, err := NewHttpCrawler()
	if err != nil {
		fmt.Println(err)

	}
	urls, err := serv.Run(cmd)
	//Add logit before send to pending channel
	urlsData, ok := urls.(map[string]int)
	if len(urlsData) == 0 || !ok {
		return
	}
	for uu, _ := range urls.(map[string]int) {
		if p.isUrlProcessed(uu) {
			continue
		}
		fmt.Printf("Try to enqueue %s\n", uu)
		pa := HtmlCrawlingPendingAddress{Link: uu, Ancestor: uri}
		p.PendingChannel <- pa
	}
	//wg.Done()

}
