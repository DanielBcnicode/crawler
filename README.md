# Crawler
## ParserDigital technical test

The idea is to create a Crawler to visit only one subdomain.
This implementation is a basic one based on a CLI app.

I developed a primary solution with the idea of increasing the complexity until I have
a scalable app.
I used interfaces in the services to make the application easy to scale or change the behaviour.

This version does not use any persistent service to store the results and the urls visited,
nor does it use any external message broker to have an asynchronous multi-pod service.
However, it uses concurrency to crawl up to 30 urls at the same time.
The number of go routines is hardcoded at 30, but it can be improved changing it via an environment var or parameter.
The current system is designed to crawl a whole site from the same application instance.
The main service is `htmlCrawlerProcessor`, it is called from the `cmd/crawler.go` application.
Is easy to implement other system (e.g.,web API) to use it to have more than one front-end.

To improve (due to the time constraint):
- Create a docker/-compose solution to compile y to create an image of the product
- Use external cache or database to store the visited urls.
- Use an event dispatcher system using pub/sub as RabitMQ or Kafka to have a perfect distributed system.
- Implement TRYs when an url is not reachable.
- Limit the crawling frequency to avoid being banned, creating a scheduler.
- A makefile with the commands used... build, test, etc..
- ...

## Requirements
- Golang
- pprof and graphviz
- An internet connection

## Composition
- Only a CLI app

## How to execute the tests, build and process
### Run tests
There is an integration test that needs internet connection. It connects with Google
Execute this commands in the main folder:
```shell
go mod vendor
go test -cover ./...
```
The current results are:
```shell
?       crawler/cmd    [no test files]
ok      crawler/internal        5.906s  coverage: 87.6% of statements
```
### Build
Execute :
```shell
go build cmd/crawler.go 
```
### Run
```shell
go run cmd/crawler.go  https://www.parserdigital.com
```
### Run with profiling
This command will create the file crawler.prof to be used with `pprof` tool
```shell
go run cmd/crawler.go -p https://www.parserdigital.com
```
To view the data generated:
```shell
pprof -http=localhost:8090 crawler.prof
```
It wil open a navigator app with the pprof data, or you can access to `http://localhost:8090`

