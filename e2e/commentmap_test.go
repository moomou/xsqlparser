package e2e_test

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/moomou/xsqlparser"
	"github.com/moomou/xsqlparser/dialect"
	"github.com/moomou/xsqlparser/sqlast"
	"github.com/moomou/xsqlparser/sqltoken"
)

func parseFile(t *testing.T, src string) *sqlast.File {
	t.Helper()
	parser, err := xsqlparser.NewParser(strings.NewReader(src), &dialect.GenericSQLDialect{}, xsqlparser.ParseComment())
	if err != nil {
		t.Fatal(err)
	}

	f, err := parser.ParseFile()
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func compareComment(t *testing.T, expect, actual []*sqlast.CommentGroup) {
	t.Helper()
	if diff := cmp.Diff(expect, actual); diff != "" {
		t.Error(diff)
	}
}

func TestNewCommentMap(t *testing.T) {

	t.Run("associate with single statement", func(t *testing.T) {
		f := parseFile(t, `
--test
SELECT * from test;
`)

		m := sqlast.NewCommentMap(f)
		compareComment(t, m[f.Stmts[0]], []*sqlast.CommentGroup{
			{
				List: []*sqlast.Comment{
					{
						Text: "test",
						From: sqltoken.NewPos(2, 1),
						To:   sqltoken.NewPos(2, 7),
					},
				},
			},
		})
	})

	t.Run("associate with multi statements", func(t *testing.T) {

		f := parseFile(t, `
--select
SELECT * from test;

/*
insert
*/
INSERT INTO tbl_name (col1,col2) VALUES(15,col1*2);
`)
		m := sqlast.NewCommentMap(f)

		compareComment(t, m[f.Stmts[0]], []*sqlast.CommentGroup{
			{
				List: []*sqlast.Comment{
					{
						Text: "select",
						From: sqltoken.NewPos(2, 1),
						To:   sqltoken.NewPos(2, 9),
					},
				},
			},
		})

		compareComment(t, m[f.Stmts[1]], []*sqlast.CommentGroup{
			{
				List: []*sqlast.Comment{
					{
						Text: "\ninsert\n",
						From: sqltoken.NewPos(5, 1),
						To:   sqltoken.NewPos(7, 3),
					},
				},
			},
		})
	})

	t.Run("create table", func(t *testing.T) {

		f := parseFile(t, `
/*associate with stmts1*/
CREATE TABLE test (
	/*associate with columndef*/
    col0 int primary key, --columndef
	/*with constraints*/
    col1 integer constraint test_constraint check (10 < col1 and col1 < 100),
    foreign key (col0, col1) references test2(col1, col2), --table constraints1
	--table constraints2
    CONSTRAINT test_constraint check(col1 > 10)
); --associate with stmts2
`)

		m := sqlast.NewCommentMap(f)
		ct := f.Stmts[0].(*sqlast.CreateTableStmt)
		compareComment(t, m[ct], []*sqlast.CommentGroup{
			{
				List: []*sqlast.Comment{
					{
						Text: "associate with stmts1",
						From: sqltoken.NewPos(2, 1),
						To:   sqltoken.NewPos(2, 26),
					},
				},
			},
			{
				List: []*sqlast.Comment{
					{
						Text: "associate with stmts2",
						From: sqltoken.NewPos(11, 4),
						To:   sqltoken.NewPos(11, 27),
					},
				},
			},
		})

		compareComment(t, m[ct.Elements[0]], []*sqlast.CommentGroup{
			{
				List: []*sqlast.Comment{
					{
						Text: "associate with columndef",
						From: sqltoken.NewPos(4, 5),
						To:   sqltoken.NewPos(4, 33),
					},
				},
			},
			{
				List: []*sqlast.Comment{
					{
						Text: "columndef",
						From: sqltoken.NewPos(5, 27),
						To:   sqltoken.NewPos(5, 38),
					},
				},
			},
		})

		compareComment(t, m[ct.Elements[1]], []*sqlast.CommentGroup{
			{
				List: []*sqlast.Comment{
					{
						Text: "with constraints",
						From: sqltoken.NewPos(6, 5),
						To:   sqltoken.NewPos(6, 25),
					},
				},
			},
		})

		compareComment(t, m[ct.Elements[2]], []*sqlast.CommentGroup{
			{
				List: []*sqlast.Comment{
					{
						Text: "table constraints1",
						From: sqltoken.NewPos(8, 60),
						To:   sqltoken.NewPos(8, 80),
					},
				},
			},
		})

		compareComment(t, m[ct.Elements[3]], []*sqlast.CommentGroup{
			{
				List: []*sqlast.Comment{
					{
						Text: "table constraints2",
						From: sqltoken.NewPos(9, 5),
						To:   sqltoken.NewPos(9, 25),
					},
				},
			},
		})
	})
}
