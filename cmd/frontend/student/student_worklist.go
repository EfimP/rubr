package student

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
type Work struct {
	Deadline time.Time
	Title    string
	Status   string
}

// Симуляция данных из базы данных (включая просроченные работы)
var works = []Work{
	{Deadline: time.Date(2025, 6, 10, 0, 0, 0, 0, time.UTC), Title: "Лабораторная работа 1", Status: "Просрочено"},
	{Deadline: time.Date(2025, 6, 20, 0, 0, 0, 0, time.UTC), Title: "Проект по физике", Status: "В процессе"},
	{Deadline: time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC), Title: "Эссе по истории", Status: "Сдано"},
	{Deadline: time.Date(2025, 7, 5, 0, 0, 0, 0, time.UTC), Title: "Курсовая работа", Status: "Не сдано"},
}

func AllWorksScreen() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Студент: Список всех работ")
	myWindow.Resize(fyne.NewSize(1920, 1080)) // Разрешение 16:9

	// Цвета
	headerTextColor := color.White
	darkBlue := color.NRGBA{R: 20, G: 40, B: 80, A: 255}
	separatorColor := color.NRGBA{R: 200, G: 200, B: 200, A: 255} // Цвет разделителя

	// --- Верхняя панель (Header) ---
	logo := canvas.NewText("ВШЭ", headerTextColor)
	logo.TextStyle.Bold = true
	logoContainer := container.New(layout.NewMaxLayout(), logo)

	headerTitleText := canvas.NewText("Список работ", headerTextColor)
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
		log.Println("Кнопка 'назад' нажата. Возврат на предыдущий экран.")
		myWindow.Close()
	})
	backButtonRow := container.New(layout.NewMaxLayout(), backButton)

	// таблица работ с разделителями и кнопкой "Подробнее"
	var tableContent []fyne.CanvasObject
	for i, work := range works {
		// Создаем строку для каждой работы
		deadlineLabel := widget.NewLabel(work.Deadline.Format("02.01.2006"))
		titleLabel := widget.NewLabel(work.Title)
		statusLabel := widget.NewLabel(work.Status)
		detailsButton := widget.NewButton("Подробнее", func(w Work) func() {
			return func() {}
		}(work)) // Копируем Work для замыкания

		// горизонтальный контейнер для строки
		row := container.NewHBox(
			deadlineLabel,
			titleLabel,
			statusLabel,
			detailsButton,
		)

		// Добавляем строку в содержимое
		tableContent = append(tableContent, row)

		//горизонтальный разделитель, если это не последняя строка
		if i < len(works)-1 {
			separator := canvas.NewLine(separatorColor)
			separator.StrokeWidth = 2
			separator.Position1 = fyne.NewPos(0, 0)
			separator.Position2 = fyne.NewPos(1920, 0) // Длина линии равна ширине окна
			separatorContainer := container.New(layout.NewMaxLayout(), separator)
			tableContent = append(tableContent, separatorContainer)
		}
	}

	// Общий контейнер для таблицы
	mainContent := container.NewVBox(tableContent...)

	centralContentPanel := container.NewVBox(
		backButtonRow,
		mainContent,
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

	log.Println("Экран 'Список всех работ' запущен.")
	myWindow.ShowAndRun()
}
