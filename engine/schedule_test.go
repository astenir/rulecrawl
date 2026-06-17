package engine

import (
	"errors"
	"testing"

	"github.com/astenir/rulecrawl/spider"
	"github.com/astenir/rulecrawl/storage/memstorage"
	"go.uber.org/zap"
)

type fakeFetcher struct {
	body []byte
	err  error
}

func (f fakeFetcher) Get(_ *spider.Request) ([]byte, error) {
	return f.body, f.err
}

type recordingScheduler struct {
	pushed []*spider.Request
}

func (s *recordingScheduler) Schedule() {}

func (s *recordingScheduler) Push(requests ...*spider.Request) {
	s.pushed = append(s.pushed, requests...)
}

func (s *recordingScheduler) Pull() *spider.Request {
	return nil
}

func TestCrawlerProcessRequestAndHandleResult(t *testing.T) {
	storage := memstorage.New()
	task := spider.NewTask(
		spider.WithName("books"),
		spider.WithStorage(storage),
		spider.WithFetcher(fakeFetcher{body: []byte("Book A")}),
		spider.WithWaitTime(0),
	)
	task.Rule.Trunk = map[string]*spider.Rule{
		"detail": {
			ParseFunc: func(ctx *spider.Context) (spider.ParseResult, error) {
				return spider.ParseResult{
					Items: []interface{}{
						ctx.Output(map[string]interface{}{
							"name": string(ctx.Body),
						}),
					},
				}, nil
			},
		},
	}

	crawler := NewEngine(WithLogger(zap.NewNop()))
	req := &spider.Request{
		Task:     task,
		URL:      "https://example.com/books/1",
		Method:   "GET",
		RuleName: "detail",
	}

	result, ok := crawler.processRequest(req)
	if !ok {
		t.Fatal("processRequest() ok = false, want true")
	}

	crawler.handleResult(result)

	if storage.Len() != 1 {
		t.Fatalf("storage.Len() = %d, want 1", storage.Len())
	}
	cell := storage.All()[0]
	data := cell.Data["Data"].(map[string]interface{})
	if data["name"] != "Book A" {
		t.Fatalf("stored name = %q, want Book A", data["name"])
	}
}

func TestCrawlerProcessRequestFetchFailureRetries(t *testing.T) {
	scheduler := &recordingScheduler{}
	task := spider.NewTask(
		spider.WithName("books"),
		spider.WithFetcher(fakeFetcher{err: errors.New("fetch failed")}),
		spider.WithWaitTime(0),
	)
	task.Rule.Trunk = map[string]*spider.Rule{
		"detail": {ParseFunc: func(_ *spider.Context) (spider.ParseResult, error) {
			return spider.ParseResult{}, nil
		}},
	}

	crawler := NewEngine(
		WithLogger(zap.NewNop()),
		WithScheduler(scheduler),
	)
	req := &spider.Request{
		Task:     task,
		URL:      "https://example.com/books/1",
		Method:   "GET",
		RuleName: "detail",
	}

	_, ok := crawler.processRequest(req)
	if ok {
		t.Fatal("processRequest() ok = true, want false")
	}
	if len(scheduler.pushed) != 1 {
		t.Fatalf("len(scheduler.pushed) = %d, want 1", len(scheduler.pushed))
	}
	if scheduler.pushed[0] != req {
		t.Fatal("processRequest did not retry the failed request")
	}
}

