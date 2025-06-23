package main

import (
	"context"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"google.golang.org/grpc"
	"image/color"
	"log"
	"math/rand"
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

	resetPasswordButton := widget.NewButton("Сброс пароля", func() {
		state.currentPage = "password_reset"
		state.window.SetContent(createContent(state))
	})
	resetPasswordButton.Importance = widget.MediumImportance

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
		resetPasswordButton,
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

func CreatePasswordResetPage(state *AppState) fyne.CanvasObject {
	logo := canvas.NewImageFromResource(resourceHselogoSvg)
	logo.FillMode = canvas.ImageFillOriginal
	logo.SetMinSize(fyne.NewSize(100, 100))

	emailEntry := widget.NewEntry()
	emailEntry.SetPlaceHolder("Введите ваш email")

	var tempPasswordEntry *widget.Entry
	var newPasswordEntry *widget.Entry
	var confirmButton *widget.Button
	var currentForm fyne.CanvasObject

	// Правая часть окна (константная часть)
	rightBackground := canvas.NewRectangle(color.RGBA{23, 44, 101, 255})
	rightText := canvas.NewText("Сброс пароля", color.White)
	rightText.TextSize = 32
	rightText.TextStyle = fyne.TextStyle{Bold: true}
	rightContent := container.NewCenter(rightText)
	rightContainer := container.NewStack(rightBackground, rightContent)

	// Изначальный контейнер для первого этапа (ввод email)
	initialForm := container.NewVBox(
		logo,
		layout.NewSpacer(),
		emailEntry,
		widget.NewButton("Далее", func() {
			conn, err := grpc.Dial("89.169.39.161:50056", grpc.WithInsecure())
			if err != nil {
				log.Printf("Failed to connect to notificationservice: %v", err)
				return
			}
			defer conn.Close()

			client := notifypb.NewNotificationServiceClient(conn)
			// Генерируем случайный 4-значный пароль
			b := make([]byte, 3)
			_, err = rand.Read(b)
			if err != nil {
				log.Printf("Failed to generate random password: %v", err)
				return
			}
			tempPassword := fmt.Sprintf("%04d", int(b[0])%10000)
			createdAt := time.Now().Format(time.RFC3339)

			_, err = client.SendPasswordResetNotification(context.Background(), &notifypb.NotificationRequest{
				Email:     emailEntry.Text,
				Message:   fmt.Sprintf("Ваш временный пароль: %s\nВведите его в приложении для создания нового пароля.", tempPassword),
				CreatedAt: createdAt,
			})
			if err != nil {
				log.Printf("Password reset notification failed: %v", err)
				return
			}

			// Инициализируем поля для второго этапа
			tempPasswordEntry = widget.NewPasswordEntry()
			tempPasswordEntry.SetPlaceHolder("Введите временный пароль")
			newPasswordEntry = widget.NewPasswordEntry()
			newPasswordEntry.SetPlaceHolder("Введите новый пароль")

			confirmButton = widget.NewButton("Подтвердить", func() {
				if tempPasswordEntry.Text == "" || newPasswordEntry.Text == "" {
					dialog.ShowInformation("Ошибка", "Заполните все поля", state.window)
					return
				}

				// Проверка временного пароля (эмуляция на клиенте)
				if tempPasswordEntry.Text != tempPassword {
					dialog.ShowInformation("Ошибка", "Неверный временный пароль", state.window)
					return
				}

				// Обновляем пароль
				userConn, err := grpc.Dial("89.169.39.161:50051", grpc.WithInsecure())
				if err != nil {
					log.Printf("Failed to connect to userservice: %v", err)
					return
				}
				defer userConn.Close()

				userClient := userpb.NewUserServiceClient(userConn)
				_, err = userClient.UpdatePassword(context.Background(), &userpb.UpdatePasswordRequest{
					Email:    emailEntry.Text,
					Password: newPasswordEntry.Text,
				})
				if err != nil {
					log.Printf("Failed to update password: %v", err)
					return
				}
				dialog.ShowInformation("Успех", "Пароль успешно обновлён", state.window)
				state.currentPage = "authorization"
				state.window.SetContent(createContent(state))
			})
			confirmButton.Importance = widget.HighImportance

			// Перестраиваем форму для второго этапа
			secondForm := container.NewVBox(
				logo,
				layout.NewSpacer(),
				tempPasswordEntry,
				newPasswordEntry,
				confirmButton,
				layout.NewSpacer(),
			)
			currentForm = secondForm
			leftContent := container.NewBorder(nil, nil, nil, nil, container.NewCenter(secondForm))
			state.window.SetContent(container.New(layout.NewGridLayout(2), leftContent, rightContainer))
		}),
		layout.NewSpacer(),
	)

	backButton := widget.NewButton("← Назад", func() {
		state.currentPage = "authorization"
		state.window.SetContent(createContent(state))
	})
	backFull := container.NewHBox(backButton)

	// Изначальный leftContent
	leftContent := container.NewBorder(
		nil, backFull, nil, nil,
		container.NewCenter(initialForm),
	)
	currentForm = initialForm

	if currentForm == initialForm {

	}

	return container.New(layout.NewGridLayout(2), leftContent, rightContainer)
}
