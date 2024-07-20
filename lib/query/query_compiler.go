package query

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/richecr/pythonicsqlgo/lib/query/model"
)

type QueryCompiler struct {
	Simple     model.SimpleAttributes
	Statements []model.Statements
	Client     *sql.DB
	GroupsDict map[string][]model.Statements
	Components []string
}

func NewQueryCompiler(client *sql.DB) *QueryCompiler {
	compiler := &QueryCompiler{
		Client:     client,
		GroupsDict: make(map[string][]model.Statements),
		Components: []string{
			"columns",
			//"join",
			"where",
			"union",
			//"group",
			//"having",
			//"order",
			"limit",
			"offset",
		},
	}
	return compiler
}

func (qc *QueryCompiler) ToSQL() string {
	var firstStatements []string
	var endStatements []string

	for key, group := range groupBy(qc.Statements) {
		qc.GroupsDict[key] = group
	}

	for _, component := range qc.Components {
		var statement string
		switch component {
		case "columns":
			statement = qc.Columns()
		case "where":
			statement = qc.Where()
		case "limit":
			statement = qc.Limit()
		case "offset":
			statement = qc.Offset()
		default:
			statement = ""
		}

		if statement != "" {
			firstStatements = append(firstStatements, statement)
		} else {
			endStatements = append(endStatements, statement)
		}
	}

	return strings.Join(append(firstStatements, endStatements...), " ")
}

func (qc *QueryCompiler) Columns() string {
	qc.Simple.IsDQL = true

	if columns, ok := qc.GroupsDict["columns"]; ok {
		tableName := qc.Simple.TableName
		sql := fmt.Sprintf("select %s from %s", columns[0].Value, tableName)
		return sql
	}
	return ""
}

func (qc *QueryCompiler) Where() string {
	var sql []string
	if wheres, ok := qc.GroupsDict["where"]; ok {
		for _, clauseWhere := range wheres {
			stmt := qc.WhereStatements(clauseWhere)
			if len(sql) == 0 {
				sql = append(sql, stmt)
			} else {
				sql = append(sql, clauseWhere.Condition, stmt)
			}
		}
		return "where " + strings.Join(sql, "")
	}
	return ""
}

func (qc *QueryCompiler) WhereStatements(statement model.Statements) string {
	switch statement.Typ {
	case "where_operator":
		return fmt.Sprintf(" %s %s '%s'", statement.Column, statement.Operator, statement.Value)
	case "where_in":
		values := strings.Join(statement.Value.([]string), "','")
		return fmt.Sprintf(" %s in ('%s')", statement.Column, values)
	case "where_like":
		return fmt.Sprintf(" %s like '%s'", statement.Column, statement.Value)
	default:
		return ""
	}
}

func (qc *QueryCompiler) Limit() string {
	if qc.Simple.Limit > 0 {
		return fmt.Sprintf("limit %d", qc.Simple.Limit)
	}
	return ""
}

func (qc *QueryCompiler) Offset() string {
	if qc.Simple.Offset > 0 {
		return fmt.Sprintf("offset %d", qc.Simple.Offset)
	}
	return ""
}

func (qc *QueryCompiler) SetOptionsBuilder(statements []model.Statements, simple model.SimpleAttributes) {
	qc.Statements = statements
	qc.Simple = simple
}

func (qc *QueryCompiler) Exec() ([]byte, error) {
	query := qc.Simple.Raw.Sql
	if query == "" {
		query = qc.ToSQL()
	}
	qc.Reset()

	var result []map[string]interface{}

	if qc.Simple.IsDQL {
		rows, err := qc.Client.Query(query)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		cols, err := rows.Columns()
		if err != nil {
			fmt.Println(err)
		}
		colTypes, err := rows.ColumnTypes()
		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		vals := make([]interface{}, len(cols))
		for i, ct := range colTypes {
			switch ct.DatabaseTypeName() {
			case "VARCHAR", "TEXT":
				vals[i] = new(string)
			case "INT":
				vals[i] = new(int)
			default:
				vals[i] = new(interface{})
			}
		}

		for rows.Next() {
			scanArgs := make([]interface{}, len(cols))
			for i := range vals {
				scanArgs[i] = &vals[i]
			}

			err = rows.Scan(scanArgs...)
			if err != nil {
				fmt.Println(err)
				continue
			}

			rowMap := make(map[string]interface{})
			for i, colName := range cols {
				valPtr := vals[i]

				var val interface{}
				switch v := valPtr.(type) {
				case *string:
					if v != nil {
						val = *v
					}
				case *int:
					if v != nil {
						val = *v
					}
				default:
					val = v
				}

				rowMap[colName] = val
			}

			result = append(result, rowMap)
		}
	} else {
		_, err := qc.Client.Exec(query)
		if err != nil {
			return nil, err
		}
	}

	jsonResult, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	return jsonResult, nil
}

func (qc *QueryCompiler) Reset() {
	qc.GroupsDict = make(map[string][]model.Statements)
}

func groupBy(statements []model.Statements) map[string][]model.Statements {
	grouped := make(map[string][]model.Statements)
	for _, stmt := range statements {
		grouped[stmt.Grouping] = append(grouped[stmt.Grouping], stmt)
	}
	return grouped
}
