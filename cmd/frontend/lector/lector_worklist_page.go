package lector

import (
	"fmt"
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog" // Для диалога подтверждения.
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// MyListItem представляет данные для одной записи в списке работ.
type MyListItem struct {
	Name     string
	IsActive bool
	DueDate  string
}

// LectorWorksPage создает и отображает страницу со списком работ лектора.
func LectorWorksPage() {
	a := app.New()
	// Установка светлой темы Fyne, чтобы оформление не менялось с темой ОС.
	a.Settings().SetTheme(theme.LightTheme())

	w := a.NewWindow("Приложение для работ")
	w.Resize(fyne.NewSize(700, 900))

	// Данные для списка работ.
	var data []MyListItem
	data = []MyListItem{
		{"Название работы 1", true, "2025-06-10"},
	}

	// --- Компоненты Заголовка ---
	headerTextColor := color.White
	logoText := canvas.NewText("ВШЭ", headerTextColor)
	logoText.TextStyle.Bold = true
	logoText.TextSize = 24
	logoText.Alignment = fyne.TextAlignCenter
	leftHeaderObject := container.NewMax(logoText)

	headerTitle := canvas.NewText("Список работ", headerTextColor)
	headerTitle.TextStyle.Bold = true
	headerTitle.TextSize = 20
	headerTitle.Alignment = fyne.TextAlignCenter

	headerContent := container.New(layout.NewBorderLayout(nil, nil, leftHeaderObject, nil),
		leftHeaderObject,
		container.NewCenter(headerTitle),
	)

	// Боковая панель (пустая).
	sideBarButtons := container.NewVBox(
		layout.NewSpacer(),
	)

	// --- Список работ (центральная часть) ---
	var myListWidget *widget.List

	// createItem создает макет для одной ячейки списка.
	createItem := func() fyne.CanvasObject {
		dueDateLabel := widget.NewLabel("Дата сдачи: ")
		nameLabel := widget.NewLabel("Название")
		nameLabel.TextStyle.Bold = true
		nameLabel.Wrapping = fyne.TextWrapWord

		textGroup := container.NewVBox(dueDateLabel, nameLabel)

		changeButton := widget.NewButton("Изменить", func() {})
		deleteButton := widget.NewButton("Удалить", func() {})

		return container.NewHBox(textGroup, layout.NewSpacer(), changeButton, deleteButton)
	}

	// updateItem обновляет содержимое ячейки списка данными из `data`.
	updateItem := func(id widget.ListItemID, item fyne.CanvasObject) {
		cellContainer := item.(*fyne.Container)
		textGroup := cellContainer.Objects[0].(*fyne.Container)
		changeButton := cellContainer.Objects[2].(*widget.Button)
		deleteButton := cellContainer.Objects[3].(*widget.Button)

		dueDateLabel := textGroup.Objects[0].(*widget.Label)
		nameLabel := textGroup.Objects[1].(*widget.Label)

		dueDateLabel.SetText("Дата сдачи: " + data[id].DueDate)
		nameLabel.SetText(data[id].Name)
		nameLabel.Refresh()
		dueDateLabel.Refresh()

		currentID := id
		changeButton.OnTapped = func() {
			fmt.Printf("Кнопка 'Изменить' для работы '%s' (ID: %d) нажата!\n", data[currentID].Name, currentID)
		}
		changeButton.Refresh()

		deleteButton.OnTapped = func() {
			// Диалог подтверждения удаления.
			dialog.ShowConfirm(
				"Подтверждение удаления",
				fmt.Sprintf("Вы уверены, что хотите удалить работу '%s'?", data[currentID].Name),
				func(confirmed bool) {
					if confirmed {
						fmt.Printf("Подтверждено удаление работы '%s' (ID: %d).\n", data[currentID].Name, currentID)
						// Удаление элемента из данных и обновление списка.
						data = append(data[:currentID], data[currentID+1:]...)
						myListWidget.Refresh()
					} else {
						fmt.Printf("Удаление работы '%s' (ID: %d) отменено.\n", data[currentID].Name, currentID)
					}
				},
				w, // Окно, к которому привязан диалог.
			)
		}
		deleteButton.Refresh()
	}

	// Инициализация виджета списка.
	myListWidget = widget.NewList(
		func() int { return len(data) },
		createItem,
		updateItem,
	)

	// Обработчик выбора элемента списка.
	myListWidget.OnSelected = func(id widget.ListItemID) {
		fmt.Printf("Выбран элемент списка: %d - %s\n", id, data[id].Name)
	}

	// --- Кнопка "Добавить" ---
	addButton := widget.NewButton("Добавить", func() {
		newDueDate := time.Now().Add(time.Hour * 24 * 7 * time.Duration(len(data)+1)).Format("2006-01-02")
		newItem := MyListItem{
			Name:     fmt.Sprintf("Новая работа %d", len(data)+1),
			IsActive: true,
			DueDate:  newDueDate,
		}
		data = append(data, newItem)
		fmt.Println("Добавлен новый элемент:", newItem.Name)
		myListWidget.Refresh() // Обновление списка.
	})
	// Контейнер для кнопки "Добавить".
	addButtonContainer := container.New(layout.NewHBoxLayout(), layout.NewSpacer(), addButton)

	// --- Фон списка ---
	listBackground := canvas.NewRectangle(color.White)
	listWithBackground := container.NewMax(listBackground, myListWidget)

	// --- Главная Компоновка Окна ---
	w.SetContent(container.NewMax(
		canvas.NewRectangle(color.NRGBA{R: 20, G: 40, B: 80, A: 255}), // Общий фон окна.
		container.NewBorder(
			headerContent,      // Верхняя часть.
			addButtonContainer, // Нижняя часть.
			sideBarButtons,     // Левая панель.
			nil,                // Правая панель.
			listWithBackground, // Центральное содержимое.
		),
	))

	w.ShowAndRun()
}
