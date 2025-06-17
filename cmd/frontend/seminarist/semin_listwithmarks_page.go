package seminarist

import (
	"image/color"
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// TaskDetails представляет структуру для деталей задания
type TaskDetails struct {
	TaskTitle  string
	Instructor string
	TaskDesc   string
	DueDate    time.Time
	SubmitDate time.Time
}

// StudentTask представляет структуру для работы каждого студента
type StudentTask struct {
	StudentName string
	Email       string
	TaskStatus  string // "проверено" или "не проверено"
	Assistant   string
	Score       string
}

// Simulated database data
var taskDetails = TaskDetails{
	TaskTitle:  "Лабораторная работа 1",
	Instructor: "Иванов И.И.",
	TaskDesc:   "Описание задания",
	DueDate:    time.Date(2025, 6, 20, 0, 0, 0, 0, time.UTC),
	SubmitDate: time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC),
}

var studentTasks = []StudentTask{
	{StudentName: "Фамилия имя отчество Студента", Email: "pochta@example.com", TaskStatus: "cтатус", Assistant: "Фамилия имя отчество Ассистента", Score: ""},
	{StudentName: "Фамилия имя отчество Студента", Email: "pochta@example.com", TaskStatus: "cтатус", Assistant: "Фамилия имя отчество Ассистента", Score: ""},
	{StudentName: "Фамилия имя отчество Студента", Email: "pochta@example.com", TaskStatus: "cтатус", Assistant: "Фамилия имя отчество Ассистента", Score: ""},
	{StudentName: "Фамилия имя отчество Студента", Email: "pochta@example.com", TaskStatus: "cтатус", Assistant: "Фамилия имя отчество Ассистента", Score: ""},
	{StudentName: "Фамилия имя отчество Студента", Email: "pochta@example.com", TaskStatus: "cтатус", Assistant: "Фамилия имя отчество Ассистента", Score: ""},
	{StudentName: "Фамилия имя отчество Студента", Email: "pochta@example.com", TaskStatus: "cтатус", Assistant: "Фамилия имя отчество Ассистента", Score: ""},
	{StudentName: "Фамилия имя отчество Студента", Email: "pochta@example.com", TaskStatus: "cтатус", Assistant: "Фамилия имя отчество Ассистента", Score: ""},
	{StudentName: "Фамилия имя отчество Студента", Email: "pochta@example.com", TaskStatus: "cтатус", Assistant: "Фамилия имя отчество Ассистента", Score: ""},
	{StudentName: "Фамилия имя отчество Студента", Email: "pochta@example.com", TaskStatus: "cтатус", Assistant: "Фамилия имя отчество Ассистента", Score: ""},
	{StudentName: "Фамилия имя отчество Студента", Email: "pochta@example.com", TaskStatus: "cтатус", Assistant: "Фамилия имя отчество Ассистента", Score: ""},
	{StudentName: "Фамилия имя отчество Студента", Email: "pochta@example.com", TaskStatus: "cтатус", Assistant: "Фамилия имя отчество Ассистента", Score: ""},
	{StudentName: "Фамилия имя отчество Студента", Email: "pochta@example.com", TaskStatus: "cтатус", Assistant: "Фамилия имя отчество Ассистента", Score: ""},
}

