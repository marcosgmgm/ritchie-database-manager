// This is the formula implementation class.
// Where you will code your methods and manipulate the inputs to perform the specific operation you wish to automate.

package formula

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gookit/color"
	"github.com/marcosgmgm/db-management/pkg/db/provider"
	"github.com/marcosgmgm/prompt-management/pkg/prompt"
	"io"
	"strings"
)

const (
	sqlTables = "select table_name from information_schema.tables where table_schema not in ('information_schema', 'pg_catalog')"
	sqlCol = "select column_name from information_schema.columns where table_name = $1"
)

type Formula struct {
	DBHost     string
	DBName     string
	DBUsername string
	DBPassword string
	DBPort     string
	DBSsl      string
	DBSchema   string
}

type stringMapper struct {}

func (t stringMapper) MapRow(row *sql.Row) (interface{}, error) {
	return nil, nil
}
func (t stringMapper) MapRows(rows *sql.Rows) ([]interface{}, error) {
	ts := make([]interface{}, 0)
	for rows.Next() {
		var t string
		err := rows.Scan(&t)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("error to scan row tables: %s", err.Error()))
		}
		ts = append(ts, t)
	}
	return ts, nil
}

type selectTableMapper struct {}

func (t selectTableMapper) MapRow(row *sql.Row) (interface{}, error) {
	return nil, nil
}
func (t selectTableMapper) MapRows(rows *sql.Rows) ([]interface{}, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error get columns name: %s", err.Error()))
	}
	vals := make([]interface{}, len(cols))
	for i := range cols {
		vals[i] = new(sql.RawBytes)
	}
	response := make([]interface{}, 0)
	response = append(response, cols)
	for rows.Next() {
		err := rows.Scan(vals...)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("error to scan row tables: %s", err.Error()))
		}
		var col []string
		for _, v := range vals {
			fv := v.(*sql.RawBytes)
			col = append(col, string(*fv))
		}
		response = append(response, col)
	}
	return response, err
}


type whereClause struct {
	column string
	value string
}

func (f Formula) Run(writer io.Writer) {
	pc, err := provider.NewPostgresConnector(f.DBHost, f.DBPort, f.DBUsername, f.DBPassword, f.DBName, f.DBSsl)
	if err != nil {
		result := color.FgRed.Render(fmt.Sprintf("error: %s.\n", err))
		if _, err := fmt.Fprintf(writer, result); err != nil {
			panic(err)
		}
		return
	}
	go pc.PingLoop()

	pe := provider.NewPostgresExecutor(pc)
	tableSelect, err := selectTable(f.DBSchema, pe)
	if err != nil {
		result := color.FgRed.Render(fmt.Sprintf("error select table: %s.\n", err))
		if _, err := fmt.Fprintf(writer, result); err != nil {
			panic(err)
		}
		return
	}
	ic, err := valuesColumn(f.DBSchema, pe, tableSelect)
	if err != nil {
		result := color.FgRed.Render(fmt.Sprintf("error clause insert: %s.\n", err))
		if _, err := fmt.Fprintf(writer, result); err != nil {
			panic(err)
		}
		return
	}
	sql := fmt.Sprintf("insert into %s", tableSelect)
	if len(ic) > 0 {
		sql = fmt.Sprintf("%s %s", sql, ic)
	}
	result := color.FgGreen.Render(fmt.Sprintf("Final SQL:\n%s\n", sql))
	if _, err := fmt.Fprintf(writer, result); err != nil {
		panic(err)
	}

	ib := prompt.NewSurveyBool()
	c, err := ib.Bool("execute sql:", []string{"yes", "no"})
	if err != nil {
		panic(err)
	}
	if !c {
		if _, err := fmt.Fprintf(writer, color.FgGreen.Render("finish without execute sql")); err != nil {
			panic(err)
		}
		return
	}

	executeInsert(f.DBSchema, pe, sql, writer)

}

func executeInsert(schema string, pe provider.PostgresExecutor, sql string, writer io.Writer) {
	r, err := pe.Exec(schema, sql)
	if err != nil {
		result := color.FgRed.Render(fmt.Sprintf("error exec sql: %s.\n", err))
		if _, err := fmt.Fprintf(writer, result); err != nil {
			panic(err)
		}
		return
	}
	ra, err := r.RowsAffected()
	if err != nil {
		result := color.FgRed.Render(fmt.Sprintf("error get rows affected: %s.\n", err))
		if _, err := fmt.Fprintf(writer, result); err != nil {
			panic(err)
		}
		return
	}
	if _, err := fmt.Fprintf(writer, color.FgGreen.Render(fmt.Sprintf("execute sql rows affected %d\n", ra))); err != nil {
		panic(err)
	}
}

func valuesColumn(schema string, pe provider.PostgresExecutor, table string) (string, error) {
	mapper := stringMapper{}
	cols, err := pe.Query(mapper, schema, sqlCol, table)
	if err != nil {
		return "", err
	}
	it := prompt.NewSurveyText()
	var vals []string
	var colsStr []string
	for _, c := range cols {
		val, err := it.Text(fmt.Sprintf("type value to field (%s):", c), false, "value column to insert")
		if err != nil {
			return "", err
		}
		vals = append(vals, val)
		colsStr = append(colsStr, c.(string))
	}
	clause := fmt.Sprintf("(%s)\n\tvalues (%s)", strings.Join(colsStr, ", "), strings.Join(vals, ", "))
	return clause, nil
}

func selectTable(schema string, pe provider.PostgresExecutor) (string, error) {
	mapper := stringMapper{}
	t, err := pe.Query(mapper, schema, sqlTables)
	if err != nil {
		return "", err
	}
	if len(t) < 1 {
		return "", errors.New("no has table in database")
	}
	s := make([]string, len(t))
	for i, v := range t {
		s[i] = fmt.Sprint(v)
	}
	is := prompt.NewSurveyList()
	return is.List("select table:", s)
}

