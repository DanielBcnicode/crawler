package internal

import (
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

func TestWebContentExtract_Run(t *testing.T) {
	type args struct {
		uri string
	}
	tests := []struct {
		name        string
		args        args
		wantRealURL string
		wantErr     bool
	}{
		{
			name:        "Test real connection to google using http",
			args:        args{uri: "http://www.gogle.com"},
			wantRealURL: "https://www.google.com/",
			wantErr:     false,
		},
		{
			name:        "Test real connection to no existent url",
			args:        args{uri: "http://www.notexisturl.com"},
			wantRealURL: "",
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewWebContentExtrat()
			gotRealURL, gotData, err := c.Run(tt.args.uri)
			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotRealURL != tt.wantRealURL {
				t.Errorf("Run() gotRealURL = %v, want %v", gotRealURL, tt.wantRealURL)
			}

			if !tt.wantErr {
				data := ""
				if gotData != nil {
					d, _ := io.ReadAll(gotData)
					data = string(d)
				}

				assert.True(t, len(data) > 100)
			}

		})
	}
}
