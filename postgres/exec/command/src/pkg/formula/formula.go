// This is the formula implementation class.
// Where you will code your methods and manipulate the inputs to perform the specific operation you wish to automate.

package formula

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gookit/color"
	"github.com/jedib0t/go-pretty/table"
	"github.com/marcosgmgm/db-management/pkg/db/provider"
	"github.com/marcosgmgm/prompt-management/pkg/prompt"
	"github.com/xwb1989/sqlparser"
	"io"
)

const exec = "exec"

type Formula struct {
	CommandType string
	DBHost      string
	DBName      string
	DBUsername  string
	DBPassword  string
	DBPort      string
	DBSsl       string
}

type column struct {
	name   string
	cType  string
	length int64
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

func (f Formula) Run(writer io.Writer) {


	pc, err := provider.NewPostgresConnector(f.DBHost, f.DBPort, f.DBUsername, f.DBPassword, f.DBName, f.DBSsl)
	if err != nil {
		result := color.FgRed.Render(fmt.Sprintf("error: %s.\n", err))
		if _, err := fmt.Fprintf(writer, result); err != nil {
			panic(err)
		}
		return
	}
	pe := provider.NewPostgresExecutor(pc)

	it := prompt.NewSurveyMultiline()
	sql, err := it.MultiLineText("type sql:", true)
	if err != nil {
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

	_, err = sqlparser.Parse(sql)
	if err != nil {
		if _, err := fmt.Fprintf(writer, color.FgGreen.Render("sql not valid")); err != nil {
			panic(err)
		}
		return
	}

	if f.CommandType == exec {
		execSql(pe, sql, writer)
	} else {
		querySql(pe, sql, writer)
	}
}

func execSql(pe provider.PostgresExecutor, sql string, writer io.Writer) {
	r, err := pe.Exec(sql)
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

func querySql(pe provider.PostgresExecutor, sqlQuery string, writer io.Writer) {
	resultTable, err := pe.Query(selectTableMapper{}, sqlQuery)
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

