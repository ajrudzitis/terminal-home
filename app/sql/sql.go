package sql

import (
	"database/sql"
	"fmt"

	"github.com/gdamore/tcell/v2"
	_ "github.com/proullon/ramsql/driver"
	"github.com/rivo/tview"
	log "github.com/sirupsen/logrus"
)

func SqlGameView(app *tview.Application, quitFn func()) {
	// set up a new database
	db, err := sql.Open("ramsql", "Experience")
	if err != nil {
		log.Errorf("failed to open database: %v", err)
		quitFn()
		return
	}

	// render a simulated shell
	terminalView := tview.NewTextView().SetChangedFunc(func() {
		app.Draw()
	})

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
			// run the query
			rows, err := db.Query(input)
			if err != nil {
				fmt.Fprintf(terminalView, "Error: %v\n", err)
				return nil
			}
			// print the rows
			for rows.Next() {
				columns, err := rows.Columns()
				if err != nil {
					fmt.Fprintf(terminalView, "Error: %v\n", err)
					return nil
				}
				values := make([]interface{}, len(columns))
				valuePtrs := make([]interface{}, len(columns))
				for i := range columns {
					valuePtrs[i] = &values[i]
				}
				rows.Scan(valuePtrs...)
				for i, col := range columns {
					fmt.Fprintf(terminalView, "%s: %s\n", col, values[i])
				}
			}
		}
		return event
	})

	fmt.Fprint(terminalView, "Welcome to my experience database!\nType 'exit' or 'quit' to quit.\n")

	mainView := tview.NewGrid().
		SetRows(0, 1, 1).
		AddItem(terminalView, 0, 0, 1, 1, 0, 0, false).
		AddItem(inputView, 2, 0, 1, 1, 0, 0, true)

	app.SetRoot(mainView, true).SetFocus(inputView)

}
