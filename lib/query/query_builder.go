package query

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/richecr/pythonicsqlgo/lib/query/model"
)

type QueryBuilder struct {
	Compiler   *QueryCompiler
	Simple     model.SimpleAttributes
	Statements []model.Statements
	Client     *sql.DB
	Components []string
}

func NewQueryBuilder(compiler *QueryCompiler) *QueryBuilder {
	return &QueryBuilder{
		Compiler: compiler,
		Simple:   model.SimpleAttributes{},
	}
}

func (qb *QueryBuilder) ToSQL() string {
	qb.Compiler.SetOptionsBuilder(qb.Statements, qb.Simple)
	fmt.Println("Teste")
	sql := qb.Simple.Raw.Sql
	if sql == "" {
		sql = qb.Compiler.ToSQL()
	}
	return sql
}

func (qb *QueryBuilder) Select(columns []string) *QueryBuilder {
	value := strings.Join(columns, ", ")
	if len(columns) == 0 {
		value = "*"
	}
	qb.Statements = append(qb.Statements, model.Statements{
		Typ:      "select",
		Grouping: "columns",
		Value:    value,
	})
	return qb
}

func (qb *QueryBuilder) SetTableName(tableName string) *QueryBuilder {
	qb.Simple.TableName = tableName
	return qb
}

func (qb *QueryBuilder) From_(tableName string) *QueryBuilder {
	return qb.SetTableName(tableName)
}

func (qb *QueryBuilder) _where(
	typ string,
	column string,
	value interface{},
	condition string,
	operator string,
) *QueryBuilder {
	qb.Statements = append(qb.Statements, model.Statements{
		Typ:       typ,
		Grouping:  "where",
		Value:     value,
		Column:    column,
		Condition: condition,
		Operator:  operator,
	})
	return qb
}

func (qb *QueryBuilder) Where(column string, value interface{}, operator string) *QueryBuilder {
	qb._where("where_operator", column, value, " and", operator)
	return qb
}

func (qb *QueryBuilder) OrWhere(column string, value interface{}, operator string) *QueryBuilder {
	qb._where("where_operator", column, value, " or", operator)
	return qb
}

func (qb *QueryBuilder) WhereIn(column string, value interface{}) *QueryBuilder {
	qb._where("where_in", column, value, " and", "in")
	return qb
}

func (qb *QueryBuilder) OrWhereIn(column string, value interface{}) *QueryBuilder {
	qb._where("where_in", column, value, " or", "in")
	return qb
}

func (qb *QueryBuilder) WhereLike(column string, value interface{}) *QueryBuilder {
	qb._where("where_like", column, value, " and", "like")
	return qb
}

func (qb *QueryBuilder) OrWhereLike(column string, value interface{}) *QueryBuilder {
	qb._where("where_like", column, value, " or", "like")
	return qb
}

func (qb *QueryBuilder) Exec() ([]byte, error) {
	qb.Compiler.SetOptionsBuilder(qb.Statements, qb.Simple)
	result, err := qb.Compiler.Exec()
	qb.Reset()
	return result, err
}

func (qb *QueryBuilder) Raw(sql string) *QueryBuilder {
	qb.Simple.IsDQL = strings.HasPrefix(strings.ToLower(sql), "select")
	qb.Simple.Raw = model.Raw{Sql: sql}
	return qb
}

func (qb *QueryBuilder) Reset() {
	qb.Statements = []model.Statements{}
	qb.Simple = model.SimpleAttributes{}
}
