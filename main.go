package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	_ "github.com/duckdb/duckdb-go/v2"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

type InputEvent struct {
	Query string `json:"query"`
}

func hello(ctx context.Context, event InputEvent) (string, error) {
	db, err := sql.Open("duckdb", "")
	if err != nil {
		fmt.Println("Error opening database connection:", err)
		return "", err
	}
	defer db.Close()

	query := event.Query

	// auto install and load extensions
	db.Exec("SET autoinstall_known_extensions=1;")
	db.Exec("SET autoload_known_extensions=1;")

	for _, querySegment := range strings.Split(query, ";") {
		if querySegment == "" {
			continue
		}
		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.Style().Format.Header = text.FormatDefault
		rows, err := db.Query(querySegment)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		columns, err := rows.Columns()
		if err != nil {
			log.Fatal(err)
		}

		tr := table.Row{}
		for _, col := range columns {
			tr = append(tr, fmt.Sprint(col))
		}
		t.AppendHeader(tr)

		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		var tableDataRows []table.Row = []table.Row{}

		for rows.Next() {
			tableRowCols := table.Row{}
			err = rows.Scan(valuePtrs...)
			if err != nil {
				log.Fatal(err)
			}
			for i := range columns {
				val := *(valuePtrs[i].(*interface{}))
				tableRowCols = append(tableRowCols, fmt.Sprint(val))
			}
			tableDataRows = append(tableDataRows, tableRowCols)
		}
		t.AppendRows(tableDataRows)

		if err := rows.Err(); err != nil {
			log.Fatal(err)
		}
		t.Render()
	}

	return "See logs below for output", nil
}

func main() {
	lambda.Start(hello)
}
