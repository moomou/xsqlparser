package sqltoken

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/moomou/xsqlparser/dialect"
)

func TestTokenizer_Tokenize(t *testing.T) {
	cases := []struct {
		name string
		in   string
		out  []*Token
	}{
		{
			name: "whitespace",
			in:   " ",
			out: []*Token{
				{
					Kind:  Whitespace,
					Value: " ",
					From:  Pos{Line: 1, Col: 1},
					To:    Pos{Line: 1, Col: 2},
				},
			},
		},
		{
			name: "whitespace and new line",
			in: `
 `,
			out: []*Token{
				{
					Kind:  Whitespace,
					Value: "\n",
					From:  Pos{Line: 1, Col: 1},
					To:    Pos{Line: 2, Col: 1},
				},
				{
					Kind:  Whitespace,
					Value: " ",
					From:  Pos{Line: 2, Col: 1},
					To:    Pos{Line: 2, Col: 2},
				},
			},
		},
		{
			name: "whitespace and tab",
			in:   "\r\n	",
			out: []*Token{
				{
					Kind:  Whitespace,
					Value: "\n",
					From:  Pos{Line: 1, Col: 1},
					To:    Pos{Line: 2, Col: 1},
				},
				{
					Kind:  Whitespace,
					Value: "\t",
					From:  Pos{Line: 2, Col: 1},
					To:    Pos{Line: 2, Col: 5},
				},
			},
		},
		{
			name: "N string",
			in:   "N'string'",
			out: []*Token{
				{
					Kind:  NationalStringLiteral,
					Value: "string",
					From:  Pos{Line: 1, Col: 1},
					To:    Pos{Line: 1, Col: 10},
				},
			},
		},
		{
			name: "N string with keyword",
			in:   "N'string' NOT",
			out: []*Token{
				{
					Kind:  NationalStringLiteral,
					Value: "string",
					From:  Pos{Line: 1, Col: 1},
					To:    Pos{Line: 1, Col: 10},
				},
				{
					Kind:  Whitespace,
					Value: " ",
					From:  Pos{Line: 1, Col: 10},
					To:    Pos{Line: 1, Col: 11},
				},
				{
					Kind: SQLKeyword,
					Value: &SQLWord{
						Value:   "NOT",
						Keyword: "NOT",
					},
					From: Pos{Line: 1, Col: 11},
					To:   Pos{Line: 1, Col: 14},
				},
			},
		},
		{
			name: "Ident",
			in:   "select",
			out: []*Token{
				{
					Kind: SQLKeyword,
					Value: &SQLWord{
						Value:   "select",
						Keyword: "SELECT",
					},
					From: Pos{Line: 1, Col: 1},
					To:   Pos{Line: 1, Col: 7},
				},
			},
		},
		{
			name: "single quote string",
			in:   "'test'",
			out: []*Token{
				{
					Kind:  SingleQuotedString,
					Value: "test",
					From:  Pos{Line: 1, Col: 1},
					To:    Pos{Line: 1, Col: 7},
				},
			},
		},
		{
			name: "quoted string",
			in:   "\"SELECT\"",
			out: []*Token{
				{
					Kind: SQLKeyword,
					Value: &SQLWord{
						Value:      "SELECT",
						Keyword:    "SELECT",
						QuoteStyle: '"',
					},
					From: Pos{Line: 1, Col: 1},
					To:   Pos{Line: 1, Col: 9},
				},
			},
		},
		{
			name: "parents with number",
			in:   "(123),",
			out: []*Token{
				{
					Kind:  LParen,
					Value: "(",
					From:  Pos{Line: 1, Col: 1},
					To:    Pos{Line: 1, Col: 2},
				},
				{
					Kind:  Number,
					Value: "123",
					From:  Pos{Line: 1, Col: 2},
					To:    Pos{Line: 1, Col: 5},
				},
				{
					Kind:  RParen,
					Value: ")",
					From:  Pos{Line: 1, Col: 5},
					To:    Pos{Line: 1, Col: 6},
				},
				{
					Kind:  Comma,
					Value: ",",
					From:  Pos{Line: 1, Col: 6},
					To:    Pos{Line: 1, Col: 7},
				},
			},
		},
		{
			name: "minus comment",
			in:   "-- test",
			out: []*Token{
				{
					Kind:  Comment,
					Value: " test",
					From:  Pos{Line: 1, Col: 1},
					To:    Pos{Line: 1, Col: 8},
				},
			},
		},
		{
			name: "minus operator",
			in:   "1-3",
			out: []*Token{
				{
					Kind:  Number,
					Value: "1",
					From:  Pos{Line: 1, Col: 1},
					To:    Pos{Line: 1, Col: 2},
				},
				{
					Kind:  Minus,
					Value: "-",
					From:  Pos{Line: 1, Col: 2},
					To:    Pos{Line: 1, Col: 3},
				},
				{
					Kind:  Number,
					Value: "3",
					From:  Pos{Line: 1, Col: 3},
					To:    Pos{Line: 1, Col: 4},
				},
			},
		},
		{
			name: "/* comment",
			in: `/* test
multiline
comment */`,
			out: []*Token{
				{
					Kind:  Comment,
					Value: " test\nmultiline\ncomment ",
					From:  Pos{Line: 1, Col: 1},
					To:    Pos{Line: 3, Col: 11},
				},
			},
		},
		{
			name: "operators",
			in:   "1/1*1+1%1=1.1-.",
			out: []*Token{
				{
					Kind:  Number,
					Value: "1",
					From:  Pos{Line: 1, Col: 1},
					To:    Pos{Line: 1, Col: 2},
				},
				{
					Kind:  Div,
					Value: "/",
					From:  Pos{Line: 1, Col: 2},
					To:    Pos{Line: 1, Col: 3},
				},
				{
					Kind:  Number,
					Value: "1",
					From:  Pos{Line: 1, Col: 3},
					To:    Pos{Line: 1, Col: 4},
				},
				{
					Kind:  Mult,
					Value: "*",
					From:  Pos{Line: 1, Col: 4},
					To:    Pos{Line: 1, Col: 5},
				},
				{
					Kind:  Number,
					Value: "1",
					From:  Pos{Line: 1, Col: 5},
					To:    Pos{Line: 1, Col: 6},
				},
				{
					Kind:  Plus,
					Value: "+",
					From:  Pos{Line: 1, Col: 6},
					To:    Pos{Line: 1, Col: 7},
				},
				{
					Kind:  Number,
					Value: "1",
					From:  Pos{Line: 1, Col: 7},
					To:    Pos{Line: 1, Col: 8},
				},
				{
					Kind:  Mod,
					Value: "%",
					From:  Pos{Line: 1, Col: 8},
					To:    Pos{Line: 1, Col: 9},
				},
				{
					Kind:  Number,
					Value: "1",
					From:  Pos{Line: 1, Col: 9},
					To:    Pos{Line: 1, Col: 10},
				},
				{
					Kind:  Eq,
					Value: "=",
					From:  Pos{Line: 1, Col: 10},
					To:    Pos{Line: 1, Col: 11},
				},
				{
					Kind:  Number,
					Value: "1.1",
					From:  Pos{Line: 1, Col: 11},
					To:    Pos{Line: 1, Col: 14},
				},
				{
					Kind:  Minus,
					Value: "-",
					From:  Pos{Line: 1, Col: 14},
					To:    Pos{Line: 1, Col: 15},
				},
				{
					Kind:  Period,
					Value: ".",
					From:  Pos{Line: 1, Col: 15},
					To:    Pos{Line: 1, Col: 16},
				},
			},
		},
		{
			name: "Neq",
			in:   "1!=2",
			out: []*Token{
				{
					Kind:  Number,
					Value: "1",
					From:  Pos{Line: 1, Col: 1},
					To:    Pos{Line: 1, Col: 2},
				},
				{
					Kind:  Neq,
					Value: "!=",
					From:  Pos{Line: 1, Col: 2},
					To:    Pos{Line: 1, Col: 4},
				},
				{
					Kind:  Number,
					Value: "2",
					From:  Pos{Line: 1, Col: 4},
					To:    Pos{Line: 1, Col: 5},
				},
			},
		},
		{
			name: "Lts",
			in:   "<<=<>",
			out: []*Token{
				{
					Kind:  Lt,
					Value: "<",
					From:  Pos{Line: 1, Col: 1},
					To:    Pos{Line: 1, Col: 2},
				},
				{
					Kind:  LtEq,
					Value: "<=",
					From:  Pos{Line: 1, Col: 2},
					To:    Pos{Line: 1, Col: 4},
				},
				{
					Kind:  Neq,
					Value: "<>",
					From:  Pos{Line: 1, Col: 4},
					To:    Pos{Line: 1, Col: 6},
				},
			},
		},
		{
			name: "Gts",
			in:   ">>=",
			out: []*Token{
				{
					Kind:  Gt,
					Value: ">",
					From:  Pos{Line: 1, Col: 1},
					To:    Pos{Line: 1, Col: 2},
				},
				{
					Kind:  GtEq,
					Value: ">=",
					From:  Pos{Line: 1, Col: 2},
					To:    Pos{Line: 1, Col: 4},
				},
			},
		},
		{
			name: "colons",
			in:   ":1::1;",
			out: []*Token{
				{
					Kind:  Colon,
					Value: ":",
					From:  Pos{Line: 1, Col: 1},
					To:    Pos{Line: 1, Col: 2},
				},
				{
					Kind:  Number,
					Value: "1",
					From:  Pos{Line: 1, Col: 2},
					To:    Pos{Line: 1, Col: 3},
				},
				{
					Kind:  DoubleColon,
					Value: "::",
					From:  Pos{Line: 1, Col: 3},
					To:    Pos{Line: 1, Col: 5},
				},
				{
					Kind:  Number,
					Value: "1",
					From:  Pos{Line: 1, Col: 5},
					To:    Pos{Line: 1, Col: 6},
				},
				{
					Kind:  Semicolon,
					Value: ";",
					From:  Pos{Line: 1, Col: 6},
					To:    Pos{Line: 1, Col: 7},
				},
			},
		},
		{
			name: "others",
			in:   "\\[{&}]",
			out: []*Token{
				{
					Kind:  Backslash,
					Value: "\\",
					From:  Pos{Line: 1, Col: 1},
					To:    Pos{Line: 1, Col: 2},
				},
				{
					Kind:  LBracket,
					Value: "[",
					From:  Pos{Line: 1, Col: 2},
					To:    Pos{Line: 1, Col: 3},
				},
				{
					Kind:  LBrace,
					Value: "{",
					From:  Pos{Line: 1, Col: 3},
					To:    Pos{Line: 1, Col: 4},
				},
				{
					Kind:  Ampersand,
					Value: "&",
					From:  Pos{Line: 1, Col: 4},
					To:    Pos{Line: 1, Col: 5},
				},
				{
					Kind:  RBrace,
					Value: "}",
					From:  Pos{Line: 1, Col: 5},
					To:    Pos{Line: 1, Col: 6},
				},
				{
					Kind:  RBracket,
					Value: "]",
					From:  Pos{Line: 1, Col: 6},
					To:    Pos{Line: 1, Col: 7},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			src := strings.NewReader(c.in)
			tokenizer := NewTokenizer(src, &dialect.GenericSQLDialect{})

			tok, err := tokenizer.Tokenize()
			if err != nil {
				t.Errorf("should be no error %v", err)
			}

			if len(tok) != len(c.out) {
				t.Fatalf("should be same length but %d, %d", len(tok), len(c.out))
			}

			for i := 0; i < len(tok); i++ {
				if tok[i].Kind != c.out[i].Kind {
					t.Errorf("%d, expected sqltoken: %d, but got %d", i, c.out[i].Kind, tok[i].Kind)
				}
				if !reflect.DeepEqual(tok[i].Value, c.out[i].Value) {
					t.Errorf("%d, expected value: %+v, but got %+v", i, c.out[i].Value, tok[i].Value)
				}
				if !reflect.DeepEqual(tok[i].From, c.out[i].From) {
					t.Errorf("%d, expected value: %+v, but got %+v", i, c.out[i].From, tok[i].From)
				}
				if !reflect.DeepEqual(tok[i].To, c.out[i].To) {
					t.Errorf("%d, expected value: %+v, but got %+v", i, c.out[i].To, tok[i].To)
				}
			}
		})
	}
}

func TestTokenizer_Pos(t *testing.T) {
	t.Run("operators", func(t *testing.T) {
		cases := []struct {
			operator string
			add      int
		}{
			{
				operator: "+",
			},
			{
				operator: "-",
			},
			{
				operator: "%",
			},
			{
				operator: "*",
			},
			{
				operator: "/",
			},
			{
				operator: ">",
			},
			{
				operator: "=",
			},
			{
				operator: "<",
			},
			{
				operator: "<=",
				add:      1,
			},
			{
				operator: "<>",
				add:      1,
			},
			{
				operator: ">=",
				add:      1,
			},
		}

		for _, c := range cases {
			t.Run(c.operator, func(t *testing.T) {
				src := fmt.Sprintf("1 %s 1", c.operator)
				tokenizer := NewTokenizer(bytes.NewBufferString(src), &dialect.GenericSQLDialect{})

				if _, err := tokenizer.Tokenize(); err != nil {
					t.Fatal(err)
				}

				if d := cmp.Diff(tokenizer.Pos(), Pos{Line: 1, Col: 6 + c.add}); d != "" {
					t.Errorf("must be same but diff: %s", d)
				}
			})
		}
	})
	t.Run("other expressions", func(t *testing.T) {
		cases := []struct {
			name   string
			src    string
			expect Pos
		}{
			{
				name: "multiline ",
				src: `1+1
asdf`,
				expect: Pos{Line: 2, Col: 5},
			},
			{
				name:   "single line comment",
				src:    `-- comments`,
				expect: Pos{Line: 1, Col: 12},
			},
			{
				name:   "statements",
				src:    `select count(id) from account`,
				expect: Pos{Line: 1, Col: 30},
			},
			{
				name: "multiline statements",
				src: `select count(id)
from account 
where name like '%test%'`,
				expect: Pos{Line: 3, Col: 25},
			},
			{
				name: "multiline comment",
				src: `/*
test comment
test comment
*/`,
				expect: Pos{Line: 4, Col: 3},
			},
			{
				name:   "single line comment",
				src:    "/* asdf */",
				expect: Pos{Line: 1, Col: 11},
			},
			{
				name:   "comment inside sql",
				src:    "select * from /* test table */ test_table where id != 123",
				expect: Pos{Line: 1, Col: 58},
			},
		}

		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				tokenizer := NewTokenizer(bytes.NewBufferString(c.src), &dialect.GenericSQLDialect{})

				if _, err := tokenizer.Tokenize(); err != nil {
					t.Fatal(err)
				}

				if d := cmp.Diff(tokenizer.Pos(), c.expect); d != "" {
					t.Errorf("must be same but diff: %s", d)
				}
			})
		}
	})

	t.Run("illegal cases", func(t *testing.T) {
		cases := []struct {
			name string
			src  string
		}{
			{
				name: "incomplete quoted string",
				src:  "'test",
			},
			{
				name: "unclosed multiline comment",
				src: `
/* test
test
`,
			},
		}

		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				tokenizer := NewTokenizer(bytes.NewBufferString(c.src), &dialect.GenericSQLDialect{})

				_, err := tokenizer.Tokenize()
				if err == nil {
					t.Errorf("must be error but blank")
				}
				t.Logf("%+v", err)

			})
		}
	})
}

