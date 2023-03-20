package main

import (
	"crawler/internal"
	"flag"
	"fmt"
	"net/url"
	"os"
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

	cmd, err := internal.NewCrawlerCommand(*uri, uint64(*deep), 1)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	serv, err := internal.NewHttpCrawler()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	_, err = serv.Run(cmd)

	fmt.Println(err)
}
