package internal

import (
	"net/url"
	"reflect"
	"testing"
)

func TestNewCrawlerCommand(t *testing.T) {
	var (
		urlTest, _        = url.ParseRequestURI("http://www.host.com")
		maxDeep    uint64 = 2
	)

	type args struct {
		url     url.URL
		maxDeep uint64
		deep    uint64
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
				url:     *urlTest,
				maxDeep: maxDeep,
				deep:    1,
			},
			want: CrawlerCommand{
				url:     *urlTest,
				maxDeep: maxDeep,
				deep:    1,
			},
			wantErr: false,
			err:     nil,
		},
		{
			name: "maxDeep is zero",
			args: args{
				url:     *urlTest,
				maxDeep: 0,
				deep:    1,
			},
			want:    CrawlerCommand{},
			wantErr: true,
			err:     ErrorMaxDeepCanNotBeZero,
		},
		{
			name: "deep is greater than maxDeep",
			args: args{
				url:     *urlTest,
				maxDeep: 2,
				deep:    4,
			},
			want:    CrawlerCommand{},
			wantErr: true,
			err:     ErrorDeepTooHigh,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewCrawlerCommand(tt.args.url, tt.args.maxDeep, tt.args.deep)
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
