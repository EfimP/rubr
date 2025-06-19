package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type AppState struct {
	currentPage string
	userID      string
	role        string
	window      fyne.Window
}

func main() {
	a := app.New()
	a.Settings().SetTheme(theme.LightTheme())
	w := a.NewWindow("Rubric Grader")

	state := &AppState{
		currentPage: "greeting",
		userID:      "",
		role:        "",
		window:      w,
	}

	w.SetContent(createContent(state))
	w.Resize(fyne.NewSize(1280, 720))
	w.ShowAndRun()
}

func createContent(state *AppState) fyne.CanvasObject {
	switch state.currentPage {
	//authorization
	case "greeting":
		return CreateGreetingPage(state)
	case "authorization":
		return CreateAuthorizationPage(state)
	case "registration":
		return CreateRegistrationPage(state)
	//superacc
	case "superacc-groups":
		return СreateGroupListPage(state)
	case "superacc-users-of-group":
		return СreateGroupUsersPage(state, GroupName)
	case "superacc-all-users":
		return СreateUsersListPage(state)
		// lector
	case "lector_works":
		return CreateLectorWorksPage(state)
	case "assistant_works":
		return CreateAssistantWorksPage(state)
	case "seminarist_works":
		return CreateSeminaristWorksPage(state)
	default:
		return container.NewVBox(widget.NewLabel("Unknown page"))
	}
}
