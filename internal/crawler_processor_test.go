package internal

import (
	"github.com/stretchr/testify/assert"
	"strings"
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
		want string
	}{
		{
			name: "Basic test 5 urls",
			args: args{
				url:      "http://www.test1.com",
				ancestor: "ROOT",
			},
			want: ` https://test1.com
   https://test1.com/one
   https://test1.com/two
   https://test1.com/three
   https://test1.com/four
 https://test1.com/one
   https://test1.com/one
   https://test1.com/two
   https://test1.com/three
   https://test1.com/four
 https://test1.com/two
   https://test1.com/one
   https://test1.com/two
   https://test1.com/three
   https://test1.com/four
 https://test1.com/three
   https://test1.com/one
   https://test1.com/two
   https://test1.com/three
   https://test1.com/four
 https://test1.com/four
   https://test1.com/one
   https://test1.com/two
   https://test1.com/three
   https://test1.com/four
`,
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
				&wg,
				"www.test1.com",
			)

			p.Run(&wgOuter, tt.args.url, tt.args.ancestor)
			res := p.Response()

			assert.Equal(t, 5, len(p.VisitedUrls()))
			assert.ElementsMatch(t, strings.Split(res.(string), "\n"), strings.Split(tt.want, "\n"))
		})
	}
}
