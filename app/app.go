package app

import (
	"github.com/ajrudzitis/ssh-resume/pty"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	log "github.com/sirupsen/logrus"
)

type ResumeApp struct {
	tviewApp *tview.Application
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
	r.tviewApp = app

	r.mainMenu()

	if err := app.Run(); err != nil {
		log.Errorf("app: failed to run application: %v", err)
	}

}

func (r *ResumeApp) mainMenu() {
	menu := tview.NewList().
		AddItem("Resume", "", 'a', func() {
			r.resume()
		}).
		AddItem("About", "", 'b', func() {
			r.about()
		}).
		AddItem("Quit", "", 'q', func() {
			r.tviewApp.Stop()
		})

	window := tview.NewFrame(menu).
		SetBorders(2, 2, 2, 2, 4, 4).
		AddText("Aleks's Resume", true, tview.AlignCenter, tcell.ColorWhite)

	r.tviewApp.SetRoot(window, true).SetFocus(menu)
}

func (r *ResumeApp) resume() {
	window := tview.NewBox().SetBorder(true).SetTitle("Resume")
	r.tviewApp.SetRoot(window, true).SetFocus(window)
}

func (r *ResumeApp) about() {

}
