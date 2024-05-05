package app

import (
	_ "embed"
	"fmt"

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
	textView.SetBorder(true).SetTitle(title)
	r.tviewApp.SetRoot(textView, true).SetFocus(textView)
}
