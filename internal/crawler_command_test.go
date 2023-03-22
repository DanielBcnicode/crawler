package internal

import (
	"net/url"
	"reflect"
	"testing"
)

func TestNewCrawlerCommand(t *testing.T) {
	var (
		urlTest, _ = url.ParseRequestURI("http://www.host.com")
	)

	type args struct {
		url url.URL
	}
	tests := []struct {
		name    string
		args    args
		want    CrawlerCommand
		wantErr bool
		err     error
	}{
		{
			name: "good parameters in constructor",
			args: args{
				url: *urlTest,
			},
			want: CrawlerCommand{
				url: *urlTest,
			},
			wantErr: false,
			err:     nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewCrawlerCommand(tt.args.url)
			if ((err != nil) != tt.wantErr) || (err != tt.err) {
				t.Errorf("NewCrawlerCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewCrawlerCommand() got = %v, want %v", got, tt.want)
			}
		})
	}
}
