package e2e_test

// All queries are from https://www.w3schools.com/sql/sql_examples.asp

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/moomou/xsqlparser"
	"github.com/moomou/xsqlparser/dialect"
)

func TestParseQuery(t *testing.T) {

	cases := []struct {
		name string
		dir  string
	}{
		{
			name: "SELECT",
			dir:  "select",
		},
		{
			name: "CREATE TABLE",
			dir:  "create_table",
		},
		{
			name: "ALTER TABLE",
			dir:  "alter",
		},
		{
			name: "DROP TABLE",
			dir:  "drop_table",
		},
		{
			name: "CREATE INDEX",
			dir:  "create_index",
		},
		{
			name: "DROP INDEX",
			dir:  "drop_index",
		},
		{
			name: "INSERT",
			dir:  "insert",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fname := fmt.Sprintf("testdata/%s/", c.dir)
			files, err := ioutil.ReadDir(fname)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			for _, f := range files {
				if !strings.HasSuffix(f.Name(), ".sql") {
					continue
				}
				t.Run(f.Name(), func(t *testing.T) {
					fi, err := os.Open(fname + f.Name())
					if err != nil {
						t.Fatalf("%+v", err)
					}
					defer fi.Close()
					parser, err := xsqlparser.NewParser(fi, &dialect.GenericSQLDialect{})
					if err != nil {
						t.Fatalf("%+v", err)
					}

					orig, err := parser.ParseStatement()
					if err != nil {
						t.Fatalf("%+v", err)
					}
					recovered := orig.ToSQLString()

					parser, err = xsqlparser.NewParser(bytes.NewBufferString(recovered), &dialect.GenericSQLDialect{})
					if err != nil {
						t.Log(recovered)
						t.Fatalf("%+v", err)
					}

					stmt2, err := parser.ParseStatement()
					if err != nil {
						t.Fatalf("%+v", err)
					}

					recovered2 := stmt2.ToSQLString()

					parser, err = xsqlparser.NewParser(bytes.NewBufferString(recovered2), &dialect.GenericSQLDialect{})
					if err != nil {
						t.Log(recovered)
						t.Fatalf("%+v", err)
					}

					stmt3, err := parser.ParseStatement()
					if err != nil {
						t.Fatalf("%+v", err)
					}

					if astdiff := xsqlparser.CompareWithoutMarker(stmt2, stmt3); astdiff != "" {
						t.Logf(recovered)
						t.Errorf("should be same ast but diff:\n %s", astdiff)
					}
				})
			}
		})
	}
}
