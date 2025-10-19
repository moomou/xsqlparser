package main

import (
	"bytes"
	"log"

	"github.com/k0kubun/pp"

	"github.com/moomou/xsqlparser"
	"github.com/moomou/xsqlparser/dialect"
	"github.com/moomou/xsqlparser/sqlast"
)

func main() {
	simpleSelect()
	complicatedSelect()
	withCTE()
	createASTList()
	commentMap()
	testVirtualTable()
}

func simpleSelect() {
	str := "SELECT * from test_table"
	parser, err := xsqlparser.NewParser(bytes.NewBufferString(str), &dialect.GenericSQLDialect{})
	if err != nil {
		log.Fatal(err)
	}

	stmt, err := parser.ParseStatement()
	if err != nil {
		log.Fatal(err)
	}
	pp.Println(stmt)

	log.Println(stmt.ToSQLString())
}

func complicatedSelect() {
	str := "SELECT orders.product, SUM(orders.quantity) AS product_units, accounts.* " +
		"FROM orders LEFT JOIN accounts ON orders.account_id = accounts.id " +
		"WHERE orders.region IN (SELECT region FROM top_regions) " +
		"ORDER BY product_units LIMIT 100"

	parser, err := xsqlparser.NewParser(bytes.NewBufferString(str), &dialect.GenericSQLDialect{})
	if err != nil {
		log.Fatal(err)
	}

	stmt, err := parser.ParseStatement()
	if err != nil {
		log.Fatal(err)
	}
	pp.Println(stmt)

	log.Println(stmt.ToSQLString())
}

func withCTE() {
	str := "WITH regional_sales AS (" +
		"SELECT region, SUM(amount) AS total_sales " +
		"FROM orders GROUP BY region) " +
		"SELECT product, SUM(quantity) AS product_units " +
		"FROM orders " +
		"WHERE region IN (SELECT region FROM top_regions) " +
		"GROUP BY region, product"

	parser, err := xsqlparser.NewParser(bytes.NewBufferString(str), &dialect.GenericSQLDialect{})
	if err != nil {
		log.Fatal(err)
	}

	stmt, err := parser.ParseStatement()
	if err != nil {
		log.Fatal(err)
	}
	pp.Println(stmt)

	log.Println(stmt.ToSQLString())
}

func createASTList() {
	src := `WITH regional_sales AS (
		SELECT region, SUM(amount) AS total_sales
		FROM orders GROUP BY region)
		SELECT product, SUM(quantity) AS product_units
		FROM orders
		WHERE region IN (SELECT region FROM top_regions)
		GROUP BY region, product;`

	parser, err := xsqlparser.NewParser(bytes.NewBufferString(src), &dialect.GenericSQLDialect{})
	if err != nil {
		log.Fatal(err)
	}

	stmt, err := parser.ParseStatement()
	if err != nil {
		log.Fatal(err)
	}
	var list []sqlast.Node

	sqlast.Inspect(stmt, func(node sqlast.Node) bool {
		switch node.(type) {
		case nil:
			return false
		default:
			list = append(list, node)
			return true
		}
	})

	pp.Println(list)
}

func commentMap() {

	src := `
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
`

	parser, err := xsqlparser.NewParser(bytes.NewBufferString(src), &dialect.GenericSQLDialect{}, xsqlparser.ParseComment())
	if err != nil {
		log.Fatal(err)
	}

	file, err := parser.ParseFile()
	if err != nil {
		log.Fatal(err)
	}

	m := sqlast.NewCommentMap(file)

	createTable := file.Stmts[0].(*sqlast.CreateTableStmt)

	pp.Println(m[createTable.Elements[0]]) // you can show `associate with columndef` and `columndef` comments
}

func testVirtualTable() {
	str := `CREATE VIRTUAL TABLE IF NOT EXISTS "conversation_fts" USING fts5(id, text, prefix = "2", prefix = "3")`

	parser, err := xsqlparser.NewParser(bytes.NewBufferString(str), &dialect.GenericSQLDialect{})
	if err != nil {
		log.Fatal(err)
	}

	stmt, err := parser.ParseStatement()
	if err != nil {
		log.Fatal(err)
	}

	// Type assert to check if it's a CreateVirtualTableStmt
	virtualTableStmt, ok := stmt.(*sqlast.CreateVirtualTableStmt)
	if !ok {
		log.Fatal("Expected CreateVirtualTableStmt")
	}

	pp.Println(virtualTableStmt)
	log.Printf("NotExists: %v", virtualTableStmt.NotExists)
	log.Printf("Table Name: %s", virtualTableStmt.Name)
	log.Printf("Using: %s", virtualTableStmt.Using)
	log.Printf("Arguments: %v", virtualTableStmt.Arguments)

	log.Println("Round-trip SQL:")
	log.Println(virtualTableStmt.ToSQLString())
}
