// This is the formula implementation class.
// Where you will code your methods and manipulate the inputs to perform the specific operation you wish to automate.

package formula

import (
	"fmt"
	"github.com/gookit/color"
	"github.com/marcosgmgm/db-management/pkg/db/provider"
	"github.com/marcosgmgm/prompt-management/pkg/prompt"
	"io"
	"strings"
)

type Formula struct {
	Table      string
	DBHost     string
	DBName     string
	DBUsername string
	DBPassword string
	DBPort     string
	DBSsl      string
	DBSchema   string
}

type column struct {
	name   string
	cType  string
	length int64
}

func (c Formula) Run(writer io.Writer) {
	var columns []column
	columnsFinal, err := loadColumn(columns)
	if err != nil {
		result := color.FgRed.Render(fmt.Sprintf("error load db columns: %s.\n", err))
		if _, err := fmt.Fprintf(writer, result); err != nil {
			panic(err)
		}
		return
	}
	columnsPk, err := loadColumnPk(columnsFinal)
	if err != nil {
		result := color.FgRed.Render(fmt.Sprintf("error define primary key: %s.\n", err))
		if _, err := fmt.Fprintf(writer, result); err != nil {
			panic(err)
		}
		return
	}

	sql := generateSql(c.Table, columnsFinal, columnsPk)
	result := color.FgGreen.Render(fmt.Sprintf("Final SQL:\n%s\n", sql))
	if _, err := fmt.Fprintf(writer, result); err != nil {
		panic(err)
	}

	ib := prompt.NewSurveyBool()
	cf, err := ib.Bool("execute sql:", []string{"yes", "no"})
	if err != nil {
		panic(err)
	}
	if !cf {
		if _, err := fmt.Fprintf(writer, color.FgGreen.Render("finish without execute sql")); err != nil {
			panic(err)
		}
		return
	}

	pc, err := provider.NewPostgresConnector(c.DBHost, c.DBPort, c.DBUsername, c.DBPassword, c.DBName, c.DBSsl)
	if err != nil {
		result := color.FgRed.Render(fmt.Sprintf("error: %s.\n", err))
		if _, err := fmt.Fprintf(writer, result); err != nil {
			panic(err)
		}
		return
	}
	pe := provider.NewPostgresExecutor(pc)

	_, err = pe.Exec(c.DBSchema, sql)
	if err != nil {
		result := color.FgRed.Render(fmt.Sprintf("error: %s.\n", err))
		if _, err := fmt.Fprintf(writer, result); err != nil {
			panic(err)
		}
		return
	}
	result = color.FgGreen.Render("table created if not exists")
	if _, err := fmt.Fprintf(writer, result); err != nil {
		panic(err)
	}

}

func generateSql(table string, columns []column, pks []string) string {
	var lines string
	for _, c := range columns {
		line := fmt.Sprintf("\t%s %s", c.name, c.cType)
		if c.length > 0 {
			line = fmt.Sprintf("%s(%d)", line, c.length)
		}
		if len(lines) > 0 {
			lines = fmt.Sprintf("%s,\n%s", lines, line)
		} else {
			lines = line
		}
	}
	if len(pks) > 0 {
		lines = fmt.Sprintf("%s,\n\tprimary key (%s)", lines, strings.Join(pks, ","))
	}
	return fmt.Sprintf("create table if not exists %s (\n%s\n)", table, lines)
}

func loadColumn(columns []column) ([]column, error) {
	it := prompt.NewSurveyText()
	name, err := it.Text("type column name:", true, "column name")
	if err != nil {
		return nil, err
	}
	is := prompt.NewSurveyList()
	clType, err := is.List("create column type:", []string{"bigint", "boolean", "character", "character varying",
		"date", "integer", "json", "jsonb", "numeric", "text", "time", "timestamp", "uuid"})
	if err != nil {
		return nil, err
	}
	var l int64
	if defineLength(clType) {
		ii := prompt.NewSurveyInt()
		l, err = ii.Int("type length column:")
		if err != nil {
			return nil, err
		}
	}
	col := column{
		name:   name,
		cType:  clType,
		length: l,
	}
	columns = append(columns, col)

	ib := prompt.NewSurveyBool()
	c, err := ib.Bool("add new column:", []string{"yes", "no"})
	if err != nil {
		return nil, err
	}
	if c {
		return loadColumn(columns)
	}
	return columns, nil
}

func loadColumnPk(columns []column) ([]string, error) {
	lCol := make([]string, len(columns))
	for i, _ := range columns {
		lCol[i] = columns[i].name
	}
	il := prompt.NewSurveyList()
	col, err := il.List("create primary key column:", lCol)
	if err != nil {
		return nil, err
	}
	var fCol []string
	fCol = append(fCol, col)
	if len(columns) > 1 {
		ib := prompt.NewSurveyBool()
		for {
			c, err := ib.Bool("add new column:", []string{"yes", "no"})
			if err != nil {
				return nil, err
			}
			if c {
				columns := removeVal(columns, col)
				addCol, err := loadColumnPk(columns)
				if err != nil {
					return nil, err
				}
				fCol = append(fCol, addCol...)
			}
			break
		}
	}
	return fCol, err
}

func removeVal(columns []column, col string) []column {
	var fCols []column
	for _, c := range columns {
		if c.name != col {
			fCols = append(fCols, c)
		}
	}
	return fCols
}

func defineLength(t string) bool {
	return t == "character" || t == "character varying"
}
