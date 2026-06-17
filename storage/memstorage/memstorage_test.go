package memstorage

import (
	"testing"

	"github.com/astenir/rulecrawl/spider"
)

func TestMemoryStorageSaveAndRead(t *testing.T) {
	storage := New()

	if err := storage.Save(dataCell("books", "Book A"), nil, dataCell("groups", "Group A")); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	if storage.Len() != 2 {
		t.Fatalf("Len() = %d, want 2", storage.Len())
	}

	cells := storage.All()
	if len(cells) != 2 {
		t.Fatalf("len(All()) = %d, want 2", len(cells))
	}

	cells[0] = dataCell("changed", "Changed")
	if got := storage.All()[0].GetTaskName(); got != "books" {
		t.Fatalf("All() leaked backing slice mutation, first task = %q", got)
	}
}

func TestMemoryStorageByTask(t *testing.T) {
	storage := New()
	if err := storage.Save(dataCell("books", "Book A"), dataCell("groups", "Group A"), dataCell("books", "Book B")); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	books := storage.ByTask("books")
	if len(books) != 2 {
		t.Fatalf("len(ByTask(\"books\")) = %d, want 2", len(books))
	}
	for _, book := range books {
		if got := book.GetTaskName(); got != "books" {
			t.Fatalf("ByTask returned task %q, want books", got)
		}
	}
}

func TestMemoryStorageReset(t *testing.T) {
	storage := New()
	if err := storage.Save(dataCell("books", "Book A")); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	storage.Reset()

	if storage.Len() != 0 {
		t.Fatalf("Len() after Reset() = %d, want 0", storage.Len())
	}
}

func dataCell(taskName string, name string) *spider.DataCell {
	return &spider.DataCell{
		Data: map[string]interface{}{
			"Task": taskName,
			"Data": map[string]interface{}{
				"name": name,
			},
		},
	}
}