func TestCrawlerProcessRequestValidateFailureRetries(t *testing.T) {
	scheduler := &recordingScheduler{}
	task := spider.NewTask(
		spider.WithName("books"),
		spider.WithFetcher(fakeFetcher{body: []byte("bad body")}),
		spider.WithWaitTime(0),
	)
	task.Rule.Trunk = map[string]*spider.Rule{
		"detail": {
			Validate: func(_ []byte) error {
				return errors.New("invalid body")
			},
			ParseFunc: func(_ *spider.Context) (spider.ParseResult, error) {
				return spider.ParseResult{}, nil
			},
		},
	}

	crawler := NewEngine(
		WithLogger(zap.NewNop()),
		WithScheduler(scheduler),
	)
	req := &spider.Request{
		Task:     task,
		URL:      "https://example.com/books/1",
		Method:   "GET",
		RuleName: "detail",
	}

	_, ok := crawler.processRequest(req)
	if ok {
		t.Fatal("processRequest() ok = true, want false")
	}
	if len(scheduler.pushed) != 1 {
		t.Fatalf("len(scheduler.pushed) = %d, want 1", len(scheduler.pushed))
	}
}

func TestCrawlerProcessRequestParseFailureDoesNotStore(t *testing.T) {
	storage := memstorage.New()
	task := spider.NewTask(
		spider.WithName("books"),
		spider.WithStorage(storage),
		spider.WithFetcher(fakeFetcher{body: []byte("Book A")}),
		spider.WithWaitTime(0),
	)
	task.Rule.Trunk = map[string]*spider.Rule{
		"detail": {
			ParseFunc: func(_ *spider.Context) (spider.ParseResult, error) {
				return spider.ParseResult{}, errors.New("parse failed")
			},
		},
	}

	crawler := NewEngine(WithLogger(zap.NewNop()))
	req := &spider.Request{
		Task:     task,
		URL:      "https://example.com/books/1",
		Method:   "GET",
		RuleName: "detail",
	}

	result, ok := crawler.processRequest(req)
	if ok {
		t.Fatal("processRequest() ok = true, want false")
	}

	crawler.handleResult(result)

	if storage.Len() != 0 {
		t.Fatalf("storage.Len() = %d, want 0", storage.Len())
	}
}

func TestAddJSTaskUsesModelProperties(t *testing.T) {
	store := &CrawlerStore{
		Hash: make(map[string]*spider.Task),
	}

	store.AddJSTask(&spider.TaskModle{
		Property: spider.Property{
			Name:     "js_task",
			URL:      "https://example.com",
			Cookie:   "session=1",
			WaitTime: 2,
			Reload:   true,
			MaxDepth: 4,
		},
		Root: `[];`,
		Rules: []spider.RuleModle{
			{Name: "detail", ParseFunc: `null;`},
		},
	})

	task, ok := store.Hash["js_task"]
	if !ok {
		t.Fatal("AddJSTask did not register task by model name")
	}

	if task.Name != "js_task" {
		t.Fatalf("task.Name = %q, want js_task", task.Name)
	}
	if task.URL != "https://example.com" {
		t.Fatalf("task.URL = %q, want https://example.com", task.URL)
	}
	if task.Cookie != "session=1" {
		t.Fatalf("task.Cookie = %q, want session=1", task.Cookie)
	}
	if task.WaitTime != 2 {
		t.Fatalf("task.WaitTime = %d, want 2", task.WaitTime)
	}
	if !task.Reload {
		t.Fatal("task.Reload = false, want true")
	}
	if task.MaxDepth != 4 {
		t.Fatalf("task.MaxDepth = %d, want 4", task.MaxDepth)
	}
	if _, ok := task.Rule.Trunk["detail"]; !ok {
		t.Fatal("AddJSTask did not register JS parse rule")
	}
}

func TestAddJSTaskKeepsTaskDefaults(t *testing.T) {
	store := &CrawlerStore{
		Hash: make(map[string]*spider.Task),
	}

	store.AddJSTask(&spider.TaskModle{
		Property: spider.Property{Name: "js_task_defaults"},
		Root:     `[];`,
	})

	task := store.Hash["js_task_defaults"]
	if task.WaitTime != 5 {
		t.Fatalf("task.WaitTime = %d, want default 5", task.WaitTime)
	}
	if task.MaxDepth != 5 {
		t.Fatalf("task.MaxDepth = %d, want default 5", task.MaxDepth)
	}
}
