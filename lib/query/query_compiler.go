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

type mapStringScan struct {
	// cp are the column pointers
	cp []interface{}
	// row contains the final result
	row      map[string]string
	colCount int
	colNames []string
}

func NewMapStringScan(columnNames []string) *mapStringScan {
	lenCN := len(columnNames)
	s := &mapStringScan{
		cp:       make([]interface{}, lenCN),
		row:      make(map[string]string, lenCN),
		colCount: lenCN,
		colNames: columnNames,
	}
	for i := 0; i < lenCN; i++ {
		s.cp[i] = new(sql.RawBytes)
	}
	return s
}

func (s *mapStringScan) Update(rows *sql.Rows) error {
	if err := rows.Scan(s.cp...); err != nil {
		return err
	}

	for i := 0; i < s.colCount; i++ {
		if rb, ok := s.cp[i].(*sql.RawBytes); ok {
			s.row[s.colNames[i]] = string(*rb)
			*rb = nil // reset pointer to discard current value to avoid a bug
		} else {
			return fmt.Errorf("Cannot convert index %d column %s to type *sql.RawBytes", i, s.colNames[i])
		}
	}
	return nil
}

func (s *mapStringScan) Get() map[string]string {
	return s.row
}

/*
*

	using a string slice
*/
type stringStringScan struct {
	// cp are the column pointers
	cp []interface{}
	// row contains the final result
	row      []string
	colCount int
	colNames []string
}

func NewStringStringScan(columnNames []string) *stringStringScan {
	lenCN := len(columnNames)
	s := &stringStringScan{
		cp:       make([]interface{}, lenCN),
		row:      make([]string, lenCN*2),
		colCount: lenCN,
		colNames: columnNames,
	}
	j := 0
	for i := 0; i < lenCN; i++ {
		s.cp[i] = new(sql.RawBytes)
		s.row[j] = s.colNames[i]
		j = j + 2
	}
	return s
}

func (s *stringStringScan) Update(rows *sql.Rows) error {
	if err := rows.Scan(s.cp...); err != nil {
		return err
	}
	j := 0
	for i := 0; i < s.colCount; i++ {
		if rb, ok := s.cp[i].(*sql.RawBytes); ok {
			s.row[j+1] = string(*rb)
			*rb = nil // reset pointer to discard current value to avoid a bug
		} else {
			return fmt.Errorf("Cannot convert index %d column %s to type *sql.RawBytes", i, s.colNames[i])
		}
		j = j + 2
	}
	return nil
}

func (s *stringStringScan) Get() []string {
	return s.row
}

// rowMapString was the first implementation but it creates for each row a new
// map and pointers and is considered as slow. see benchmark
func rowMapString(columnNames []string, rows *sql.Rows) (map[string]string, error) {
	lenCN := len(columnNames)
	ret := make(map[string]string, lenCN)

	columnPointers := make([]interface{}, lenCN)
	for i := 0; i < lenCN; i++ {
		columnPointers[i] = new(sql.RawBytes)
	}

	if err := rows.Scan(columnPointers...); err != nil {
		return nil, err
	}

	for i := 0; i < lenCN; i++ {
		if rb, ok := columnPointers[i].(*sql.RawBytes); ok {
			ret[columnNames[i]] = string(*rb)
		} else {
			return nil, fmt.Errorf("Cannot convert index %d column %s to type *sql.RawBytes", i, columnNames[i])
		}
	}

	return ret, nil
}

func fck(err error) {
	if err != nil {
		fmt.Println(err)
	}
}
