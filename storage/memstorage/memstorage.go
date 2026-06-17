package memstorage

import (
	"sync"

	"github.com/astenir/rulecrawl/spider"
)

type MemoryStorage struct {
	mu    sync.RWMutex
	cells []*spider.DataCell
}

func New() *MemoryStorage {
	return &MemoryStorage{
		cells: make([]*spider.DataCell, 0),
	}
}

func (s *MemoryStorage) Save(datas ...*spider.DataCell) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, data := range datas {
		if data != nil {
			s.cells = append(s.cells, data)
		}
	}

	return nil
}

func (s *MemoryStorage) All() []*spider.DataCell {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cells := make([]*spider.DataCell, len(s.cells))
	copy(cells, s.cells)

	return cells
}

func (s *MemoryStorage) ByTask(taskName string) []*spider.DataCell {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cells := make([]*spider.DataCell, 0)
	for _, cell := range s.cells {
		if cell.GetTaskName() == taskName {
			cells = append(cells, cell)
		}
	}

	return cells
}

func (s *MemoryStorage) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.cells)
}

func (s *MemoryStorage) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cells = s.cells[:0]
}
