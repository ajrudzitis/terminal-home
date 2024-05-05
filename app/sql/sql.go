package sql

import (
	"bufio"
	"database/sql"
	"fmt"
	"strings"

	_ "embed"

	"github.com/gdamore/tcell/v2"
	_ "github.com/glebarez/go-sqlite"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rivo/tview"
	log "github.com/sirupsen/logrus"
)

//go:embed resources/init.sql
var initalSql string

func SqlGameView(app *tview.Application, quitFn func()) {
	// set up a new database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		log.Errorf("failed to open database: %v", err)
		quitFn()
		return
	}

	// load database from init.sql
	initDb(db)

	// render a simulated shell
	terminalView := tview.NewTextView().SetChangedFunc(func() {
		app.Draw()
	})
	// TODO add some type of auto scroll
	terminalView.SetScrollable(false)

	inputView := tview.NewInputField().SetLabel("sql> ").SetFieldWidth(0).SetFieldBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
	inputView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			// get the input
			input := inputView.GetText()

			if input == "exit" || input == "quit" {
				quitFn()
				return nil
			}

			// clear the input
			inputView.SetText("")

			// record the query
			fmt.Fprintf(terminalView, "> %s\n", input)

			// hack: make a nice query for showing tables
			if input == ".tables" || input == "show tables" {
				input = "SELECT * FROM sqlite_master WHERE type='table';"
			}

			// run the query
			rows, err := db.Query(input)
			if err != nil {
				fmt.Fprintf(terminalView, "Error: %v\n", err)
				return nil
			}
			defer rows.Close()

			// print the rows
			t := table.NewWriter()
			t.SetOutputMirror(terminalView)

			firstRow := true
			for rows.Next() {
				columns, err := rows.Columns()
				if firstRow {

					if err != nil {
						fmt.Fprintf(terminalView, "Error: %v\n", err)
						return nil
					}
					row := table.Row{}
					for _, column := range columns {
						row = append(row, column)
					}
					t.AppendHeader(row)
					firstRow = false
				}

				values := make([]interface{}, len(columns))
				valuePtrs := make([]interface{}, len(columns))
				for i := range columns {
					valuePtrs[i] = &values[i]
				}
				rows.Scan(valuePtrs...)
				row := table.Row{}
				for _, value := range values {
					row = append(row, value)
				}
				t.AppendRow(row)
			}
			t.Render()

			// if there was no row to print, print the number of rows affected
			if firstRow {
				fmt.Fprint(terminalView, "Empty result\n")
			}

		}
		return event
	})

	fmt.Fprint(terminalView, "Welcome to my experience database!\nType 'exit' or 'quit' to quit.\n\n")
	fmt.Fprint(terminalView, "Hint: try 'show tables'\n\n")

	mainView := tview.NewGrid().
		SetRows(0, 1, 1).
		AddItem(terminalView, 0, 0, 1, 1, 0, 0, false).
		AddItem(inputView, 2, 0, 1, 1, 0, 0, true)

	app.SetRoot(mainView, true).SetFocus(inputView)

}

func initDb(db *sql.DB) {
	// split initialsql into statements by line
	scanner := bufio.NewScanner(strings.NewReader(initalSql))
	for scanner.Scan() {
		statement := scanner.Text()
		_, err := db.Exec(statement)
		if err != nil {
			log.Errorf("sql: failed to execute statement: %v", err)
		}
	}

}