func BenchmarkTokenizer_Tokenize(b *testing.B) {
	cases := []struct {
		name string
		src  string
	}{
		{
			name: "select",
			src: `SELECT COUNT(customer_id), country 
FROM customers 
GROUP BY country 
HAVING COUNT(customer_id) > 3`,
		},
		{
			name: "complex select",
			src: `SELECT start_terminal,
       start_time,
       duration_seconds,
       ROW_NUMBER() OVER (ORDER BY start_time)
                    AS row_number
  FROM tutorial.dc_bikeshare_q1_2012
 WHERE start_time < '2012-01-08'`,
		},
		{
			name: "insert",
			src:  `INSERT INTO tbl_name (a,b,c) VALUES(1,2,3),(4,5,6),(7,8,9);`,
		},

		{

			name: "multi line comment",
			src: `
create table account (
    account_id serial primary key,  --aaa
	/*bbb*/
    name varchar(255) not null,
    email /*ccc*/ varchar(255) unique not null --ddd
);

--eee

/*fff
ggg
*/
select 1 from test; --hhh
/*jjj*/ --kkk
select 1 from test; /*lll*/ --mmm
--nnn
`,
		},
	}

	for _, c := range cases {
		b.Run(c.name, func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				in := bytes.NewBufferString(c.src)
				tokenizer := NewTokenizer(in, &dialect.GenericSQLDialect{})

				if _, err := tokenizer.Tokenize(); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkTokenizer_Tokenize_WithoutComment(b *testing.B) {
	cases := []struct {
		name string
		src  string
	}{
		{
			name: "select",
			src: `SELECT COUNT(customer_id), country 
FROM customers 
GROUP BY country 
HAVING COUNT(customer_id) > 3`,
		},
		{
			name: "complex select",
			src: `SELECT start_terminal,
       start_time,
       duration_seconds,
       ROW_NUMBER() OVER (ORDER BY start_time)
                    AS row_number
  FROM tutorial.dc_bikeshare_q1_2012
 WHERE start_time < '2012-01-08'`,
		},
		{
			name: "insert",
			src:  `INSERT INTO tbl_name (a,b,c) VALUES(1,2,3),(4,5,6),(7,8,9);`,
		},
	}

	for _, c := range cases {
		b.Run(c.name, func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				in := bytes.NewBufferString(c.src)
				tokenizer := NewTokenizerWithOptions(in, Dialect(&dialect.GenericSQLDialect{}), DisableParseComment())

				if _, err := tokenizer.Tokenize(); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
