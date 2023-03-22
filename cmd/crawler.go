package main

import (
	"crawler/internal"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"sync"
)

func main() {
	deep := flag.Int("d", 1, "-d crawler deep level")
	//args := flag.Args()

	flag.Parse()

	fmt.Println("deep: ", *deep)
	fmt.Println("Arguments number: ", flag.NArg())
	fmt.Println("Args: ", flag.Arg(0))

	uri, err := url.ParseRequestURI(flag.Arg(0))
	if err != nil {
		fmt.Println("the parameter specified is not a valid url: ", flag.Arg(0))
		os.Exit(1)
	}

	fmt.Println("Crawling the url: ", uri)
	fmt.Println("With protocol: ", uri.Scheme, uri.Opaque)

	// This must be the core main
	wg := sync.WaitGroup{}
	pendingChan := make(chan internal.HtmlCrawlingPendingAddress)
	stopChan := make(chan int)
	serv := internal.NewHttpCrawler(
		internal.NewWebContentExtrat(),
	)
	pendingService := internal.NewHtmlCrawler(serv, pendingChan, stopChan, &wg, uri.Host)

	wg.Add(1)

	go pendingService.Execute(&wg, uri.String(), internal.ROOT)

	cs := make(chan os.Signal, 1)
	signal.Notify(cs, os.Interrupt)
	go func() {
		<-cs
		log.Println("Goodbye")
		os.Exit(1)
	}()

	wg.Wait()
}
