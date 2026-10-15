// Command sheets demonstrates reading and writing a Google Sheet via the
// Sheets API.
//
// Usage:
//
//	go run ./cmd/sheets <spreadsheetID>
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"google-scripts/go/internal/auth"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("usage: %s <spreadsheetID>", os.Args[0])
	}
	spreadsheetID := os.Args[1]

	ctx := context.Background()
	client, err := auth.GetClient(ctx, "credentials.json", "token.json", sheets.SpreadsheetsScope)
	if err != nil {
		log.Fatalf("auth: %v", err)
	}

	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("creating sheets client: %v", err)
	}

	const readRange = "A1:B10"
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err != nil {
		log.Fatalf("reading values: %v", err)
	}

	fmt.Println("既存の値:")
	for _, row := range resp.Values {
		fmt.Println(row)
	}

	writeRange := "A1:B2"
	values := &sheets.ValueRange{
		Values: [][]interface{}{
			{"name", "score"},
			{"Alice", 90},
		},
	}
	_, err = srv.Spreadsheets.Values.Update(spreadsheetID, writeRange, values).
		ValueInputOption("RAW").
		Do()
	if err != nil {
		log.Fatalf("writing values: %v", err)
	}
	fmt.Println("書き込み完了")
}
