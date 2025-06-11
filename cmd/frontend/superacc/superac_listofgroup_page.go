package superacc

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// CriterionEntry представляет набор виджетов для ввода одного критерия.
// Содержит поля ввода для названия, описания, комментария и оценки,
// а также контейнер для их отображения в строке таблицы.
type CriterionEntry struct {
	NameEntry        *widget.Entry
	DescriptionEntry *widget.Entry
	CommentEntry     *widget.Entry
	EvaluationEntry  *widget.Entry
	Container        *fyne.Container
}

// я переделал ShowBlockingCriteriaPage котрый отображает страницу для управления блокирующими критериями.
// Создает окно с заголовком, списком критериев, кнопками "Добавить", "Удалить" и "Далее",
// а также возможностью прокрутки списка критериев.
func ShowBlockingCriteriaPage() {
	a := app.New()
	w := a.NewWindow("Список групп")
	w.Resize(fyne.NewSize(1280, 720))
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
		// Здесь можно добавить логику перехода на предыдущую страницу
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

	// Компонвка страницы: заголовок, кнопка "Назад", центральный контент, нижние кнопки
	w.SetContent(container.NewStack(
		canvas.NewRectangle(color.NRGBA{R: 20, G: 40, B: 80, A: 255}), // Фон окна
		container.NewBorder(
			container.NewVBox(headerContent, backButtonContainer),
			bottomButtonsWithPadding,
			nil,
			nil,
			centralContent,
		),
	))
	w.ShowAndRun()
}
