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
	WaitGroup      *sync.WaitGroup
	BaseHost       string
	pending        map[string]PoolItem
	visited        map[string]PoolItem
	sync.Mutex
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
	}
}
func (p *htmlCrawlingPendingPull) Execute() {
	wg := sync.WaitGroup{}
	wg.Add(1)
	go p.pendingHandler(&wg)
	for i := 0; i < 10; i++ {

	}
	wg.Wait()

	for {
		select {
		case <-p.StopChannel:
			//p.WaitGroup.Done()
			return
		case pendingUrl, ok := <-p.PendingChannel:
			if !ok {
				fmt.Println("Pichaaaaaa esta cerrado")

			}
			u, err := url.Parse(pendingUrl.Link)
			if err != nil {
				//log err
				fmt.Println(err)
				continue
			}
			if u.Host != p.BaseHost {
				//log err
				fmt.Printf("Host different!!! %s, must be %s\n", u.Host, p.BaseHost)
				continue
			}
			fmt.Printf("URL: %+v\n", pendingUrl)

			p.Lock()
			p.pending[pendingUrl.Link] = PoolItem{
				Link:     pendingUrl.Link,
				Ancestor: pendingUrl.Ancestor,
				Visited:  false,
			}
			p.Unlock()

			//wg.Add(1)
			//go p.Craw(pendingUrl.Link, &wg)
			//wg.Wait()
		}
	}
}

func (p *htmlCrawlingPendingPull) pendingHandler(wg *sync.WaitGroup) {
	for {
		select {
		case <-p.StopChannel:
			//p.WaitGroup.Done()
			return
		case _ = <-time.After(4 * time.Second): //Check if there are some pending
			if len(p.pending) == 0 {
				wg.Done()
				return
			}
		case pendingUrl, ok := <-p.PendingChannel:
			if !ok {
				fmt.Println("Pichaaaaaa esta cerrado")

			}
			u, err := url.Parse(pendingUrl.Link)
			if err != nil {
				//log err
				fmt.Println(err)
				continue
			}
			if u.Host != p.BaseHost {
				//log err
				fmt.Printf("Host different!!! %s, must be %s\n", u.Host, p.BaseHost)
				continue
			}
			fmt.Printf("URL: %+v\n", pendingUrl)

			p.Lock()
			p.pending[pendingUrl.Link] = PoolItem{
				Link:     pendingUrl.Link,
				Ancestor: pendingUrl.Ancestor,
				Visited:  false,
			}
			p.Unlock()
		}
	}
}

func (p *htmlCrawlingPendingPull) Craw(uri string, wg *sync.WaitGroup) {
	u, err := url.Parse(uri)
	if err != nil { // Check error and log
		wg.Done()
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
	for uu, _ := range urls.(map[string]int) {
		fmt.Printf("Try to enqueue %s\n", uu)
		//pa := HtmlCrawlingPendingAddress{Link: uu}
		//p.PendingChannel <- pa
	}
	wg.Done()

}
