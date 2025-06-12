package main

import (
	"context"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"google.golang.org/grpc"
	"image/color"
	"log"
	userpb "rubr/proto/user" // Алиас для user.proto
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
	case "superacc_usrs":
		return createGroupListPage(state, leftBackground)
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
		if loginEntry.Text == "superacc" && passwordEntry.Text == "fimoz" {
			state.currentPage = "superacc_usrs"
			state.window.SetContent(createContent(state))
			return
		}
		log.Println("Login successful, Token:", resp.Token, " user id: ", resp.UserId)
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
		log.Printf("Registration successful, UserID: %s", resp.UserId)
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

type CriterionEntry struct {
	NameEntry        *widget.Entry
	DescriptionEntry *widget.Entry
	CommentEntry     *widget.Entry
	EvaluationEntry  *widget.Entry
	Container        *fyne.Container
}

func createGroupListPage(state *AppState, leftBackground *canvas.Image) fyne.CanvasObject {
	w := state.window
	headerTextColor := color.White

	logoText := canvas.NewText("ВШЭ", headerTextColor)
	logoText.TextStyle.Bold = true
	logoText.TextSize = 24
	logoText.Alignment = fyne.TextAlignCenter
	leftHeaderObject := container.NewStack(logoText)

	headerTitle := canvas.NewText("Список групп", headerTextColor)
	headerTitle.TextStyle.Bold = true
	headerTitle.TextSize = 20
	headerTitle.Alignment = fyne.TextAlignCenter

	// логотип слева, текст заголовка по центру
	headerContent := container.New(layout.NewBorderLayout(nil, nil, leftHeaderObject, nil),
		leftHeaderObject,
		container.NewCenter(headerTitle),
	)

	// "Назад" для возврата на предыдущую страницу
	backButton := widget.NewButton("Назад", func() {
		fmt.Println("Кнопка 'Назад' нажата. Возврат на предыдущую страницу.")
		state.currentPage = "greeting"
		w.SetContent(createContent(state))
	})
	backButtonContainer := container.NewHBox(layout.NewSpacer(), backButton)

	criteriaListContainer := container.NewVBox()

	columnHeaders := container.New(layout.NewGridLayoutWithColumns(4),
		container.NewPadded(widget.NewLabelWithStyle("Название группы", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
		container.NewPadded(widget.NewLabelWithStyle("Описание", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
		container.NewPadded(widget.NewLabelWithStyle("Дисциплинны", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
	)

	// Слайс для хранения активных критериев
	var activeCriteria []*CriterionEntry

	addCriterionEntry := func() {
		nameEntry := widget.NewEntry()
		nameEntry.SetPlaceHolder("Название группы")
		nameEntry.TextStyle = fyne.TextStyle{Monospace: false}
		nameEntryContainer := container.NewMax(nameEntry)
		nameEntryContainer.Resize(fyne.NewSize(250, 60))

		descriptionEntry := widget.NewEntry()
		descriptionEntry.SetPlaceHolder("Описание")
		descriptionEntry.TextStyle = fyne.TextStyle{Monospace: false}
		descriptionEntryContainer := container.NewMax(descriptionEntry)
		descriptionEntryContainer.Resize(fyne.NewSize(250, 60))

		commentEntry := widget.NewEntry()
		commentEntry.SetPlaceHolder("Дисциплинны")
		commentEntry.TextStyle = fyne.TextStyle{Monospace: false}
		commentEntryContainer := container.NewMax(commentEntry)
		commentEntryContainer.Resize(fyne.NewSize(250, 60))

		nextbottom := widget.NewButton("Подробнее", func() {
			fmt.Println("Кнопка 'Далее' нажата. Собираем данные критериев:")
		})

		criterionRow := container.New(layout.NewGridLayoutWithColumns(4),
			container.NewPadded(container.NewPadded(nameEntryContainer)),
			container.NewPadded(container.NewPadded(descriptionEntryContainer)),
			container.NewPadded(container.NewPadded(commentEntryContainer)),
			container.NewPadded(container.NewPadded(nextbottom)),
		)

		//добавление строки в контейнер списка критериев
		criteriaListContainer.Add(criterionRow)
		criteriaListContainer.Refresh() // Обновляем UI

		// Сохраняет критерий в слайс activeCriteria
		activeCriteria = append(activeCriteria, &CriterionEntry{
			NameEntry:        nameEntry,
			DescriptionEntry: descriptionEntry,
			CommentEntry:     commentEntry,
			Container:        criterionRow,
		})
	}

	// удаление критерия из таблицы
	deleteCriterionEntry := func() {
		if len(activeCriteria) > 0 {
			//индекс последнего критерия
			lastIndex := len(activeCriteria) - 1
			lastCriterion := activeCriteria[lastIndex]
			// удаление строки из UI
			criteriaListContainer.Remove(lastCriterion.Container)
			// Удаляем критерий из слайса
			activeCriteria = activeCriteria[:lastIndex]
			// Обновляем UI
			criteriaListContainer.Refresh()
			fmt.Println("Последний критерий удален.")
		} else {
			fmt.Println("Нет критериев для удаления.")
		}
	}

	// пустая строка
	addCriterionEntry()

	// Метка для списка критериев
	listLabel := widget.NewLabelWithStyle("Список групп", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	//скролл для списка критериев
	scrollableCriteria := container.NewVScroll(criteriaListContainer)
	scrollableCriteria.SetMinSize(fyne.NewSize(0, 400)) // Устанавливаем минимальную высоту для прокрутки

	// Кнопка "Добавить" для создания новой строки критерия
	addButton := widget.NewButton("Добавить", func() {
		addCriterionEntry()
	})

	// Кнопка "Удалить" для удаления последней строки критерия
	deleteButton := widget.NewButton("Удалить", func() {
		deleteCriterionEntry()
	})

	// Кнопка "Далее" для обработки введенных данных
	nextButton := widget.NewButton("Далее", func() {
		fmt.Println("Кнопка 'Далее' нажата. Собираем данные групп:")
		for i, criterion := range activeCriteria {
			fmt.Printf("Группа %d:\n", i+1)
			fmt.Printf("  Название: %s\n", criterion.NameEntry.Text)
			fmt.Printf("  Описание: %s\n", criterion.DescriptionEntry.Text)
			fmt.Printf("  Дисциплина: %s\n", criterion.CommentEntry.Text)
		}
	})

	// добавить удалить далее
	bottomButtons := container.New(layout.NewHBoxLayout(),
		addButton,
		deleteButton,
		layout.NewSpacer(),
		nextButton,
	)
	bottomButtonsWithPadding := container.NewPadded(bottomButtons)

	// Фон центральной области (белый прямоугольник)
	contentBackground := canvas.NewRectangle(color.White)

	// Панель критериев, содержащая заголовки, метку и прокручиваемую область
	criteriaPanel := container.NewVBox(
		container.NewPadded(columnHeaders),
		listLabel,
		scrollableCriteria,
	)

	// Центральный контент с фоном и отступами
	centralContent := container.NewStack(
		contentBackground,
		container.NewPadded(criteriaPanel),
	)
	
	return container.NewStack(
		canvas.NewRectangle(color.NRGBA{R: 20, G: 40, B: 80, A: 255}), // Фон окна
		container.NewBorder(
			container.NewVBox(headerContent, backButtonContainer),
			bottomButtonsWithPadding,
			nil,
			nil,
			centralContent,
		),
	)
}
