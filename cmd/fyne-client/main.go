package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type AppState struct {
	currentPage string
	window      fyne.Window
}

func main() {
	// Инициализация приложения и окна
	a := app.New()
	a.Settings().SetTheme(theme.LightTheme())
	w := a.NewWindow("Rubric Grader")
	state := &AppState{currentPage: "greeting", window: w}

	// Установка начального содержимого
	w.SetContent(createContent(state))

	// Установка размера окна
	w.Resize(fyne.NewSize(1280, 720))
	w.ShowAndRun()
}

// createContent создает содержимое окна в зависимости от текущей страницы
func createContent(state *AppState) fyne.CanvasObject {
	leftBackground := canvas.NewImageFromFile("bin/logo/hse_logo.svg")
	leftBackground.FillMode = canvas.ImageFillStretch

	switch state.currentPage {
	//authorization
	case "greeting":
		return CreateGreetingPage(state, leftBackground)
	case "authorization":
		return CreateAuthorizationPage(state, leftBackground)
	case "registration":
		return CreateRegistrationPage(state, leftBackground)
	//superacc
	case "superacc-groups":
		return СreateGroupListPage(state, leftBackground)
	case "superacc-users-of-group":
		return СreateGroupUsersPage(state, leftBackground, GroupName)
	case "superacc-all-users":
		return СreateUsersListPage(state, leftBackground)
	default:
		return container.NewVBox(widget.NewLabel("Unknown page"))
	}
}
