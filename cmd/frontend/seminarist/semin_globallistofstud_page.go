package seminarist

import (
	"image/color"
	"log"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// Структура для хранения информации о работе
type Work1 struct {
	Date        time.Time
	Title       string
	Studentname string
	Email       string // Добавлено поле для почты
}

// Структура для хранения информации об ассистентах
type Assistant1 struct {
	ID   int
	Name string
}

// Симуляция данных из базы данных
var works1 = []Work1{
	{Date: time.Date(2025, 6, 10, 0, 0, 0, 0, time.UTC), Title: "Лабораторная работа 1", Studentname: "Иванов И.И.", Email: "ivanov@example.com"},
	{Date: time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC), Title: "Проект 1", Studentname: "Петров П.П.", Email: "petrov@example.com"},
	{Date: time.Date(2025, 6, 20, 0, 0, 0, 0, time.UTC), Title: "Лабораторная работа 2", Studentname: "Сидоров С.С.", Email: "sidorov@example.com"},
}

var assistants1 = []Assistant1{
	{ID: 1, Name: "Ассистент 1"},
	{ID: 2, Name: "Ассистент 2"},
	{ID: 3, Name: "Ассистент 3"},
}

func ListofStudensandAssistanse() {
	myApp := app.New()
	myWindow := myApp.NewWindow("seminarist")
	myWindow.Resize(fyne.NewSize(1920, 1080)) // Разрешение 16:9

	// Цвета
	headerTextColor := color.White
	darkBlue := color.NRGBA{R: 20, G: 40, B: 80, A: 255}

	// --- Верхняя панель (Header) ---
	logo := canvas.NewText("ВШЭ", headerTextColor)
	logo.TextStyle.Bold = true
	logo.TextSize = 28
	logoContainer := container.New(layout.NewMaxLayout(), logo)

	headerTitleText := canvas.NewText("Студенты", headerTextColor)
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

	// --- Таблица с данными ---
	table := widget.NewTable(
		func() (int, int) { return len(works1), 3 }, // Количество строк и столбцов
		func() fyne.CanvasObject {
			// Шаблон для ячейки
			return container.NewVBox(widget.NewLabel(""))
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			row, col := id.Row, id.Col

			// Убедимся, что ячейка — это контейнер
			containerObj, ok := cell.(*fyne.Container)
			if !ok {
				newContainer := container.NewHBox(widget.NewLabel(""))
				cell = newContainer
				containerObj = newContainer
			}

			switch col {
			case 0: // ФИО (не кликабельное)
				label := widget.NewLabel(works1[row].Studentname)
				label.Wrapping = fyne.TextWrapWord
				containerObj.Objects = []fyne.CanvasObject{label}
			case 1: // Почта (не кликабельная)
				label := widget.NewLabel(works1[row].Email)
				label.Wrapping = fyne.TextWrapWord
				containerObj.Objects = []fyne.CanvasObject{label}
			case 2: // Выпадающий список ассистентов
				assistantNames := make([]string, len(assistants1))
				for i, a := range assistants1 {
					assistantNames[i] = a.Name
				}
				selectAssistant := widget.NewSelect(assistantNames, func(selected string) {
					if selected != "" {
						log.Println(time.Now().Format("15:04:05"), "Выбран ассистент для", works[row].Studentname, ":", selected)
					}
				})
				selectAssistant.SetSelectedIndex(-1) // Начальное значение — ничего не выбрано
				containerObj.Objects = []fyne.CanvasObject{selectAssistant}
			}
			cell.Refresh()
		},
	)

	// Настройка таблицы
	table.SetColumnWidth(0, 400)          // Ширина для ФИО
	table.SetColumnWidth(1, 400)          // Ширина для почты
	table.SetColumnWidth(2, 200)          // Ширина для списка ассистентов
	table.Resize(fyne.NewSize(1400, 560)) // Фиксированная ширина и высота для прокрутки

	// Общий контейнер для таблицы с прокруткой
	tableScroll := container.NewScroll(table)
	tableScroll.SetMinSize(fyne.NewSize(1400, 560)) // Устанавливаем минимальный размер для прокрутки

	// --- Кнопка "Далее" с модальным попапом ---
	nextButton := widget.NewButton("Далее", func() {
		allAssigned := true
		var unassignedStudents []string
		for i := 0; i < len(works); i++ {
			// Проверяем, есть ли выпадающий список и выбран ли ассистент
			cell := table.CreateRenderer().Objects()[i*3+2]
			if containerObj, ok := cell.(*fyne.Container); ok {
				selectAssistant := containerObj.Objects[0].(*widget.Select)
				if selectAssistant.Selected == "" {
					allAssigned = false
					unassignedStudents = append(unassignedStudents, works[i].Studentname)
				}
			}
		}
		if !allAssigned {
			// Создаем модальный попап с предупреждением
			popupContent := container.NewVBox(
				widget.NewLabel("Не все ассистенты назначены!"),
				widget.NewLabel("Назначьте ассистентов для: "+strings.Join(unassignedStudents, ", ")),
				widget.NewButton("OK", func() {

				}),
			)
			popup := widget.NewPopUp(popupContent, myWindow.Canvas())
			popup.Show()
		} else {
			// Создаем модальный попап об успешном завершении
			popupContent := container.NewVBox(
				widget.NewLabel("Все ассистенты назначены."),
				widget.NewButton("OK", func() {

					// Здесь можно добавить переход на следующий экран
					log.Println(time.Now().Format("15:04:05"), "Переход к следующему шагу.")
				}),
			)
			popup := widget.NewPopUp(popupContent, myWindow.Canvas())
			popup.Show()
		}
	})

	// --- Центральная секция ---
	assignmentContent := container.NewVBox(tableScroll, nextButton)

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
