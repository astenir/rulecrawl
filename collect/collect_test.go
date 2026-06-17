package collect

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/astenir/crawler/spider"
)

func TestBrowserFetchRejectsNonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("forbidden"))
	}))
	defer server.Close()

	fetcher := BrowserFetch{Timeout: time.Second}
	req := &spider.Request{
		URL:  server.URL,
		Task: &spider.Task{},
	}

	_, err := fetcher.Get(req)
	if err == nil {
		t.Fatal("Get() error = nil, want error")
	}
}
