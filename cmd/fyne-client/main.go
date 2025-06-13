package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"google.golang.org/grpc"
	"log"
	rubricpb "rubr/proto/rubric"
	userpb "rubr/proto/user"
	workpb "rubr/proto/work"
)

type AppState struct {
	currentPage  string
	userID       string
	role         string
	window       fyne.Window
	userClient   userpb.UserServiceClient
	workClient   workpb.WorkServiceClient
	rubricClient rubricpb.RubricServiceClient
}

func main() {
	a := app.New()
	a.Settings().SetTheme(theme.LightTheme())
	w := a.NewWindow("Rubric Grader")

	// Подключение к userservice
	userConn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to userservice: %v", err)
	}
	defer userConn.Close()
	userClient := userpb.NewUserServiceClient(userConn)

	// Подключение к workservice
	workConn, err := grpc.Dial("localhost:50053", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to workservice: %v", err)
	}
	defer workConn.Close()
	workClient := workpb.NewWorkServiceClient(workConn)

	// Подключение к rubricservice
	rubricConn, err := grpc.Dial("localhost:50055", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to rubricservice: %v", err)
	}
	defer rubricConn.Close()
	rubricClient := rubricpb.NewRubricServiceClient(rubricConn)

	state := &AppState{
		currentPage:  "greeting",
		userID:       "",
		role:         "",
		window:       w,
		userClient:   userClient,
		workClient:   workClient,
		rubricClient: rubricClient,
	}

	w.SetContent(createContent(state))
	w.Resize(fyne.NewSize(1280, 720))
	w.ShowAndRun()
}

// createContent создает содержимое окна в зависимости от текущей страницы
func createContent(state *AppState) fyne.CanvasObject {
	leftBackground := canvas.NewImageFromFile("bin/logo/hse_logo.svg")
	leftBackground.FillMode = canvas.ImageFillStretch

	switch state.currentPage {
	case "greeting":
		return CreateGreetingPage(state, leftBackground)
	case "authorization":
		return CreateAuthorizationPage(state, leftBackground)
	case "registration":
		return CreateRegistrationPage(state, leftBackground)
	case "superacc_usrs":
		return СreateGroupListPage(state, leftBackground)
	case "lector_works":
		return CreateLectorWorksPage(state, leftBackground)
	default:
		return container.NewVBox(widget.NewLabel("Unknown page"))
	}
}
