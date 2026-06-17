package spider

import (
	"testing"
)

type fakeFetcher struct {
	body []byte
}

func (f fakeFetcher) Get(_ *Request) ([]byte, error) {
	return f.body, nil
}

func TestRequestFetchWithoutLimiter(t *testing.T) {
	req := &Request{
		Task: &Task{
			Options: Options{
				Fetcher: fakeFetcher{body: []byte("ok")},
			},
		},
	}

	body, err := req.Fetch()
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if string(body) != "ok" {
		t.Fatalf("Fetch() body = %q, want %q", string(body), "ok")
	}
}

func TestRequestFetchRequiresFetcher(t *testing.T) {
	req := &Request{Task: &Task{}}

	_, err := req.Fetch()
	if err == nil {
		t.Fatal("Fetch() error = nil, want error")
	}
}
