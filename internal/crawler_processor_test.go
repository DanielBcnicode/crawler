package internal

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func Test_htmlCrawlerProcessor_Run(t *testing.T) {

	type args struct {
		url      string
		ancestor string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Basic test 5 urls",
			args: args{
				url:      "http://www.test1.com",
				ancestor: "ROOT",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wg := sync.WaitGroup{}
			wgOuter := sync.WaitGroup{}
			wgOuter.Add(1)
			fp := NewHttpCrawler(&FakeHtmlContentExtractor{})
			p := NewHtmlCrawler(
				fp,
				make(chan HtmlCrawlingPendingAddress),
				make(chan int),
				&wg,
				"www.test1.com",
			)

			p.Run(&wgOuter, tt.args.url, tt.args.ancestor)

			assert.Equal(t, 5, len(p.VisitedUrls()))
		})
	}
}
