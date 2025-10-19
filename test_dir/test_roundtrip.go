package main

import (
	"fmt"
	"log"

	"github.com/moomou/xsqlparser"
)

func main() {
	sql := `CREATE VIRTUAL TABLE IF NOT EXISTS "conversation_fts" USING fts5(id, text, prefix = "2", prefix = "3")`

	stmt, err := xsqlparser.Parse(sql)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Original SQL:")
	fmt.Println(sql)
	fmt.Println("\nParsed SQL back to string:")
	fmt.Println(stmt.ToSQLString())
}