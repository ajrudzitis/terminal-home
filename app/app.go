package app

import (
	"github.com/ajrudzitis/ssh-resume/pty"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	log "github.com/sirupsen/logrus"
)

func Run(pty *pty.Pty) {
	screen, err := tcell.NewTerminfoScreenFromTty(pty)
	if err != nil {
		log.Errorf("app: failed to create terminfo screen: %v", err)
		return
	}

	app := tview.NewApplication()
	app.SetScreen(screen)
	form := tview.NewForm().
		AddInputField("Enter Card Number", "", 16, nil, nil).
		AddButton("Submit", nil).
		AddButton("Quit", func() {
			app.Stop()
		})
	form.SetBorder(true).SetTitle("Enter Payment Details").SetTitleAlign(tview.AlignLeft)
	if err := app.SetRoot(form, true).Run(); err != nil {
		log.Errorf("failed to run app: %v", err)
	}

}
