package pages

import (
	"fmt"
	"image/color"
	"time" // Для генерации даты сдачи

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas" // Убедитесь, что этот импорт есть
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// Структура для данных нашей ячейки
type MyListItem struct {
	Name     string
	IsActive bool
	DueDate  string // Добавим поле для даты сдачи
}

func LectorWorksPage() {
	a := app.New()
	w := a.NewWindow("Приложение для работ")
	w.Resize(fyne.NewSize(700, 900)) // Увеличиваем размер для лучшего отображения

	// Данные для нашего списка: изначально только одна ячейка
	data := []MyListItem{
		{"Название работы", true, "2025-06-10"}, // Изначальное название и дата
	}

	// --- Компоненты Заголовка и Боковой панели ---

	// В этой версии мы всегда используем текстовую эмблему.
	// Код для SVG полностью удален.
	logoText := canvas.NewText("ВШЭ", color.White)
	logoText.TextStyle.Bold = true
	logoText.TextSize = 24
	logoText.Alignment = fyne.TextAlignCenter
	// Оборачиваем текст в MaxLayout, чтобы он занимал свое минимальное пространство и был по центру,
	// без вызова SetMinSize, так как fyne.Container не имеет этого метода.
	leftHeaderObject := container.NewMax(logoText)

	// Заголовок "Список работ"
	headerTitle := canvas.NewText("Список работ", color.White)
	headerTitle.TextStyle.Bold = true
	headerTitle.TextSize = 20
	headerTitle.Alignment = fyne.TextAlignCenter

	// Верхняя панель: эмблема слева, заголовок по центру
	headerContent := container.New(layout.NewBorderLayout(nil, nil, leftHeaderObject, nil),
		leftHeaderObject, // Используем текстовую эмблему
		container.NewCenter(headerTitle),
	)

	// Боковая панель без кнопок - просто синяя полоса
	sideBarButtons := container.NewVBox(
		layout.NewSpacer(),
	)

	// --- Список работ (центральная часть) ---
	var myListWidget *widget.List // Объявляем myListWidget здесь

	// createItem: создает шаблон ячейки
	createItem := func() fyne.CanvasObject {
		dueDateLabel := widget.NewLabel("Дата сдачи: ")
		nameLabel := widget.NewLabel("Название")
		nameLabel.TextStyle.Bold = true
		nameLabel.Wrapping = fyne.TextWrapWord // Это позволяет тексту переноситься и увеличивать высоту ячейки

		textGroup := container.NewVBox(dueDateLabel, nameLabel)

		changeButton := widget.NewButton("Изменить", func() {
			// Этот обработчик будет переопределен в updateItem
		})

		return container.NewHBox(textGroup, layout.NewSpacer(), changeButton)
	}

	// updateItem: обновляет содержимое ячейки
	updateItem := func(id widget.ListItemID, item fyne.CanvasObject) {
		cellContainer := item.(*fyne.Container)
		textGroup := cellContainer.Objects[0].(*fyne.Container)
		changeButton := cellContainer.Objects[2].(*widget.Button)

		dueDateLabel := textGroup.Objects[0].(*widget.Label)
		nameLabel := textGroup.Objects[1].(*widget.Label)

		dueDateLabel.SetText("Дата сдачи: " + data[id].DueDate)
		nameLabel.SetText(data[id].Name)
		nameLabel.Refresh()
		dueDateLabel.Refresh()

		currentID := id
		changeButton.OnTapped = func() {
			fmt.Printf("Кнопка 'Изменить' для работы '%s' (ID: %d) нажата! (Будет открыта страница редактирования)\n", data[currentID].Name, currentID)
		}
		changeButton.Refresh()
	}

	myListWidget = widget.NewList(
		func() int {
			return len(data)
		},
		createItem,
		updateItem,
	)

	myListWidget.OnSelected = func(id widget.ListItemID) {
		fmt.Printf("Выбран элемент списка: %d - %s\n", id, data[id].Name)
	}

	// --- Кнопка "Добавить" внизу с логикой добавления ---
	addButton := widget.NewButton("Добавить", func() {
		newDueDate := time.Now().Add(time.Hour * 24 * 7 * time.Duration(len(data)+1)).Format("2006-01-02")
		newItem := MyListItem{
			Name:     "Новая работа",
			IsActive: true,
			DueDate:  newDueDate,
		}
		data = append(data, newItem)
		fmt.Println("Добавлен новый элемент:", newItem.Name)

		myListWidget.Refresh() // Обновляем весь список, чтобы показать новый элемент
	})
	addButtonContainer := container.New(layout.NewHBoxLayout(), layout.NewSpacer(), addButton)

	// Белый фон для списка (центральной части)
	listBackground := canvas.NewRectangle(color.White)
	listWithBackground := container.NewMax(listBackground, myListWidget)

	// --- Главная Компоновка Окна ---
	w.SetContent(container.NewMax(
		canvas.NewRectangle(color.NRGBA{R: 20, G: 40, B: 80, A: 255}), // Единый синий фон для всего окна
		container.NewBorder(
			headerContent,
			addButtonContainer,
			sideBarButtons,
			nil,
			listWithBackground,
		),
	))

	w.ShowAndRun()
}
