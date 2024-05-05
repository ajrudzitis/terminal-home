package app

import (
	"github.com/ajrudzitis/ssh-resume/pty"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	log "github.com/sirupsen/logrus"
)

type ResumeApp struct {
}

func (r *ResumeApp) Run(pty *pty.Pty) {
	// create a new terminfo screen
	screen, err := tcell.NewTerminfoScreenFromTty(pty)
	if err != nil {
		log.Errorf("app: failed to create terminfo screen: %v", err)
		return
	}

	// create a new application and set the screen
	app := tview.NewApplication()
	app.SetScreen(screen)

	menu := tview.NewList().
		AddItem("List item 1", "Some explanatory text", 'a', nil).
		AddItem("List item 2", "Some explanatory text", 'b', nil).
		AddItem("List item 3", "Some explanatory text", 'c', nil).
		AddItem("List item 4", "Some explanatory text", 'd', nil).
		AddItem("Quit", "Press to exit", 'q', func() {
			app.Stop()
		})

	window := tview.NewFrame(menu).
		SetBorders(2, 2, 2, 2, 4, 4).
		AddText("Aleks's Resume", true, tview.AlignCenter, tcell.ColorWhite)

	if err := app.SetRoot(window, true).SetFocus(menu).Run(); err != nil {
		panic(err)
	}

}
