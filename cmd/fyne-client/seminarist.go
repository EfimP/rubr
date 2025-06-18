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
	workpb "rubr/proto/work"
	"time"
)

type Work struct {
	ID          int32
	Date        time.Time
	Title       string
	StudentName string
}

type Task struct {
	ID    int32
	Date  time.Time
	Title string
}

func CreateSeminaristWorksPage(state *AppState) fyne.CanvasObject {
	// Цвета
	headerTextColor := color.White
	darkBlue := color.NRGBA{R: 20, G: 40, B: 80, A: 255}
	separatorColor := color.NRGBA{R: 200, G: 200, B: 200, A: 255}

	// Подключение к WorkService
	client, err := grpc.Dial("localhost:50053", grpc.WithInsecure())
	if err != nil {
		log.Printf("Failed to connect to workservice: %v", err)
		return container.NewVBox(widget.NewLabel("Ошибка подключения к серверу"))
	}
	workClient := workpb.NewWorkServiceClient(client)

	// Загрузка работ студентов (первая вкладка)
	studentWorksResp, err := workClient.GetStudentWorksForSeminarist(context.Background(), &workpb.GetStudentWorksForSeminaristRequest{
		SeminaristId: state.userID,
	})
	if err != nil {
		log.Printf("Failed to load student works: %v", err)
		return container.NewVBox(widget.NewLabel("Ошибка загрузки работ студентов"))
	}
	if studentWorksResp.Error != "" {
		log.Println("Error from server:", studentWorksResp.Error)
		return container.NewVBox(widget.NewLabel("Ошибка: " + studentWorksResp.Error))
	}

	studentWorks := make([]Work, len(studentWorksResp.Works))
	for i, w := range studentWorksResp.Works {
		createdAt, _ := time.Parse(time.RFC3339, w.CreatedAt)
		studentWorks[i] = Work{
			ID:          w.Id,
			Date:        createdAt,
			Title:       w.Title,
			StudentName: w.StudentName,
		}
	}

	// Загрузка задач лектора (вторая вкладка)
	tasksResp, err := workClient.GetTasksForSeminarist(context.Background(), &workpb.GetTasksForSeminaristRequest{
		SeminaristId: state.userID,
	})
	if err != nil {
		log.Printf("Failed to load tasks: %v", err)
		return container.NewVBox(widget.NewLabel("Ошибка загрузки задач"))
	}
	if tasksResp.Error != "" {
		log.Println("Error from server:", tasksResp.Error)
		return container.NewVBox(widget.NewLabel("Ошибка: " + tasksResp.Error))
	}

	tasks := make([]Task, len(tasksResp.Tasks))
	for i, t := range tasksResp.Tasks {
		deadline, _ := time.Parse(time.RFC3339, t.Deadline)
		tasks[i] = Task{
			ID:    t.Id,
			Date:  deadline,
			Title: t.Title,
		}
	}

	// Верхняя панель (Header)
	logo := canvas.NewText("ВШЭ", headerTextColor)
	logo.TextStyle.Bold = true
	logoContainer := container.New(layout.NewMaxLayout(), logo)

	headerTitleText := canvas.NewText("Панель семинариста", headerTextColor)
	headerTitleText.TextStyle.Bold = true
	headerTitleText.Alignment = fyne.TextAlignCenter

	header := container.New(layout.NewBorderLayout(nil, nil, logoContainer, nil),
		logoContainer,
		container.NewCenter(headerTitleText),
	)
	headerBackground := canvas.NewRectangle(darkBlue)
	headerWithBackground := container.NewStack(headerBackground, header)

	// Кнопка "назад"
	backButton := widget.NewButton("Выйти из аккаунта", func() {
		log.Println("Кнопка 'Выйти из аккаунта' нажата. Возврат на экран авторизации.")
		state.currentPage = "greeting"
		state.window.SetContent(createContent(state))
	})
	backButtonRow := container.New(layout.NewMaxLayout(), backButton)

	// Первая вкладка: Работы студентов
	studentWorksContent := createWorksTable(studentWorks, separatorColor)
	studentWorksTab := container.NewTabItem("Работы студентов", studentWorksContent)

	// Вторая вкладка: Задачи лектора
	tasksContent := createTasksTable(tasks, separatorColor)
	tasksTab := container.NewTabItem("Работы от лектора", tasksContent)

	// Контейнер вкладок
	tabs := container.NewAppTabs(studentWorksTab, tasksTab) // ?

	centralContentPanel := container.NewVBox(
		backButtonRow,
		tabs,
	)

	contentBackground := canvas.NewRectangle(color.White)
	centralContentWithBackground := container.NewStack(contentBackground, centralContentPanel)

	return container.NewBorder(
		headerWithBackground,
		nil,
		nil,
		nil,
		centralContentWithBackground,
	)
}

func createWorksTable(works []Work, separatorColor color.Color) fyne.CanvasObject {
	var tableContent []fyne.CanvasObject
	for i, work := range works {
		titleLabel := widget.NewLabel(work.Title)
		studentName := widget.NewLabel(work.StudentName)
		date := widget.NewLabel(work.Date.Format("02.01.2006"))

		detailsButton := widget.NewButton("Перейти", func() {
			log.Printf("Перейти к работе ID: %d", work.ID)
		})

		row := container.NewHBox(
			titleLabel,
			studentName,
			date,
			detailsButton,
		)

		tableContent = append(tableContent, row)

		if i < len(works)-1 {
			separator := canvas.NewLine(separatorColor)
			separator.StrokeWidth = 2
			separator.Position1 = fyne.NewPos(0, 0)
			separator.Position2 = fyne.NewPos(1920, 0)
			separatorContainer := container.New(layout.NewMaxLayout(), separator)
			tableContent = append(tableContent, separatorContainer)
		}
	}

	return container.NewVBox(tableContent...)
}

func createTasksTable(tasks []Task, separatorColor color.Color) fyne.CanvasObject {
	var tableContent []fyne.CanvasObject
	for i, task := range tasks {
		titleLabel := widget.NewLabel(task.Title)
		date := widget.NewLabel(task.Date.Format("02.01.2006"))

		detailsButton := widget.NewButton("Перейти", func() {
			log.Printf("Перейти к задаче ID: %d", task.ID)
		})

		row := container.NewHBox(
			titleLabel,
			date,
			detailsButton,
		)

		tableContent = append(tableContent, row)

		if i < len(tasks)-1 {
			separator := canvas.NewLine(separatorColor)
			separator.StrokeWidth = 2
			separator.Position1 = fyne.NewPos(0, 0)
			separator.Position2 = fyne.NewPos(1920, 0)
			separatorContainer := container.New(layout.NewMaxLayout(), separator)
			tableContent = append(tableContent, separatorContainer)
		}
	}

	return container.NewVBox(tableContent...)
}
