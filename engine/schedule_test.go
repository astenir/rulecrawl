package engine

import (
	"testing"

	"github.com/astenir/rulecrawl/spider"
)

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
