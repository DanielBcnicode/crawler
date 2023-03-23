package main

import (
	"crawler/internal"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"runtime/pprof"
	"sync"
)

func main() {
	profileActive := flag.Bool("p", false, "-p profile activation")
	flag.Parse()

	fmt.Println("Profile: ", *profileActive)
	fmt.Println("Arguments number: ", flag.NArg())
	fmt.Println("Args: ", flag.Arg(0))

	uri, err := url.ParseRequestURI(flag.Arg(0))
	if err != nil {
		fmt.Println("the parameter specified is not a valid url: ", flag.Arg(0))
		os.Exit(1)
	}
	fmt.Println("Crawling the url: ", uri)

	if *profileActive {
		f, err := os.Create("crawler.prof")
		if err != nil {
			fmt.Println(err)
			return
		}
		_ = pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	wg := sync.WaitGroup{}
	pendingChan := make(chan internal.HtmlCrawlingPendingAddress)
	serv := internal.NewHttpCrawler(
		internal.NewWebContentExtrat(),
	)
	processorService := internal.NewHtmlCrawler(serv, pendingChan, &wg, uri.Host)

	wg.Add(1)

	go processorService.Run(&wg, uri.String(), internal.ROOT)

	cs := make(chan os.Signal, 1)
	signal.Notify(cs, os.Interrupt)
	go func() {
		<-cs
		log.Println("Goodbye")
		fmt.Println(processorService.Response())
		if *profileActive {
			pprof.StopCPUProfile()
		}
		os.Exit(1)
	}()

	wg.Wait()
	resp := processorService.Response()
	fmt.Println(resp.(string))
}
