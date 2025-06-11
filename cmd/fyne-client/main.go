package main

import (
	"context"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"google.golang.org/grpc"
	"image/color"
	"log"
	pb "rubr/proto/user"
)

type AppState struct {
	currentPage string
	window      fyne.Window
}

func main() {
	// Инициализация приложения и окна
	a := app.New()
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
	case "greeting":
		return createGreetingPage(state, leftBackground)
	case "authorization":
		return createAuthorizationPage(state, leftBackground)
	case "registration":
		return createRegistrationPage(state, leftBackground)
	default:
		return container.NewVBox(widget.NewLabel("Unknown page"))
	}
}

// createGreetingPage создает страницу приветствия
func createGreetingPage(state *AppState, leftBackground *canvas.Image) fyne.CanvasObject {
	loginButton := widget.NewButton("Авторизоваться", func() {
		state.currentPage = "authorization"
		state.window.SetContent(createContent(state))
	})
	loginButton.Importance = widget.HighImportance

	registerButton := widget.NewButton("Зарегистрироваться", func() {
		state.currentPage = "registration"
		state.window.SetContent(createContent(state))
	})
	registerButton.Importance = widget.MediumImportance

	leftContent := container.NewVBox(
		layout.NewSpacer(),
		loginButton,
		registerButton,
		layout.NewSpacer(),
	)
	leftContainer := container.NewStack(leftBackground, container.NewCenter(leftContent))

	rightBackground := canvas.NewRectangle(color.RGBA{23, 44, 101, 255})
	rightText := widget.NewLabel("Вход в систему оценивания")
	rightText.TextStyle = fyne.TextStyle{Bold: true}
	rightContent := container.NewCenter(rightText)
	rightContainer := container.NewStack(rightBackground, rightContent)

	return container.New(layout.NewGridLayout(2), leftContainer, rightContainer)
}

// createAuthorizationPage создает страницу авторизации
func createAuthorizationPage(state *AppState, leftBackground *canvas.Image) fyne.CanvasObject {
	loginEntry := widget.NewEntry()
	loginEntry.SetPlaceHolder("Введите логин")

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Введите пароль")

	enterButton := widget.NewButton("Войти в аккаунт", func() {
		// Вызов авторизации через gRPC
		conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
		if err != nil {
			log.Printf("Failed to connect to gRPC: %v", err)
			return
		}
		defer conn.Close()

		client := pb.NewUserServiceClient(conn)
		resp, err := client.Login(context.Background(), &pb.LoginRequest{
			Email:    loginEntry.Text,
			Password: passwordEntry.Text,
		})
		if err != nil {
			log.Printf("Login failed: %v", err)
			return
		}
		if resp.Error != "" {
			log.Println("Login error:", resp.Error)
			return
		}
		log.Println("Login successful, Token:", resp.Token)
		// Здесь можно сохранить токен и перейти на страницу профиля
	})
	enterButton.Importance = widget.HighImportance

	leftContent := container.NewVBox(
		layout.NewSpacer(),
		loginEntry,
		passwordEntry,
		enterButton,
		layout.NewSpacer(),
	)
	leftContainer := container.NewStack(leftBackground, container.NewCenter(leftContent))

	rightBackground := canvas.NewRectangle(color.RGBA{23, 44, 101, 255})
	rightText := widget.NewLabel("Войдите в аккаунт")
	rightText.TextStyle = fyne.TextStyle{Bold: true}
	rightContent := container.NewCenter(rightText)
	rightContainer := container.NewStack(rightBackground, rightContent)

	return container.New(layout.NewGridLayout(2), leftContainer, rightContainer)
}

// createRegistrationPage создает страницу регистрации
func createRegistrationPage(state *AppState, leftBackground *canvas.Image) fyne.CanvasObject {
	emailEntry := widget.NewEntry()
	emailEntry.SetPlaceHolder("Введите почту")

	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Имя")

	surnameEntry := widget.NewEntry()
	surnameEntry.SetPlaceHolder("Фамилия")

	patronymicEntry := widget.NewEntry()
	patronymicEntry.SetPlaceHolder("Отчество")

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Введите пароль")

	enterButton := widget.NewButton("Зарегистрироваться", func() {
		// Вызов регистрации через gRPC
		conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
		if err != nil {
			log.Printf("Failed to connect to gRPC: %v", err)
			return
		}
		defer conn.Close()

		client := pb.NewUserServiceClient(conn)
		resp, err := client.RegisterUser(context.Background(), &pb.RegisterUserRequest{
			Email:      emailEntry.Text,
			Password:   passwordEntry.Text,
			Name:       nameEntry.Text,
			Surname:    surnameEntry.Text,
			Patronymic: patronymicEntry.Text,
		})
		if err != nil {
			log.Printf("Registration failed: %v", err)
			return
		}
		if resp.Error != "" {
			log.Println("Registration error:", resp.Error)
			return
		}
		log.Println("Registration successful, UserID:", resp.UserId)
		// Здесь можно вернуться на GreetingPage или перейти на профиль
		state.currentPage = "greeting"
		state.window.SetContent(createContent(state))
	})
	enterButton.Importance = widget.HighImportance

	leftContent := container.NewVBox(
		layout.NewSpacer(),
		nameEntry,
		surnameEntry,
		patronymicEntry,
		emailEntry,
		passwordEntry,
		enterButton,
		layout.NewSpacer(),
	)
	leftContainer := container.NewStack(leftBackground, container.NewCenter(leftContent))

	rightBackground := canvas.NewRectangle(color.RGBA{23, 44, 101, 255})
	rightText := widget.NewLabel("Зарегистрируйтесь")
	rightText.TextStyle = fyne.TextStyle{Bold: true}
	rightContent := container.NewCenter(rightText)
	rightContainer := container.NewStack(rightBackground, rightContent)

	return container.New(layout.NewGridLayout(2), leftContainer, rightContainer)
}
