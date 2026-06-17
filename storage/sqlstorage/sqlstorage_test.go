package sqlstorage

import (
	"testing"

	"github.com/astenir/rulecrawl/engine"
	"github.com/astenir/rulecrawl/spider"
	"github.com/astenir/rulecrawl/sqldb"
	"go.uber.org/zap"
)

type fakeDB struct {
	createCount int
	insertCount int
	lastInsert  sqldb.TableData
}

func (f *fakeDB) CreateTable(_ sqldb.TableData) error {
	f.createCount++
	return nil
}

func (f *fakeDB) Insert(t sqldb.TableData) error {
	f.insertCount++
	f.lastInsert = t
	return nil
}

func TestSaveFlushesWhenBatchCountReached(t *testing.T) {
	const (
		taskName = "test_task"
		ruleName = "detail"
	)

	engine.Store.Add(&spider.Task{
		Options: spider.Options{Name: taskName},
		Rule: spider.RuleTree{
			Trunk: map[string]*spider.Rule{
				ruleName: {ItemFields: []string{"name"}},
			},
		},
	})
	defer delete(engine.Store.Hash, taskName)

	db := &fakeDB{}
	storage := &SQLStorage{
		db:    db,
		Table: make(map[string]struct{}),
		options: options{
			logger:     zap.NewNop(),
			BatchCount: 2,
		},
	}

	if err := storage.Save(dataCell(taskName, ruleName, "first")); err != nil {
		t.Fatalf("first Save() error = %v", err)
	}
	if db.insertCount != 0 {
		t.Fatalf("insertCount after first Save() = %d, want 0", db.insertCount)
	}

	if err := storage.Save(dataCell(taskName, ruleName, "second")); err != nil {
		t.Fatalf("second Save() error = %v", err)
	}
	if db.createCount != 1 {
		t.Fatalf("createCount = %d, want 1", db.createCount)
	}
	if db.insertCount != 1 {
		t.Fatalf("insertCount = %d, want 1", db.insertCount)
	}
	if db.lastInsert.DataCount != 2 {
		t.Fatalf("DataCount = %d, want 2", db.lastInsert.DataCount)
	}
}

func dataCell(taskName string, ruleName string, name string) *spider.DataCell {
	return &spider.DataCell{
		Data: map[string]interface{}{
			"Task": taskName,
			"Rule": ruleName,
			"Data": map[string]interface{}{
				"name": name,
			},
			"URL":  "https://example.com",
			"Time": "2026-06-17 00:00:00",
		},
	}
}
