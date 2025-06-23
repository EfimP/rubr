package main

import (
	"context"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"google.golang.org/grpc"
	"image/color"
	"log"
	"strconv"
	"time"

	notifypb "rubr/proto/notification"
	userpb "rubr/proto/user"
)

func CreateGreetingPage(state *AppState) fyne.CanvasObject {
	logo := canvas.NewImageFromResource(resourceHselogoSvg)
	logo.FillMode = canvas.ImageFillOriginal
	logo.SetMinSize(fyne.NewSize(100, 100))

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

	// Централизуем всё содержимое: логотип + кнопки
	centeredContent := container.NewVBox(
		logo,
		widget.NewLabel(""), // Можно использовать spacer при желании
		loginButton,
		registerButton,
	)

	leftContainer := container.NewCenter(centeredContent)

	// Правая часть — белый прямоугольник с текстом
	rightBackground := canvas.NewRectangle(color.RGBA{23, 44, 101, 255})
	rightText := canvas.NewText("Вход в систему оценивания", color.White)
	rightText.TextSize = 32
	rightText.TextStyle = fyne.TextStyle{Bold: true}
	rightContent := container.NewCenter(rightText)
	rightContainer := container.NewStack(rightBackground, rightContent)

	return container.New(layout.NewGridLayout(2), leftContainer, rightContainer)
}

func CreateAuthorizationPage(state *AppState) fyne.CanvasObject {
	logo := canvas.NewImageFromResource(resourceHselogoSvg)
	logo.FillMode = canvas.ImageFillOriginal
	logo.SetMinSize(fyne.NewSize(100, 100))

	loginEntry := widget.NewEntry()
	loginEntry.SetPlaceHolder("Введите логин")

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Введите пароль")

	enterButton := widget.NewButton("Войти в аккаунт", func() {
		conn, err := grpc.Dial("89.169.39.161:50051", grpc.WithInsecure())
		if err != nil {
			log.Printf("Failed to connect to userservice: %v", err)
			return
		}
		defer conn.Close()

		client := userpb.NewUserServiceClient(conn)
		resp, err := client.Login(context.Background(), &userpb.LoginRequest{
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

		state.userID = resp.UserId
		state.role = resp.Role
		switch state.role {
		case "lecturer":
			state.currentPage = "lector_works"
		case "superaccount":
			state.currentPage = "superacc-groups"
		case "assistant":
			state.currentPage = "assistant_works"
		case "student":
			state.currentPage = "student_grades"
		case "seminarist":
			state.currentPage = "seminarist_works"
		default:
			state.currentPage = "greeting"
		}
		state.window.SetContent(createContent(state))
	})
	enterButton.Importance = widget.HighImportance

	backButton := widget.NewButton("← Назад", func() {
		state.currentPage = "greeting"
		state.window.SetContent(createContent(state))
	})
	backFull := container.NewHBox(backButton)

	form := container.NewVBox(
		logo,
		layout.NewSpacer(),
		loginEntry,
		passwordEntry,
		enterButton,
		layout.NewSpacer(),
	)

	leftContent := container.NewBorder(
		nil, backFull, nil, nil,
		container.NewCenter(form),
	)

	rightBackground := canvas.NewRectangle(color.RGBA{23, 44, 101, 255})

	rightText := canvas.NewText("Войдите в аккаунт", color.White)
	rightText.TextSize = 32
	rightText.TextStyle = fyne.TextStyle{Bold: true}

	rightContent := container.NewCenter(rightText)
	rightContainer := container.NewStack(rightBackground, rightContent)

	return container.New(layout.NewGridLayout(2), leftContent, rightContainer)
}
func CreateRegistrationPage(state *AppState) fyne.CanvasObject {
	logo := canvas.NewImageFromResource(resourceHselogoSvg)
	logo.FillMode = canvas.ImageFillOriginal
	logo.SetMinSize(fyne.NewSize(100, 100))

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
		conn, err := grpc.Dial("89.169.39.161:50051", grpc.WithInsecure())
		if err != nil {
			log.Printf("Failed to connect to userservice: %v", err)
			return
		}
		defer conn.Close()

		client := userpb.NewUserServiceClient(conn)
		resp, err := client.RegisterUser(context.Background(), &userpb.RegisterUserRequest{
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

		connNotificate, err := grpc.Dial("89.169.39.161:50056", grpc.WithInsecure())
		if err != nil {
			log.Printf("Failed to connect to userservice: %v", err)
			return
		}
		defer conn.Close()

		userIDint64, err := strconv.ParseInt(resp.UserId, 10, 32)
		if err != nil {
			log.Printf("Некорректный ID пользователя: %v", err)
			return
		}
		userID := int32(userIDint64)

		clientNotificate := notifypb.NewNotificationServiceClient(connNotificate)
		respNotificate, err := clientNotificate.SendRegistrationNotification(context.Background(), &notifypb.NotificationRequest{
			UserId:    userID,
			Email:     emailEntry.Text,
			Message:   "Created new account",
			CreatedAt: time.Now().GoString(),
		})
		if err != nil {
			log.Printf("Notification failed: %v", err)
			return
		}
		if respNotificate.Error != "" {
			log.Println("Notification error:", respNotificate.Error)
			return
		}

		log.Printf("Registration successful, UserID: %s", resp.UserId)
		state.currentPage = "greeting"
		state.window.SetContent(createContent(state))
	})
	enterButton.Importance = widget.HighImportance

	backButton := widget.NewButton("← Назад", func() {
		state.currentPage = "greeting"
		state.window.SetContent(createContent(state))
	})
	backFull := container.NewHBox(backButton)

	form := container.NewVBox(
		logo,
		layout.NewSpacer(),
		nameEntry,
		surnameEntry,
		patronymicEntry,
		emailEntry,
		passwordEntry,
		enterButton,
		layout.NewSpacer(),
	)

	leftContent := container.NewBorder(
		nil, backFull, nil, nil,
		container.NewCenter(form),
	)

	rightBackground := canvas.NewRectangle(color.RGBA{23, 44, 101, 255})

	rightText := canvas.NewText("Зарегистрируйтесь", color.White)
	rightText.TextSize = 32
	rightText.TextStyle = fyne.TextStyle{Bold: true}

	rightContent := container.NewCenter(rightText)
	rightContainer := container.NewStack(rightBackground, rightContent)

	return container.New(layout.NewGridLayout(2), leftContent, rightContainer)
}
