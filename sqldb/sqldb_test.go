package sqldb

import "testing"

func TestBuildCreateTableSQLQuotesIdentifiers(t *testing.T) {
	sql, err := buildCreateTableSQL(TableData{
		TableName: "douban_book_list",
		ColumnNames: []Field{
			{Title: "书名", Type: "MEDIUMTEXT"},
			{Title: "URL", Type: "VARCHAR(255)"},
		},
		AutoKey: true,
	})
	if err != nil {
		t.Fatalf("buildCreateTableSQL() error = %v", err)
	}

	want := "CREATE TABLE IF NOT EXISTS `douban_book_list` (`id` INT(12) NOT NULL PRIMARY KEY AUTO_INCREMENT,`书名` MEDIUMTEXT,`URL` VARCHAR(255)) ENGINE=MyISAM DEFAULT CHARSET=utf8;"
	if sql != want {
		t.Fatalf("buildCreateTableSQL() = %q, want %q", sql, want)
	}
}

func TestBuildInsertSQLQuotesIdentifiers(t *testing.T) {
	sql, err := buildInsertSQL(TableData{
		TableName: "douban_book_list",
		ColumnNames: []Field{
			{Title: "书名", Type: "MEDIUMTEXT"},
			{Title: "URL", Type: "VARCHAR(255)"},
		},
		Args:      []interface{}{"Book A", "https://example.com", "Book B", "https://example.org"},
		DataCount: 2,
	})
	if err != nil {
		t.Fatalf("buildInsertSQL() error = %v", err)
	}

	want := "INSERT INTO `douban_book_list`(`书名`,`URL`) VALUES (?,?),(?,?);"
	if sql != want {
		t.Fatalf("buildInsertSQL() = %q, want %q", sql, want)
	}
}

func TestBuildSQLRejectsInvalidIdentifiers(t *testing.T) {
	tests := []struct {
		name string
		data TableData
	}{
		{
			name: "table with dot",
			data: TableData{
				TableName:   "crawler.books",
				ColumnNames: []Field{{Title: "name", Type: "MEDIUMTEXT"}},
				AutoKey:     true,
			},
		},
		{
			name: "column with backtick",
			data: TableData{
				TableName:   "books",
				ColumnNames: []Field{{Title: "na`me", Type: "MEDIUMTEXT"}},
				AutoKey:     true,
			},
		},
		{
			name: "column with space",
			data: TableData{
				TableName:   "books",
				ColumnNames: []Field{{Title: "book name", Type: "MEDIUMTEXT"}},
				AutoKey:     true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := buildCreateTableSQL(tt.data)
			if err == nil {
				t.Fatal("buildCreateTableSQL() error = nil, want error")
			}
		})
	}
}

func TestBuildInsertSQLRejectsMismatchedArgs(t *testing.T) {
	_, err := buildInsertSQL(TableData{
		TableName:   "books",
		ColumnNames: []Field{{Title: "name", Type: "MEDIUMTEXT"}},
		Args:        []interface{}{"Book A", "Book B"},
		DataCount:   1,
	})
	if err == nil {
		t.Fatal("buildInsertSQL() error = nil, want error")
	}
}
