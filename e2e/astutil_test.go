package e2e_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/moomou/xsqlparser"
	"github.com/moomou/xsqlparser/sqlastutil"
	"github.com/moomou/xsqlparser/dialect"
	"github.com/moomou/xsqlparser/sqlast"
)

func TestInspect(t *testing.T) {
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
			fname := fmt.Sprintf("./testdata/%s/", c.dir)
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

					stmt, err := parser.ParseStatement()
					if err != nil {
						t.Fatalf("%+v", err)
					}
					sqlast.Inspect(stmt, func(node sqlast.Node) bool {
						// fmt.Printf("%T\n", node)
						return true
					})
				})
			}
		})
	}
}

func TestApply(t *testing.T) {
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
			fname := fmt.Sprintf("./testdata/%s/", c.dir)
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

					stmt, err := parser.ParseStatement()
					if err != nil {
						t.Fatalf("%+v", err)
					}
					sqlastutil.Apply(stmt, func(c *sqlastutil.Cursor) bool {
						// fmt.Printf("%T\n", node)
						return true
					}, nil)
				})
			}
		})
	}
}
