package student

import (
	"fmt"
	"image/color"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// Структура для хранения информации о предмете и оценках
type Subject struct {
	Name    string
	Grades  []float64
	Average float64
	Details string
}

// Симуляция данных из базы данных
var subjects = []Subject{
	{
		Name:    "Математика",
		Grades:  []float64{4.0, 3.5, 4.5, 4.0},
		Average: 4.0,
		Details: "Оценки: 4.0 (контрольная 1), 3.5 (контрольная 2), 4.5 (экзамен), 4.0 (итог)",
	},
	{
		Name:    "Физика",
		Grades:  []float64{3.0, 4.0, 3.5},
		Average: 3.5,
		Details: "Оценки: 3.0 (лабораторная), 4.0 (контрольная), 3.5 (экзамен)",
	},
	{
		Name:    "Информатика",
		Grades:  []float64{5.0, 4.5, 5.0},
		Average: 4.833,
		Details: "Оценки: 5.0 (проект), 4.5 (тест), 5.0 (экзамен)",
	},
	{
		Name:    "Химия",
		Grades:  []float64{4.0, 4.0, 3.5, 4.5},
		Average: 4.0,
		Details: "Оценки: 4.0 (лабораторная), 4.0 (контрольная), 3.5 (тест), 4.5 (экзамен)",
	},
	{
		Name:    "История",
		Grades:  []float64{3.5, 4.0, 4.0},
		Average: 3.833,
		Details: "Оценки: 3.5 (эссе), 4.0 (тест), 4.0 (экзамен)",
	},
}

func StudentGradesScreen() {
	myApp := app.New()
	myWindow := myApp.NewWindow("студент: Оценки студента")
	myWindow.Resize(fyne.NewSize(1600, 900)) // Базовый размер, адаптируется

	// Цвета
	headerTextColor := color.White
	darkBlue := color.NRGBA{R: 20, G: 40, B: 80, A: 255}
	//separatorColor := color.NRGBA{R: 200, G: 200, B: 200, A: 255} // Цвет разделителя

	// --- Верхняя панель (Header) ---
	logo := canvas.NewText("ВШЭ", headerTextColor)
	logo.TextStyle.Bold = true
	logoContainer := container.New(layout.NewMaxLayout(), logo)

	headerTitleText := canvas.NewText("ОЦЕНКИ СТУДЕНТА", headerTextColor)
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

	// Создание таблицы с разделителями
	var tableContent []fyne.CanvasObject
	for i, subject := range subjects {
		// Создаем строку для каждого предмета
		nameLabel := widget.NewLabel(subject.Name)
		gradesButton := widget.NewButton("", nil)
		gradesText := fmt.Sprintf("%.2f, %.2f, %.2f", subject.Grades[0], subject.Grades[1], subject.Grades[2])
		gradesButton.SetText(gradesText)
		gradesButton.OnTapped = func(s Subject) func() {
			return func() {
				log.Println("Нажата оценка для предмета:", s.Name)
				dialog.ShowInformation("Детали оценки", s.Details, myWindow)
			}
		}(subject) // Копируем subject для замыкания
		averageLabel := widget.NewLabel(fmt.Sprintf("%.2f", subject.Average))

		// Создаем горизонтальный контейнер для строки
		row := container.NewHBox(
			nameLabel,
			gradesButton,
			averageLabel,
		)

		// Добавляем строку в содержимое
		tableContent = append(tableContent, row)

		// Добавляем разделитель, если это не последняя строка
		if i < len(subjects)-1 {
			separator := canvas.NewLine(color.Black)
			separator.StrokeWidth = 2
			separator.Position1 = fyne.NewPos(0, 0)
			separator.Position2 = fyne.NewPos(1600, 0) // Длина линии равна ширине окна
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

	log.Println("Экран оценок студента запущен.")
	myWindow.ShowAndRun()
}
