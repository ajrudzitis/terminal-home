package app

import (
	_ "embed"
	"fmt"

	"github.com/ajrudzitis/terminal-resume/pty"
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
	menu.SetBorder(true).SetTitle("Main Menu")

	mainView := tview.NewGrid().
		SetRows(0, 0, 0).
		SetColumns(0, 0, 0).
		AddItem(menu, 1, 1, 1, 1, 0, 0, true)

	r.tviewApp.SetRoot(mainView, true).SetFocus(menu)
}

//go:embed resources/resume.txt
var resumeContent string

func (r *ResumeApp) resume() {
	r.textView("Resume", resumeContent)
}

//go:embed resources/about.txt
var aboutContent string

func (r *ResumeApp) about() {
	r.textView("About", aboutContent)
}

func (r *ResumeApp) textView(title, content string) {
	textView := tview.NewTextView().SetChangedFunc(func() {
		r.tviewApp.Draw()
	})
	fmt.Fprint(textView, content)
	textView.SetBorder(true).SetTitle(title).SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRune && event.Rune() == 'q' {
			r.mainMenu()
		}
		return event

	})

	helpInfo := tview.NewTextView().
		SetText("Press 'q' to return to the main menu")

	mainView := tview.NewGrid().
		SetRows(0, 1).
		AddItem(textView, 0, 0, 1, 1, 0, 0, true).
		AddItem(helpInfo, 1, 0, 1, 1, 0, 0, false)
	r.tviewApp.SetRoot(mainView, true).SetFocus(textView)
}
