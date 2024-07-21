package pythonic

import (
	"fmt"

	"github.com/richecr/pythonic_core/lib/dialects"
	"github.com/richecr/pythonic_core/lib/query"
	"github.com/richecr/pythonic_core/lib/query/model"
)

type PythonicSQL struct {
	Uri     string
	Dialect string
	Client  dialects.Client
	Query   *query.QueryBuilder
}

func NewPythonicSQL(config model.DatabaseConfiguration) (*PythonicSQL, error) {
	sql := &PythonicSQL{
		Uri:     config.Config.Uri,
		Dialect: config.Client,
	}

	client, err := sql.getClient()
	if err != nil {
		return nil, fmt.Errorf("error: %s", err.Error())
	}
	sql.Client = *client
	sql.Query = client.Builder

	return sql, nil
}

func (p *PythonicSQL) getClient() (*dialects.Client, error) {
	var client = dialects.NewClient(p.Dialect, p.Uri)

	return client, nil
}
