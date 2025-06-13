package lector

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type Group struct {
	Name     string
	Criteria []string
}

func Testpage() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Student")

	// Данные групп и критериев
	groups := []Group{
		{Name: "Группа 1", Criteria: []string{"Критерий 1.1", "Критерий 1.2"}},
		{Name: "Группа 2", Criteria: []string{"Критерий 2.1", "Критерий 2.2"}},
	}
	selectedGroupIndex := -1

	// Контейнер для содержимого
	contentContainer := container.New(layout.NewMaxLayout(), widget.NewLabel("Выберите группу и критерий"))

	// Кнопка "Создать" для перехода на следующую страницу
	createButton := widget.NewButton("Создать", func() {
		// Здесь будет логика перехода на следующую страницу после разработки бэкенда
		dialog.ShowInformation("Успех", "Переход на следующую страницу (бэкэнд не реализован)", myWindow)
	})

	// Контейнер для правой части с кнопкой "Создать" внизу
	mainContent := container.NewBorder(nil, createButton, nil, nil, contentContainer)

	// Список групп
	groupList := widget.NewList(
		func() int {
			return len(groups)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Template")
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			item.(*widget.Label).SetText(groups[id].Name)
		},
	)

	// Кнопка добавления группы
	addGroupButton := widget.NewButton("Добавить группу", func() {
		entry := widget.NewEntry()
		dialog.ShowForm("Новая группа", "Создать", "Отмена", []*widget.FormItem{
			{Text: "Название группы", Widget: entry},
		}, func(b bool) {
			if b && entry.Text != "" {
				groups = append(groups, Group{Name: entry.Text, Criteria: []string{}})
				groupList.Refresh()
			}
		}, myWindow)
	})

	// Список критериев
	criteriaList := widget.NewList(
		func() int {
			if selectedGroupIndex >= 0 && selectedGroupIndex < len(groups) {
				return len(groups[selectedGroupIndex].Criteria)
			}
			return 0
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Template")
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if selectedGroupIndex >= 0 && selectedGroupIndex < len(groups) {
				item.(*widget.Label).SetText(groups[selectedGroupIndex].Criteria[id])
			}
		},
	)

	// Кнопка добавления критерия
	addCriterionButton := widget.NewButton("Добавить критерий", func() {
		if selectedGroupIndex < 0 {
			dialog.ShowInformation("Ошибка", "Сначала выберите группу", myWindow)
			return
		}
		entry := widget.NewEntry()
		dialog.ShowForm("Новый критерий", "Создать", "Отмена", []*widget.FormItem{
			{Text: "Название критерия", Widget: entry},
		}, func(b bool) {
			if b && entry.Text != "" {
				groups[selectedGroupIndex].Criteria = append(groups[selectedGroupIndex].Criteria, entry.Text)
				criteriaList.Refresh()
			}
		}, myWindow)
	})

	// Обработчик выбора группы
	groupList.OnSelected = func(id widget.ListItemID) {
		selectedGroupIndex = id
		criteriaList.Refresh()
		contentContainer.Objects = []fyne.CanvasObject{widget.NewLabel("Выберите критерий")}
		contentContainer.Refresh()
	}

	// Обработчик выбора критерия
	criteriaList.OnSelected = func(id widget.ListItemID) {
		// Создаем поля ввода
		entries := make(map[float64]*widget.Entry)
		scores := []float64{0.0, 0.25, 0.50, 0.75, 1.00}
		weightEntry := widget.NewEntry()
		weightEntry.SetPlaceHolder("Вес критерия")

		formItems := []*widget.FormItem{}
		for _, score := range scores {
			entry := widget.NewEntry()
			entries[score] = entry
			entry.SetPlaceHolder(fmt.Sprintf("Комментарий для %.2f", score))
			formItems = append(formItems, &widget.FormItem{
				Text:   fmt.Sprintf("Оценка %.2f", score),
				Widget: entry,
			})
		}
		formItems = append(formItems, &widget.FormItem{
			Text:   "Вес",
			Widget: weightEntry,
		})

		content := container.NewVBox(
			widget.NewForm(formItems...),
		)

		contentContainer.Objects = []fyne.CanvasObject{content}
		contentContainer.Refresh()
	}

	// Макет интерфейса
	groupContainer := container.NewVBox(groupList, addGroupButton)
	criteriaContainer := container.NewVBox(criteriaList, addCriterionButton)
	leftPanel := container.NewHSplit(groupContainer, criteriaContainer)
	leftPanel.SetOffset(0.5)
	split := container.NewHSplit(leftPanel, mainContent)
	split.SetOffset(0.3)

	myWindow.SetContent(split)
	myWindow.Resize(fyne.NewSize(800, 600))
	myWindow.ShowAndRun()
}
