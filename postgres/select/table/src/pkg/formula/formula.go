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
	"github.com/jedib0t/go-pretty/table"
	"io"
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
	wc, err := whereColumns(f.DBSchema, pe, tableSelect)
	if err != nil {
		result := color.FgRed.Render(fmt.Sprintf("error choose column: %s.\n", err))
		if _, err := fmt.Fprintf(writer, result); err != nil {
			panic(err)
		}
		return
	}
	sql := fmt.Sprintf("select * from %s\n", tableSelect)
	if len(wc) > 0 {
		sql = fmt.Sprintf("%s%s", sql, wc)
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

	executeSelect(f.DBSchema, pe, sql, writer)

}

func executeSelect(schema string, pe provider.PostgresExecutor, sqlQuery string, writer io.Writer) {
	resultTable, err := pe.Query(selectTableMapper{}, schema, sqlQuery)
	if err != nil {
		result := color.FgRed.Render(fmt.Sprintf("error: %s.\n", err))
		if _, err := fmt.Fprintf(writer, result); err != nil {
			panic(err)
		}
		return
	}
	t := table.NewWriter()
	t.SetOutputMirror(writer)
	for i := range resultTable {
		val := resultTable[i].([]string)
		r := table.Row{}
		for h := range val {
			r = append(r, val[h])
		}
		if i == 0 {
			t.AppendHeader(r)
		} else {
			t.AppendRows([]table.Row{r})
		}
	}
	t.SetStyle(table.StyleLight)
	t.Render()
}

func whereColumns(schema string, pe provider.PostgresExecutor, table string) (string, error) {
	mapper := stringMapper{}
	cols, err := pe.Query(mapper, schema, sqlCol, table)
	if err != nil {
		return "", err
	}
	s := []string{"no where"}
	for _, v := range cols {
		s = append(s, v.(string))
	}
	is := prompt.NewSurveyList()
	sc, err := is.List("select columns for where clause:", s)
	if err != nil {
		return "", err
	}
	var cw []whereClause
	for sc != "no where" {
		it := prompt.NewSurveyText()
		w, err := it.Text(fmt.Sprintf("type value where to column (%s)", sc), true)
		if err != nil {
			return "", err
		}
		cw = append(cw, whereClause{
			column: sc,
			value:  w,
		})
		s = removeVal(s, sc)
		sc, err = is.List("select columns for where clause:", s)
		if err != nil {
			return "", err
		}
	}
	if len(cw) > 0 {
		wc := "\twhere 1 = 1\n"
		for _, w := range cw {
			wc = fmt.Sprintf("%s\tand %s = %s", wc, w.column, w.value)
		}
		return wc, nil
	}
	return "", nil
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

func removeVal(vals []string, val string) []string {
	var fVals []string
	for _, v := range vals {
		if v != val {
			fVals = append(fVals, v)
		}
	}
	return fVals
}

