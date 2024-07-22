package dialects

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/mattn/go-sqlite3"

	"github.com/richecr/pythonic_core/lib/query"
)

type Client struct {
	Uri      string
	Dialect  string
	Database *sql.DB
	Compiler *query.QueryCompiler
	Builder  *query.QueryBuilder
}

func NewClient(dialect string, uri string) *Client {
	client := &Client{
		Uri:      uri,
		Dialect:  dialect,
		Database: nil,
		Compiler: nil,
		Builder:  nil,
	}
	client.init()
	return client
}

func (c *Client) init() error {
	db, err := sql.Open(c.Dialect, c.Uri)
	if err != nil {
		return fmt.Errorf("error: %s", err.Error())
	} else {
		c.Database = db
		c.Compiler = query.NewQueryCompiler(db)
		c.Builder = query.NewQueryBuilder(c.Compiler)
	}

	return nil
}

func (c *Client) IsConnected() (bool, error) {
	err := c.Database.Ping()
	if err != nil {
		return false, fmt.Errorf("error: %s", err.Error())
	}
	return true, nil
}

func (c *Client) Disconnect() error {
	err := c.Database.Close()
	if err != nil {
		return err
	}
	return nil
}