func AssignmentScreen() {
	myApp := app.New()
	mainWindow := myApp.NewWindow("ассистент: Задание")
	mainWindow.Resize(fyne.NewSize(1920, 1080))

	// Colors
	textColor := color.White
	blueShade := color.NRGBA{R: 20, G: 40, B: 80, A: 255}

	// Header
	uniLogo := canvas.NewText("ВШЭ", textColor)
	uniLogo.TextStyle.Bold = true
	uniLogo.TextSize = 28
	logoSection := container.New(layout.NewHBoxLayout(), uniLogo)

	taskTitleText := canvas.NewText("Задание", textColor)
	taskTitleText.TextStyle.Bold = true
	taskTitleText.Alignment = fyne.TextAlignCenter

	headerPanel := container.New(layout.NewBorderLayout(nil, nil, logoSection, nil),
		logoSection,
		container.NewCenter(taskTitleText),
	)
	headerBackground := canvas.NewRectangle(blueShade)
	headerWithBackground := container.NewStack(headerBackground, headerPanel)

	// Back button
	returnButton := widget.NewButton("назад", func() {
		log.Println(time.Now().Format("15:04:05"), "Кнопка 'назад' нажата. Возврат на предыдущий экран.")
		mainWindow.Close()
	})
	buttonRow := container.New(layout.NewHBoxLayout(), returnButton)

	// Assignment info with scrolling
	taskDescLabel := widget.NewLabel("Описание: " + taskDetails.TaskDesc)
	taskDescLabel.Wrapping = fyne.TextWrapWord

	taskInfo := container.NewVBox(
		widget.NewLabel("Тема работы: "+taskDetails.TaskTitle),
		widget.NewLabel("ФИО лектора: "+taskDetails.Instructor),
		taskDescLabel,
		widget.NewLabel("Дедлайн: "+taskDetails.DueDate.Format("02.01.2006")),
		widget.NewLabel("Дата сдачи: "+taskDetails.SubmitDate.Format("02.01.2006")),
	)
	taskInfoScroll := container.NewScroll(taskInfo)
	taskInfoScroll.SetMinSize(fyne.NewSize(1300, 200))

	// Student works table
	taskTable := widget.NewTable(
		func() (int, int) { return len(studentTasks), 5 },
		func() fyne.CanvasObject {
			return container.NewHBox(widget.NewLabel(""))
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			row, col := id.Row, id.Col
			cellContainer := cell.(*fyne.Container)

			switch col {
			case 0: // ФИО почта
				label := widget.NewLabel(studentTasks[row].StudentName + " (" + studentTasks[row].Email + ")")
				label.Alignment = fyne.TextAlignLeading
				cellContainer.Objects = []fyne.CanvasObject{label}
			case 1: // Статус
				label := widget.NewLabel(studentTasks[row].TaskStatus)
				label.Alignment = fyne.TextAlignLeading
				cellContainer.Objects = []fyne.CanvasObject{label}
			case 2: // ФИО ассистента
				label := widget.NewLabel(studentTasks[row].Assistant)
				label.Alignment = fyne.TextAlignLeading
				cellContainer.Objects = []fyne.CanvasObject{label}
			case 3: // Оценка
				label := widget.NewLabel(studentTasks[row].Score)
				label.Alignment = fyne.TextAlignLeading
				cellContainer.Objects = []fyne.CanvasObject{label}
			case 4: // Кнопки
				viewBtn := widget.NewButton("посмотреть", func() {
					log.Println(time.Now().Format("15:04:05"), "Нажата кнопка 'посмотреть' для", studentTasks[row].StudentName)
				})
				uploadBtn := widget.NewButton("загрузить", func() {
					log.Println(time.Now().Format("15:04:05"), "Нажата кнопка 'загрузить работу' для", studentTasks[row].StudentName)
				})
				cellContainer.Objects = []fyne.CanvasObject{viewBtn, uploadBtn}
			}
			cell.Refresh()
		},
	)

	// Table column widths
	taskTable.SetColumnWidth(0, 400) // ФИО почта
	taskTable.SetColumnWidth(1, 100) // Статус
	taskTable.SetColumnWidth(2, 300) // ФИО ассистента
	taskTable.SetColumnWidth(3, 50)  // Оценка
	taskTable.SetColumnWidth(4, 200) // Кнопки

	// Scrollable table
	tableScrollArea := container.NewScroll(taskTable)
	tableScrollArea.SetMinSize(fyne.NewSize(1300, 400))

	// Declare and initialize mainLayout before toggleButton
	var mainLayout *fyne.Container

	// Initialize mainLayout after toggleButton definition
	mainLayout = container.NewVBox(
		buttonRow,
		taskInfoScroll,
		tableScrollArea,
	)

	contentBackground := canvas.NewRectangle(color.White)
	contentWithBackground := container.NewStack(contentBackground, mainLayout)

	// Window layout
	mainWindow.SetContent(container.NewBorder(
		headerWithBackground,
		nil,
		nil,
		nil,
		contentWithBackground,
	))

	log.Println(time.Now().Format("15:04:05"), "Экран 'Задание' запущен.")
	mainWindow.ShowAndRun()
}
