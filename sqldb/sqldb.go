package sqldb

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"unicode"

	"go.uber.org/zap"
)

type DBer interface {
	CreateTable(t TableData) error
	Insert(t TableData) error
}

type Sqldb struct {
	options
	db *sql.DB
}

type Field struct {
	Title string
	Type  string
}
type TableData struct {
	TableName   string
	ColumnNames []Field       // 标题字段
	Args        []interface{} // 数据
	DataCount   int           // 插入数据的数量
	AutoKey     bool
}

func New(opts ...Option) (*Sqldb, error) {
	options := defaultOptions
	for _, opt := range opts {
		opt(&options)
	}

	d := &Sqldb{}
	d.options = options

	if err := d.OpenDB(); err != nil {
		return nil, err
	}

	return d, nil
}

func (d *Sqldb) OpenDB() error {
	db, err := sql.Open("mysql", d.sqlURL)
	if err != nil {
		return err
	}

	db.SetMaxOpenConns(2048)
	db.SetMaxIdleConns(2048)

	if err = db.Ping(); err != nil {
		return err
	}

	d.db = db

	return nil
}

func (d *Sqldb) CreateTable(t TableData) error {
	sql, err := buildCreateTableSQL(t)
	if err != nil {
		return err
	}

	d.logger.Debug("crate table", zap.String("sql", sql))

	_, err = d.db.Exec(sql)

	return err
}

func buildCreateTableSQL(t TableData) (string, error) {
	if len(t.ColumnNames) == 0 {
		return "", errors.New("column can not be empty")
	}

	tableName, err := quoteIdentifier(t.TableName)
	if err != nil {
		return "", err
	}

	columns := make([]string, 0, len(t.ColumnNames)+1)
	if t.AutoKey {
		columns = append(columns, "`id` INT(12) NOT NULL PRIMARY KEY AUTO_INCREMENT")
	}

	for _, field := range t.ColumnNames {
		columnName, err := quoteIdentifier(field.Title)
		if err != nil {
			return "", err
		}
		columns = append(columns, columnName+` `+field.Type)
	}

	return `CREATE TABLE IF NOT EXISTS ` + tableName + ` (` + strings.Join(columns, ",") + `) ENGINE=MyISAM DEFAULT CHARSET=utf8;`, nil
}

func (d *Sqldb) Insert(t TableData) error {
	sql, err := buildInsertSQL(t)
	if err != nil {
		return err
	}

	d.logger.Debug("insert table", zap.String("sql", sql))
	_, err = d.db.Exec(sql, t.Args...)

	return err
}

func buildInsertSQL(t TableData) (string, error) {
	if len(t.ColumnNames) == 0 {
		return "", errors.New("empty column")
	}

	if t.DataCount <= 0 {
		return "", errors.New("data count must be positive")
	}

	if len(t.Args) != len(t.ColumnNames)*t.DataCount {
		return "", errors.New("args count does not match columns and data count")
	}

	tableName, err := quoteIdentifier(t.TableName)
	if err != nil {
		return "", err
	}

	columnNames := make([]string, 0, len(t.ColumnNames))
	for _, field := range t.ColumnNames {
		columnName, err := quoteIdentifier(field.Title)
		if err != nil {
			return "", err
		}
		columnNames = append(columnNames, columnName)
	}

	blank := ",(" + strings.Repeat(",?", len(t.ColumnNames))[1:] + ")"
	sql := `INSERT INTO ` + tableName + `(` + strings.Join(columnNames, ",") + `) VALUES `
	sql += strings.Repeat(blank, t.DataCount)[1:] + `;`

	return sql, nil
}

func quoteIdentifier(identifier string) (string, error) {
	if identifier == "" {
		return "", errors.New("identifier can not be empty")
	}

	for i, r := range identifier {
		switch {
		case r == '_':
		case unicode.IsLetter(r):
		case unicode.IsDigit(r):
			if i == 0 {
				return "", fmt.Errorf("identifier %q must not start with a digit", identifier)
			}
		default:
			return "", fmt.Errorf("identifier %q contains invalid character %q", identifier, r)
		}
	}

	return "`" + identifier + "`", nil
}
