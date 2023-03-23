package internal

import (
	"github.com/stretchr/testify/assert"
	"io"
	"net/url"
	"strings"
	"testing"
)

type FakeHtmlContentExtractor struct {
}

func (f *FakeHtmlContentExtractor) Run(uri string) (realURL string, data io.Reader, err error) {
	html0 := `<!DOCTYPE html>
<html><body>
<ul><li><a href="/one">All</a></li>
    <li><a href="/two">On-site</a></li>
	<li><a href="/three">Hybrid</a></li>
	<li><a href="/four">Remote</a></li>
</ul></body></html>`
	html1 := `<!DOCTYPE html>
<html><body>
<ul><li><a href="https://test1.com/one">All</a></li>
    <li><a href="https://test1.com/two">On-site</a></li>
	<li><a href="https://test1.com/three">Hybrid</a></li>
	<li><a href="https://test1.com/four">Remote</a></li>
</ul></body></html>`
	switch uri {
	case "https://test1.com/one":
		return "https://test1.com/one", strings.NewReader(html1), nil
	case "https://test1.com/two":
		return "https://test1.com/two", strings.NewReader(html1), nil
	case "https://test1.com/three":
		return "https://test1.com/three", strings.NewReader(html1), nil
	case "https://test1.com/four":
		return "https://test1.com/four", strings.NewReader(html1), nil
	case "http://www.test1.com":
		return "https://test1.com", strings.NewReader(html0), nil
	default:
		return "", nil, ErrorHTTP
	}
}
func TestHttpCrawler_Run(t *testing.T) {
	url1, _ := url.Parse("http://www.test1.com")
	url2, _ := url.Parse("http://www.test2.com")
	type args struct {
		command CrawlerCommand
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]int
		want1   string
		wantErr bool
		error   error
	}{
		{
			name: "test happy path",
			args: args{
				command: CrawlerCommand{url: *url1},
			},
			want: map[string]int{
				"https://test1.com/four":  0,
				"https://test1.com/one":   0,
				"https://test1.com/three": 0,
				"https://test1.com/two":   0,
			},
			want1:   "https://test1.com",
			wantErr: false,
			error:   nil,
		},
		{
			name: "error in http",
			args: args{
				command: CrawlerCommand{url: *url2},
			},
			want:    nil,
			want1:   "",
			wantErr: true,
			error:   ErrorHTTP,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewHttpCrawler(&FakeHtmlContentExtractor{})

			got, got1, err := c.Run(tt.args.command)
			if tt.wantErr && assert.NotNil(t, err) {
				assert.Equal(t, tt.error, err)
			}
			assert.Equalf(t, tt.want, got, "Run(%v)", tt.args.command)
			assert.Equalf(t, tt.want1, got1, "Run(%v)", tt.args.command)
		})
	}
}
