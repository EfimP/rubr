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

// Структура для хранения информации о работе
type Work2 struct {
	Date        time.Time
	Title       string
	Studentname string
}

// Структура для хранения информации об ассистентах
type Assistant2 struct {
	ID   int
	Name string
}

// Симуляция данных из базы данных
var works2 = []Work{
	{Date: time.Date(2025, 6, 10, 0, 0, 0, 0, time.UTC), Title: "НАЗВАНЕ", Studentname: "ФАМИЛИЯ ИМЯ ОТЧЕСТВО"},
}

var assistants2 = []Assistant2{
	{ID: 1, Name: "Ассистент 1"},
	{ID: 2, Name: "Ассистент 2"},
	{ID: 3, Name: "Ассистент 3"},
}

func AssignmentScreen() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Семиннарист: Задание")
	myWindow.Resize(fyne.NewSize(1920, 1080)) // Разрешение 16:9

	// Цвета
	headerTextColor := color.White
	darkBlue := color.NRGBA{R: 20, G: 40, B: 80, A: 255}

	// --- Верхняя панель (Header) ---
	logo := canvas.NewText("ВШЭ", headerTextColor)
	logo.TextStyle.Bold = true
	logo.TextSize = 28
	logoContainer := container.New(layout.NewMaxLayout(), logo)

	headerTitleText := canvas.NewText("Задание", headerTextColor)
	headerTitleText.TextStyle.Bold = true
	headerTitleText.Alignment = fyne.TextAlignCenter

	header := container.New(layout.NewBorderLayout(nil, nil, logoContainer, nil),
		logoContainer,
		container.NewCenter(headerTitleText),
	)
	headerBackground := canvas.NewRectangle(darkBlue)
	headerWithBackground := container.NewStack(headerBackground, header)

	// --- Кнопка "назад" ---
	backButton := widget.NewButton("назад", func() {
		log.Println(time.Now().Format("15:04:05"), "Кнопка 'назад' нажата. Возврат на предыдущий экран.")
		myWindow.Close()
	})
	backButtonRow := container.New(layout.NewMaxLayout(), backButton)

	// --- Левый сайдбар ---
	sidebarBackground := canvas.NewRectangle(darkBlue)
	sidebar := container.New(layout.NewMaxLayout(), sidebarBackground, container.NewVBox(layout.NewSpacer(), layout.NewSpacer()))

	// --- Центральная секция с данными и кнопкой ---
	assignmentContent := container.NewVBox()

	// Поля для отображения данных
	work := works[0] // Берем первую работу из базы данных
	title := widget.NewLabel(work.Title)
	title.TextStyle.Bold = true
	title.Alignment = fyne.TextAlignLeading

	assistent := widget.NewLabel(work.Studentname)
	assistent.TextStyle.Bold = true
	assistent.Alignment = fyne.TextAlignLeading

	// Описание с прокруткой
	description := widget.NewLabel("")
	description.TextStyle.Bold = true
	description.Alignment = fyne.TextAlignLeading
	description.Wrapping = fyne.TextWrapWord // Включаем перенос слов для лучшей читаемости
	descriptionScroll := container.NewScroll(description)
	descriptionScroll.SetMinSize(fyne.NewSize(1200, 300)) // Фиксированный размер для прокрутки
	descriptionScroll.ScrollToTop()                       // Включаем вертикальную прокрутку

	// Поле для дедлайна
	deadline := widget.NewLabel(work.Date.Format("02.01.2006"))
	deadline.TextStyle.Bold = true
	deadline.Alignment = fyne.TextAlignLeading

	// Кнопка "Назначить ассистента" с выпадающим списком
	var selectAssistant *widget.Select
	assignButton := widget.NewButton("Назначить ассистента", func() {
		if selectAssistant == nil {
			// Создаем выпадающий список при первом нажатии
			assistantNames := make([]string, len(assistants))
			for i, a := range assistants {
				assistantNames[i] = a.Name
			}
			selectAssistant = widget.NewSelect(assistantNames, func(selected string) {
				log.Println(time.Now().Format("15:04:05"), "Выбран ассистент:", selected)
			})
			selectAssistant.SetSelectedIndex(-1) // Нет начального выбора
			assignmentContent.Add(selectAssistant)
		} else {
			// Показываем/скрываем список
			if selectAssistant.Visible() {
				selectAssistant.Hide()
			} else {
				selectAssistant.Show()
			}
		}
		assignmentContent.Refresh()
	})

	// Добавляем поля и кнопку в содержимое
	assignmentContent.Add(container.NewHBox(widget.NewLabel("Название работы:"), title))
	assignmentContent.Add(container.NewHBox(widget.NewLabel("ФИО:"), assistent))
	assignmentContent.Add(container.NewHBox(widget.NewLabel("Описание работы:"), descriptionScroll))
	assignmentContent.Add(container.NewHBox(widget.NewLabel("Дедлайн:"), deadline))
	assignmentContent.Add(assignButton)

	// Центральный контейнер
	centralContent := container.NewHBox(
		sidebar,
		assignmentContent,
	)

	centralContentPanel := container.NewVBox(
		backButtonRow,
		centralContent,
	)

	contentBackground := canvas.NewRectangle(color.White)
	centralContentWithBackground := container.NewStack(contentBackground, centralContentPanel)

	// --- Общая компоновка окна ---
	myWindow.SetContent(container.NewBorder(
		headerWithBackground,
		nil,
		nil,
		nil,
		centralContentWithBackground,
	))

	log.Println(time.Now().Format("15:04:05"), "Экран 'Задание' запущен.")
	myWindow.ShowAndRun()
}
