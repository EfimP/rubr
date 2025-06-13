package lector

import (
	"fmt"
	"image/color"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type GroupEntry struct {
	Name string
}

// AssessmentCriterionEntry представляет запись для списка критериев оценки.
type AssessmentCriterionEntry struct {
	Name string
}

// DescriptionScore represents the description for a specific score.
type DescriptionScore struct {
	Score       float32
	Description string
	Entry       *widget.Entry // Ссылка на виджет Entry для управления цветом
}

// Global data stores for this page
var groups []*GroupEntry
var assessmentCriteria []*AssessmentCriterionEntry
var scoreDescriptions []*DescriptionScore

// Initializing the fixed score descriptions
func init() {
	scores := []float32{0.0, 0.25, 0.5, 0.75, 1.0}
	for _, score := range scores {
		scoreDescriptions = append(scoreDescriptions, &DescriptionScore{
			Score:       score,
			Description: "", // Initially empty
		})
	}

	// Add some initial data for demonstration
	groups = append(groups, &GroupEntry{Name: "Группа 1"})
	groups = append(groups, &GroupEntry{Name: "Группа 2"})

	assessmentCriteria = append(assessmentCriteria, &AssessmentCriterionEntry{Name: "Критерий А"})
	assessmentCriteria = append(assessmentCriteria, &AssessmentCriterionEntry{Name: "Критерий Б"})
}

// updateGroupsListUI updates the UI for the list of groups.
var updateGroupsListUI func()

// updateCriteriaListUI updates the UI for the list of assessment criteria.
var updateCriteriaListUI func()

// updateScoreDescriptionsUI updates the UI for the score descriptions.
var updateScoreDescriptionsUI func()

// ShowMainCriteriaPage displays the "Основные критерии" page.
func ShowMainCriteriaPage() {
	a := app.New()
	a.Settings().SetTheme(theme.LightTheme()) // Устанавливаем постоянную белую тему

	w := a.NewWindow("Супер-акк: Основные критерии")
	w.Resize(fyne.NewSize(1200, 720))

	// Цвета
	headerTextColor := color.White
	darkBlue := color.NRGBA{R: 20, G: 40, B: 80, A: 255}
	mediumGrayDivider := color.NRGBA{R: 180, G: 180, B: 180, A: 255}

	// --- Верхняя панель (Header) ---
	logo := canvas.NewText("ВШЭ", headerTextColor) // Логотип
	logo.TextStyle.Bold = true
	logo.TextSize = 28
	logoContainer := container.NewMax(logo)

	headerTitleText := canvas.NewText("Основные критерии", headerTextColor)
	headerTitleText.TextStyle.Bold = true
	headerTitleText.TextSize = 24
	headerTitleText.Alignment = fyne.TextAlignCenter

	header := container.New(layout.NewBorderLayout(nil, nil, logoContainer, nil),
		container.NewPadded(logoContainer),
		container.NewCenter(headerTitleText),
	)
	headerBackground := canvas.NewRectangle(darkBlue)
	headerWithBackground := container.NewStack(headerBackground, header)

	// --- Левая боковая панель ---
	sidePanelBackground := canvas.NewRectangle(headerTextColor)

	sidePanel := container.NewVBox(
		layout.NewSpacer(),
	)
	//sidePanel.SetMinSize(fyne.NewSize(60, 0))
	sidePanelWithBackground := container.NewStack(sidePanelBackground, sidePanel)

	// --- Кнопка "назад" ---
	backButton := widget.NewButton("назад", func() {
		log.Println("Кнопка 'назад' нажата. Возврат на предыдущую страницу.")
		w.Close() // Пример: закрыть текущее окно
	})
	backButtonRow := container.NewHBox(backButton, layout.NewSpacer())

	// --- Первая вертикальная часть: Список групп ---
	groupsListContainer := container.NewVBox()
	groupsScroll := container.NewVScroll(groupsListContainer)
	groupsScroll.SetMinSize(fyne.NewSize(200, 400)) // Примерный размер

	updateGroupsListUI = func() {
		groupsListContainer.RemoveAll()
		groupsListContainer.Add(widget.NewLabelWithStyle("список групп и их названия", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}))
		for i, g := range groups {
			entry := widget.NewEntry()
			entry.SetText(g.Name)
			// Обновляем данные при изменении поля
			idx := i // Capture index for closure
			entry.OnChanged = func(s string) {
				groups[idx].Name = s
			}
			groupsListContainer.Add(container.NewPadded(entry))
		}
		groupsListContainer.Refresh()
	}
	updateGroupsListUI() // Initial draw

	addGroupButton := widget.NewButton("добавить", func() {
		groups = append(groups, &GroupEntry{Name: ""})
		updateGroupsListUI()
	})
	deleteGroupButton := widget.NewButton("удалить", func() {
		if len(groups) > 0 {
			groups = groups[:len(groups)-1] // Удаляем последний
			updateGroupsListUI()
		}
	})
	groupButtons := container.New(layout.NewHBoxLayout(), addGroupButton, deleteGroupButton)
	groupSection := container.NewVBox(
		groupsScroll,
		groupButtons,
	)

	// --- Вторая вертикальная часть: Список критериев ---
	criteriaListContainer := container.NewVBox()
	criteriaScroll := container.NewVScroll(criteriaListContainer)
	criteriaScroll.SetMinSize(fyne.NewSize(200, 400)) // Примерный размер

	updateCriteriaListUI = func() {
		criteriaListContainer.RemoveAll()
		criteriaListContainer.Add(widget.NewLabelWithStyle("список критериев и их названия", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}))
		for i, c := range assessmentCriteria {
			entry := widget.NewEntry()
			entry.SetText(c.Name)
			idx := i // Capture index
			entry.OnChanged = func(s string) {
				assessmentCriteria[idx].Name = s
			}
			criteriaListContainer.Add(container.NewPadded(entry))
		}
		criteriaListContainer.Refresh()
	}
	updateCriteriaListUI() // Initial draw

	addCriterionButton := widget.NewButton("добавить", func() {
		assessmentCriteria = append(assessmentCriteria, &AssessmentCriterionEntry{Name: ""})
		updateCriteriaListUI()
	})
	deleteCriterionButton := widget.NewButton("удалить", func() {
		if len(assessmentCriteria) > 0 {
			assessmentCriteria = assessmentCriteria[:len(assessmentCriteria)-1] // Удаляем последний
			updateCriteriaListUI()
		}
	})
	criterionButtons := container.New(layout.NewHBoxLayout(), addCriterionButton, deleteCriterionButton)
	criterionSection := container.NewVBox(
		criteriaScroll,
		criterionButtons,
	)

	// --- Третья вертикальная часть: Описания оценок ---
	scoreDescriptionsContainer := container.NewVBox()
	scoreDescriptionsScroll := container.NewVScroll(scoreDescriptionsContainer)
	scoreDescriptionsScroll.SetMinSize(fyne.NewSize(250, 300)) // Примерный размер

	updateScoreDescriptionsUI = func() {
		scoreDescriptionsContainer.RemoveAll()
		for _, ds := range scoreDescriptions {
			label := widget.NewLabel(fmt.Sprintf("%.2f", ds.Score)) // 0.0, 0.25, etc.
			entry := widget.NewEntry()
			entry.SetText(ds.Description)
			entry.SetPlaceHolder("описание")

			// Сохраняем ссылку на entry в структуре для управления цветом
			ds.Entry = entry

			entry.OnChanged = func(s string) {
				ds.Description = s
				if s == "" {
					entry.TextStyle.Bold = true
				} else {
					entry.TextStyle.Monospace = true
				}
				entry.Refresh()
			}
			// Устанавливаем начальный цвет, если описание пустое
			if ds.Description == "" {
				entry.TextStyle.Bold = true
			}

			scoreDescriptionsContainer.Add(
				container.New(layout.NewVBoxLayout(),
					label,
					container.NewMax(entry), // Растягивает поле ввода
				),
			)
		}
		scoreDescriptionsContainer.Refresh()
	}
	updateScoreDescriptionsUI() // Initial draw

	// --- Поле ввода "вес" и кнопка "Создать" внизу ---
	weightEntry := widget.NewEntry()
	weightEntry.SetPlaceHolder("введите вес")
	weightEntry.SetText("1.0")

	weightLabel := widget.NewLabel("вес")
	weightInput := container.New(layout.NewHBoxLayout(), weightLabel, weightEntry)
	createButton := widget.NewButton("Создать", func() {
		log.Println("Кнопка 'Создать' нажата. Сохраняем все данные.")
		log.Println("Текущий вес:", weightEntry.Text)
		log.Println("Группы:", groups)
		log.Println("Критерии:", assessmentCriteria)
		log.Println("Описания оценок:")
		for _, ds := range scoreDescriptions {
			log.Printf("  %.2f: '%s'\n", ds.Score, ds.Description)
		}
	})

	bottomButtons := container.New(layout.NewHBoxLayout(),
		weightInput,
		layout.NewSpacer(), // Растягивает пространство
		createButton,
	)

	// --- Основной контент страницы (три вертикальные части) ---
	mainContentGrid := container.New(layout.NewHBoxLayout(), // HBoxLayout для трех частей
		container.NewPadded(groupSection),
		canvas.NewRectangle(mediumGrayDivider), // Разделитель
		container.NewPadded(criterionSection),
		canvas.NewRectangle(mediumGrayDivider),       // Разделитель
		container.NewPadded(scoreDescriptionsScroll), // Третья часть - только прокручиваемая область
	)
	// Добавляем Spacer, чтобы растянуть части по горизонтали равномерно
	mainContentWithSpacer := container.New(layout.NewHBoxLayout(), mainContentGrid, layout.NewSpacer())

	centralContentPanel := container.NewVBox(
		container.NewPadded(backButtonRow),
		mainContentWithSpacer, // Основной контент с тремя частями
		container.NewPadded(bottomButtons),
	)

	contentBackground := canvas.NewRectangle(color.White)
	centralContentWithBackground := container.NewStack(contentBackground, container.NewPadded(centralContentPanel))

	// --- Общая компоновка окна ---
	w.SetContent(container.NewBorder(
		headerWithBackground,         // Верхняя панель
		nil,                          // Нижняя панель
		sidePanelWithBackground,      // Левая боковая панель
		nil,                          // Правая панель
		centralContentWithBackground, // Центральный контент
	))

	w.ShowAndRun()
}
