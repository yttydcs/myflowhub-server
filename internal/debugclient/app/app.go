package app

import (
	"context"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"github.com/yttydcs/myflowhub-server/internal/debugclient/ui"
	"github.com/yttydcs/myflowhub-server/internal/debugclient/ui/theme"
)

func Run() {
	fyneApp := app.NewWithID("myflowhub.debugclient")
	theme.Apply(fyneApp)
	window := fyneApp.NewWindow("MyFlowHub Debug Client")

	ctx, cancel := context.WithCancel(context.Background())
	controller := ui.New(fyneApp, ctx)
	window.SetContent(controller.Build(window))
	window.Resize(fyne.NewSize(900, 640))

	window.SetCloseIntercept(func() {
		controller.Shutdown()
		cancel()
		time.Sleep(120 * time.Millisecond)
		window.Close()
	})

	window.ShowAndRun()
}
