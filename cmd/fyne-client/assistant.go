package main

import (
	"context"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"image/color"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc"
	gradingpb "rubr/proto/grade"
	rubricpb "rubr/proto/rubric"
	workassignmentpb "rubr/proto/workassignment"
)

type AssistantWorkItem struct {
	WorkID          int32
	TaskID          int32
	TaskTitle       string
	StudentEmail    string
	StudentFullName string
}

type BlockingCriterionEntry struct {
	CriterionID     int32
	CommentEntry    *widget.Entry
	EvaluationEntry *widget.Entry
	Container       *fyne.Container
}

type MainCriterionEntry struct {
	CriterionID  int32
	Select       *widget.Select
	CommentEntry *widget.Entry
}

func CreateAssistantWorksPage(state *AppState, leftBackground *canvas.Image) fyne.CanvasObject {
	userIDint64, err := strconv.ParseInt(state.userID, 10, 32)
	if err != nil {
		log.Printf("Некорректный ID пользователя: %v", err)
		return container.NewVBox(widget.NewLabel("Ошибка: некорректный ID пользователя"))
	}
	userID := int32(userIDint64)

	conn, err := grpc.Dial("localhost:50054", grpc.WithInsecure())
	if err != nil {
		log.Printf("Не удалось подключиться к сервису назначений работ: %v", err)
		return container.NewVBox(widget.NewLabel("Ошибка подключения к сервису"))
	}
	defer conn.Close()

	client := workassignmentpb.NewWorkAssignmentServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// получение списка работ
	resp, err := client.GetWorksForAssistant(ctx, &workassignmentpb.GetWorksForAssistantRequest{AssistantId: userID})
	if err != nil {
		log.Printf("Не удалось получить работы: %v", err)
		return container.NewVBox(widget.NewLabel("Ошибка загрузки работ"))
	}
	if resp.Error != "" {
		log.Printf("Ошибка от сервиса: %s", resp.Error)
		return container.NewVBox(widget.NewLabel(resp.Error))
	}

	// обработка списка работ
	var data []AssistantWorkItem
	for _, work := range resp.Works {
		fullName := work.StudentName + " " + work.StudentSurname
		if work.StudentPatronymic != "" {
			fullName += " " + work.StudentPatronymic
		}
		data = append(data, AssistantWorkItem{
			WorkID:          work.WorkId,
			TaskID:          work.TaskId,
			TaskTitle:       work.TaskTitle,
			StudentEmail:    work.StudentEmail,
			StudentFullName: fullName,
		})
	}

	headerTextColor := color.White
	logoText := canvas.NewText("ВШЭ", headerTextColor)
	logoText.TextStyle.Bold = true
	logoText.TextSize = 24
	logoText.Alignment = fyne.TextAlignCenter
	leftHeaderObject := container.NewMax(logoText)

	headerTitle := canvas.NewText("Работы на проверку", headerTextColor)
	headerTitle.TextStyle.Bold = true
	headerTitle.TextSize = 20
	headerTitle.Alignment = fyne.TextAlignCenter

	headerContent := container.New(layout.NewBorderLayout(nil, nil, leftHeaderObject, nil),
		leftHeaderObject,
		container.NewCenter(headerTitle),
	)

	// вывод списка работ
	var myListWidget *widget.List
	createItem := func() fyne.CanvasObject {
		taskTitleLabel := widget.NewLabel("Задание: ")
		taskTitleLabel.TextStyle.Bold = true
		studentEmailLabel := widget.NewLabel("Email студента: ")
		studentNameLabel := widget.NewLabel("Студент: ")
		return container.NewVBox(taskTitleLabel, studentEmailLabel, studentNameLabel)
	}

	updateItem := func(id widget.ListItemID, item fyne.CanvasObject) {
		vbox := item.(*fyne.Container)
		taskTitleLabel := vbox.Objects[0].(*widget.Label)
		studentEmailLabel := vbox.Objects[1].(*widget.Label)
		studentNameLabel := vbox.Objects[2].(*widget.Label)

		taskTitleLabel.SetText("Задание: " + data[id].TaskTitle)
		studentEmailLabel.SetText("Email студента: " + data[id].StudentEmail)
		studentNameLabel.SetText("Студент: " + data[id].StudentFullName)
	}

	myListWidget = widget.NewList(
		func() int { return len(data) },
		createItem,
		updateItem,
	)

	// Обработчик выбора элемента списка
	myListWidget.OnSelected = func(id widget.ListItemID) {
		workID := data[id].WorkID
		taskID := data[id].TaskID
		state.currentPage = "assistant_work_details"
		state.window.SetContent(CreateAssistantWorkDetailsPage(state, workID, taskID, leftBackground))
		myListWidget.UnselectAll()
	}

	listBackground := canvas.NewRectangle(color.White)
	listWithBackground := container.NewMax(listBackground, myListWidget)

	return container.NewMax(
		canvas.NewRectangle(color.NRGBA{R: 20, G: 40, B: 80, A: 255}),
		container.NewBorder(
			headerContent,
			nil,
			nil,
			nil,
			listWithBackground,
		),
	)
}

func CreateAssistantWorkDetailsPage(state *AppState, workID int32, taskID int32, leftBackground *canvas.Image) fyne.CanvasObject {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.Dial("localhost:50054", grpc.WithInsecure())
	if err != nil {
		log.Printf("Не удалось подключиться к сервису: %v", err)
		return container.NewVBox(widget.NewLabel("Ошибка подключения к сервису"))
	}
	defer conn.Close()

	client := workassignmentpb.NewWorkAssignmentServiceClient(conn)
	resp, err := client.GetWorkDetails(ctx, &workassignmentpb.GetWorkDetailsRequest{WorkId: workID})
	if err != nil {
		log.Printf("Не удалось получить детали работы: %v", err)
		return container.NewVBox(widget.NewLabel("Ошибка загрузки деталей работы"))
	}
	if resp.Error != "" {
		log.Printf("Ошибка от сервиса: %s", resp.Error)
		return container.NewVBox(widget.NewLabel(resp.Error))
	}

	// Настройка заголовка
	headerTextColor := color.White
	logoText := canvas.NewText("ВШЭ", headerTextColor)
	logoText.TextStyle.Bold = true
	logoText.TextSize = 24
	logoText.Alignment = fyne.TextAlignCenter
	leftHeaderObject := container.NewStack(logoText)

	headerTitle := canvas.NewText("Задание", headerTextColor)
	headerTitle.TextStyle.Bold = true
	headerTitle.TextSize = 20
	headerTitle.Alignment = fyne.TextAlignCenter

	headerContent := container.New(layout.NewBorderLayout(nil, nil, leftHeaderObject, nil),
		leftHeaderObject,
		container.NewCenter(headerTitle),
	)

	titleLabel := widget.NewLabel("Название: " + resp.TaskTitle)
	titleLabel.TextStyle.Bold = true

	descriptionEntry := widget.NewMultiLineEntry()
	descriptionEntry.SetText(resp.TaskDescription)
	descriptionEntry.Disable()
	scrollableDescription := container.NewVScroll(descriptionEntry)
	scrollableDescription.SetMinSize(fyne.NewSize(0, descriptionEntry.MinSize().Height*5))

	deadlineTime, err := time.Parse(time.RFC3339, resp.TaskDeadline)
	if err != nil {
		log.Printf("Ошибка парсинга дедлайна: %v", err)
		return container.NewVBox(widget.NewLabel("Ошибка обработки дедлайна"))
	}
	deadlineLabel := widget.NewLabel("Дедлайн: " + deadlineTime.Format("02.01.2006 15:04"))

	createdAtTime, err := time.Parse(time.RFC3339, resp.CreatedAt)
	if err != nil {
		log.Printf("Ошибка парсинга даты сдачи: %v", err)
		return container.NewVBox(widget.NewLabel("Ошибка обработки даты сдачи"))
	}
	createdAtLabel := widget.NewLabel("Дата сдачи: " + createdAtTime.Format("02.01.2006 15:04"))

	statusLabel := widget.NewLabel("Статус: " + resp.Status)

	downloadButton := widget.NewButton("Загрузить работу", func() {
		if resp.ContentUrl == "" {
			dialog.ShowInformation("Ошибка", "Ссылка на работу отсутствует", state.window)
			return
		}

		parsedURL, err := url.Parse(resp.ContentUrl)
		if err != nil {
			log.Printf("Некорректная ссылка: %v", err)
			dialog.ShowError(err, state.window)
			return
		}

		linkEntry := widget.NewEntry()
		linkEntry.SetText(parsedURL.String())
		linkEntry.Disable()

		copyButton := widget.NewButton("Копировать", func() {
			state.window.Clipboard().SetContent(linkEntry.Text)
			dialog.ShowInformation("Успех", "Ссылка скопирована в буфер обмена", state.window)
		})

		dialogContent := container.NewVBox(
			widget.NewLabel("Ссылка на работу: "),
			linkEntry,
			container.NewHBox(copyButton),
		)

		dialog.ShowCustom("Ссылка на работу", "Закрыть", dialogContent, state.window)
	})

	gradeButton := widget.NewButton("Оценить", func() {
		state.currentPage = "grading_blocking"
		state.window.SetContent(CreateBlockingCriteriaGradingPage(state, workID, taskID, leftBackground))
	})

	backButton := widget.NewButton("Назад", func() {
		state.currentPage = "assistant_works"
		state.window.SetContent(CreateAssistantWorksPage(state, leftBackground))
	})

	buttonsContainer := container.NewHBox(backButton, layout.NewSpacer(), downloadButton, gradeButton)

	inputGrid := container.NewVBox(
		titleLabel,
		scrollableDescription,
		deadlineLabel,
		createdAtLabel,
		statusLabel,
	)

	contentBackground := canvas.NewRectangle(color.White)
	contentWithPadding := container.NewPadded(inputGrid)
	centralContent := container.NewStack(contentBackground, contentWithPadding)

	return container.NewStack(
		canvas.NewRectangle(color.NRGBA{R: 20, G: 40, B: 80, A: 255}),
		container.NewBorder(
			headerContent,
			buttonsContainer,
			nil,
			nil,
			centralContent,
		),
	)
}

func CreateBlockingCriteriaGradingPage(state *AppState, workID int32, taskID int32, leftBackground *canvas.Image) fyne.CanvasObject {
	w := state.window

	// Подключение к gRPC сервису для загрузки данных
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rubricConn, err := grpc.Dial("localhost:50055", grpc.WithInsecure())
	if err != nil {
		log.Printf("Failed to connect to rubricservice: %v", err)
		return container.NewVBox(widget.NewLabel("Ошибка подключения к сервису"))
	}
	defer rubricConn.Close()

	rubricClient := rubricpb.NewRubricServiceClient(rubricConn)

	// Загрузка блокирующих критериев
	resp, err := rubricClient.LoadTaskBlockingCriterias(ctx, &rubricpb.LoadTaskBlockingCriteriasRequest{TaskId: taskID})
	if err != nil {
		log.Printf("Failed to load blocking criteria: %v", err)
		return container.NewVBox(widget.NewLabel("Ошибка загрузки критериев: " + err.Error()))
	}
	if resp.Error != "" {
		return container.NewVBox(widget.NewLabel("Ошибка: " + resp.Error))
	}

	// Заголовок
	headerTextColor := color.White
	logoText := canvas.NewText("ВШЭ", headerTextColor)
	logoText.TextStyle.Bold = true
	logoText.TextSize = 24
	logoText.Alignment = fyne.TextAlignCenter
	leftHeaderObject := container.NewStack(logoText)

	headerTitle := canvas.NewText("Блокирующие критерии", headerTextColor)
	headerTitle.TextStyle.Bold = true
	headerTitle.TextSize = 20
	headerTitle.Alignment = fyne.TextAlignCenter

	headerContent := container.New(layout.NewBorderLayout(nil, nil, leftHeaderObject, nil),
		leftHeaderObject,
		container.NewCenter(headerTitle),
	)

	backButton := widget.NewButton("Назад", func() {
		state.currentPage = "assistant_work_details"
		w.SetContent(CreateAssistantWorkDetailsPage(state, workID, taskID, leftBackground))
	})
	backButtonContainer := container.NewHBox(layout.NewSpacer(), backButton)

	// Колонки таблицы
	criteriaListContainer := container.NewVBox()
	columnHeaders := container.New(layout.NewGridLayoutWithColumns(4),
		container.NewPadded(widget.NewLabelWithStyle("Название критерия", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
		container.NewPadded(widget.NewLabelWithStyle("Описание критерия", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
		container.NewPadded(widget.NewLabelWithStyle("Комментарий", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
		container.NewPadded(widget.NewLabelWithStyle("Финальная оценка", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
	)

	// Список критериев
	var activeCriteria []*BlockingCriterionEntry
	for _, crit := range resp.Criteria {
		// Поля только для чтения
		nameEntry := widget.NewLabel(crit.Name)
		nameEntryContainer := container.NewMax(nameEntry)
		nameEntryContainer.Resize(fyne.NewSize(250, 60))

		descriptionEntry := widget.NewLabel(crit.Description)
		descriptionEntryContainer := container.NewMax(descriptionEntry)
		descriptionEntryContainer.Resize(fyne.NewSize(250, 60))

		// Редактируемое поле комментария
		commentEntry := widget.NewEntry()
		commentEntry.SetText(crit.Comment) // Заполняем комментарий из базы
		commentEntryContainer := container.NewMax(commentEntry)
		commentEntryContainer.Resize(fyne.NewSize(250, 60))

		// Редактируемое поле оценки
		evaluationEntry := widget.NewEntry()
		evaluationEntry.SetText(strconv.FormatInt(crit.FinalMark, 10)) // Заполняем финальную оценку из базы
		evaluationEntryContainer := container.NewMax(evaluationEntry)
		evaluationEntryContainer.Resize(fyne.NewSize(200, 60))

		criterionRow := container.New(layout.NewGridLayoutWithColumns(4),
			container.NewPadded(nameEntryContainer),
			container.NewPadded(descriptionEntryContainer),
			container.NewPadded(commentEntryContainer),
			container.NewPadded(evaluationEntryContainer),
		)

		criteriaListContainer.Add(criterionRow)
		criteriaListContainer.Refresh()

		activeCriteria = append(activeCriteria, &BlockingCriterionEntry{
			CriterionID:     crit.Id,
			CommentEntry:    commentEntry,
			EvaluationEntry: evaluationEntry,
			Container:       criterionRow,
		})
	}

	listLabel := widget.NewLabelWithStyle("Список блокирующих критериев", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	scrollableCriteria := container.NewVScroll(criteriaListContainer)
	scrollableCriteria.SetMinSize(fyne.NewSize(0, 400))

	// Кнопка "Далее"
	nextButton := widget.NewButton("Далее", func() {
		// Создаем новый контекст и соединение для сохранения данных
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		gradingConn, err := grpc.Dial("localhost:50057", grpc.WithInsecure())
		if err != nil {
			log.Printf("Failed to connect to rubricservice: %v", err)
			dialog.ShowError(err, w)
			return
		}
		defer gradingConn.Close()

		gradingClient := gradingpb.NewGradingServiceClient(gradingConn)

		// Проверка ввода и сохранение оценок
		for _, criterion := range activeCriteria {
			if criterion.EvaluationEntry.Text == "" {
				dialog.ShowInformation("Ошибка", "Введите оценку для всех критериев", w)
				return
			}
			finalMark, err := strconv.ParseFloat(criterion.EvaluationEntry.Text, 32)
			if err != nil {
				dialog.ShowInformation("Ошибка", "Оценка должна быть числом", w)
				return
			}
			log.Printf("Saving mark for criterion ID %d: mark=%f, comment=%s", criterion.CriterionID, finalMark, criterion.CommentEntry.Text)
			resp, err := gradingClient.SetBlockingCriteriaMark(ctx, &gradingpb.SetBlockingCriteriaMarkRequest{
				WorkId:      workID,
				CriterionId: criterion.CriterionID,
				Mark:        float32(finalMark),
				Comment:     criterion.CommentEntry.Text,
			})
			if err != nil {
				log.Printf("Failed to save blocking criteria mark for criterion ID %d: %v", criterion.CriterionID, err)
				dialog.ShowError(err, w)
				return
			}
			if resp.Error != "" {
				log.Printf("SetBlockingCriteriaMark error for criterion ID %d: %s", criterion.CriterionID, resp.Error)
				dialog.ShowInformation("Ошибка", resp.Error, w)
				return
			}
		}
		log.Printf("All blocking criteria marks saved successfully")
		state.currentPage = "grading_main"
		w.SetContent(CreateMainCriteriaGradingPage(state, workID, taskID, leftBackground))
	})

	// Нижняя панель с кнопкой "Далее"
	bottomButtons := container.New(layout.NewHBoxLayout(),
		layout.NewSpacer(),
		nextButton,
	)
	bottomButtonsWithPadding := container.NewPadded(bottomButtons)

	// Сборка интерфейса
	contentBackground := canvas.NewRectangle(color.White)
	criteriaPanel := container.NewVBox(
		container.NewPadded(columnHeaders),
		listLabel,
		scrollableCriteria,
	)
	centralContent := container.NewStack(
		contentBackground,
		container.NewPadded(criteriaPanel),
	)

	return container.NewStack(
		canvas.NewRectangle(color.NRGBA{R: 20, G: 40, B: 80, A: 255}),
		container.NewBorder(
			container.NewVBox(headerContent, backButtonContainer),
			bottomButtonsWithPadding,
			nil,
			nil,
			centralContent,
		),
	)
}

func CreateMainCriteriaGradingPage(state *AppState, workID int32, taskID int32, leftBackground *canvas.Image) fyne.CanvasObject {
	w := state.window

	// Подключение к gRPC сервису
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.Dial("localhost:50055", grpc.WithInsecure())
	if err != nil {
		log.Printf("Ошибка подключения к RubricService: %v", err)
		return container.NewVBox(widget.NewLabel("Ошибка подключения к сервису"))
	}
	defer conn.Close()

	client := rubricpb.NewRubricServiceClient(conn)

	// Загрузка основных критериев
	resp, err := client.LoadTaskMainCriterias(ctx, &rubricpb.LoadTaskMainCriteriasRequest{TaskId: taskID})
	if err != nil || resp.Error != "" {
		errorMsg := resp.Error
		if err != nil {
			errorMsg = err.Error()
		}
		log.Printf("Ошибка загрузки основных критериев: %s", errorMsg)
		return container.NewVBox(widget.NewLabel("Ошибка загрузки критериев: " + errorMsg))
	}

	// Заголовок
	headerTextColor := color.White
	logoText := canvas.NewText("ВШЭ", headerTextColor)
	logoText.TextStyle.Bold = true
	logoText.TextSize = 24
	logoText.Alignment = fyne.TextAlignCenter
	leftHeaderObject := container.NewStack(logoText)

	headerTitle := canvas.NewText("Основные критерии", headerTextColor)
	headerTitle.TextStyle.Bold = true
	headerTitle.TextSize = 20
	headerTitle.Alignment = fyne.TextAlignCenter

	headerContent := container.New(layout.NewBorderLayout(nil, nil, leftHeaderObject, nil),
		leftHeaderObject,
		container.NewCenter(headerTitle),
	)
	headerContent = container.NewStack(canvas.NewRectangle(color.NRGBA{R: 20, G: 40, B: 80, A: 255}), headerContent)

	// Список групп и критериев
	selectedGroupIndex := -1
	var entries []MainCriterionEntry
	// Карта для отслеживания всех критериев с их группами и именами
	type criterionInfo struct {
		groupName string
		critName  string
	}
	criteriaInfo := make(map[int32]criterionInfo)
	totalCriteriaCount := 0

	// Инициализация entries для всех критериев
	for _, group := range resp.Groups {
		for _, crit := range group.Criteria {
			commentEntry := widget.NewEntry()
			selectOptions := []string{"0.0", "0.25", "0.50", "0.75", "1.00"}
			selectWidget := widget.NewSelect(selectOptions, func(s string) {
				switch s {
				case "0.0":
					commentEntry.SetText(crit.Comment_000)
				case "0.25":
					commentEntry.SetText(crit.Comment_025)
				case "0.50":
					commentEntry.SetText(crit.Comment_050)
				case "0.75":
					commentEntry.SetText(crit.Comment_075)
				case "1.00":
					commentEntry.SetText(crit.Comment_100)
				}
			})

			entries = append(entries, MainCriterionEntry{
				CriterionID:  crit.Id,
				Select:       selectWidget,
				CommentEntry: commentEntry,
			})
			criteriaInfo[crit.Id] = criterionInfo{
				groupName: group.GroupName,
				critName:  crit.Name,
			}
			totalCriteriaCount++
		}
	}

	log.Printf("Initialized %d criteria entries", len(entries))

	contentContainer := container.New(layout.NewMaxLayout(), widget.NewLabel("Выберите группу и критерий"))

	groupList := widget.NewList(
		func() int {
			return len(resp.Groups)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Template")
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			item.(*widget.Label).SetText(resp.Groups[id].GroupName)
		},
	)

	criteriaList := widget.NewList(
		func() int {
			if selectedGroupIndex >= 0 && selectedGroupIndex < len(resp.Groups) {
				return len(resp.Groups[selectedGroupIndex].Criteria)
			}
			return 0
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Template")
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if selectedGroupIndex >= 0 && selectedGroupIndex < len(resp.Groups) {
				item.(*widget.Label).SetText(resp.Groups[selectedGroupIndex].Criteria[id].Name)
			}
		},
	)

	groupList.OnSelected = func(id widget.ListItemID) {
		selectedGroupIndex = id
		criteriaList.Refresh()
		contentContainer.Objects = []fyne.CanvasObject{widget.NewLabel("Выберите критерий")}
		contentContainer.Refresh()
	}

	criteriaList.OnSelected = func(id widget.ListItemID) {
		if selectedGroupIndex >= 0 && selectedGroupIndex < len(resp.Groups) && id >= 0 && id < len(resp.Groups[selectedGroupIndex].Criteria) {
			crit := resp.Groups[selectedGroupIndex].Criteria[id]

			// Находим соответствующий entry для критерия
			var selectedEntry *MainCriterionEntry
			for i := range entries {
				if entries[i].CriterionID == crit.Id {
					selectedEntry = &entries[i]
					break
				}
			}

			if selectedEntry == nil {
				log.Printf("No entry found for criterion ID %d", crit.Id)
				return
			}

			// Создаем контейнер для отображения комментариев
			commentsContainer := container.NewVBox(
				widget.NewLabelWithStyle("Комментарии лектора:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
				widget.NewLabel(fmt.Sprintf("Для 0.0: %s", crit.Comment_000)),
				widget.NewLabel(fmt.Sprintf("Для 0.25: %s", crit.Comment_025)),
				widget.NewLabel(fmt.Sprintf("Для 0.75: %s", crit.Comment_075)),
				widget.NewLabel(fmt.Sprintf("Для 1.00: %s", crit.Comment_100)),
			)

			content := container.NewVBox(
				widget.NewLabel("Критерий: "+crit.Name),
				container.NewHBox(widget.NewLabel("Оценка:"), selectedEntry.Select),
				container.NewHBox(widget.NewLabel("Комментарий:"), selectedEntry.CommentEntry),
				commentsContainer,
			)

			contentContainer.Objects = []fyne.CanvasObject{content}
			contentContainer.Refresh()
		}
	}

	groupContainer := container.NewVBox(groupList)
	criteriaContainer := container.NewVBox(criteriaList)
	leftPanel := container.NewHSplit(groupContainer, criteriaContainer)
	leftPanel.SetOffset(0.5)

	mainContent := container.NewBorder(nil, nil, nil, nil, contentContainer)
	split := container.NewHSplit(leftPanel, mainContent)
	split.SetOffset(0.3)

	// Кнопки
	backButton := widget.NewButton("Назад", func() {
		state.currentPage = "grading_blocking"
		w.SetContent(CreateBlockingCriteriaGradingPage(state, workID, taskID, leftBackground))
	})

	finalizeButton := widget.NewButton("Завершить оценку", func() {
		// Проверка, что все критерии оценены
		evaluatedCriteria := make(map[int32]struct{})
		var missingCriteria []string
		for _, entry := range entries {
			if entry.Select.Selected == "" {
				info, exists := criteriaInfo[entry.CriterionID]
				if exists {
					missingCriteria = append(missingCriteria, fmt.Sprintf("- Группа: %s, Критерий: %s", info.groupName, info.critName))
					log.Printf("Criterion ID %d (Group: %s, Name: %s) has no mark selected", entry.CriterionID, info.groupName, info.critName)
				} else {
					log.Printf("Criterion ID %d has no mark selected (no group info)", entry.CriterionID)
				}
				continue
			}
			evaluatedCriteria[entry.CriterionID] = struct{}{}
		}

		if len(missingCriteria) > 0 {
			errorMsg := "Необходимо оценить все критерии. Пропущены следующие критерии:\n" + strings.Join(missingCriteria, "\n")
			dialog.ShowInformation("Ошибка", errorMsg, w)
			return
		}

		// Создаем новый контекст и соединение для сохранения данных
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		conn, err := grpc.Dial("localhost:50057", grpc.WithInsecure())
		if err != nil {
			log.Printf("Failed to connect to gradingservice: %v", err)
			dialog.ShowError(err, w)
			return
		}
		defer conn.Close()

		client := gradingpb.NewGradingServiceClient(conn)

		// Сохранение оценок
		for _, entry := range entries {
			mark, err := strconv.ParseFloat(entry.Select.Selected, 32)
			if err != nil {
				log.Printf("Failed to parse mark for criterion ID %d: %v", entry.CriterionID, err)
				dialog.ShowInformation("Ошибка", "Некорректное значение оценки", w)
				return
			}
			log.Printf("Saving mark for criterion ID %d: mark=%f, comment=%s", entry.CriterionID, mark, entry.CommentEntry.Text)
			resp, err := client.SetMainCriteriaMark(ctx, &gradingpb.SetMainCriteriaMarkRequest{
				WorkId:      workID,
				CriterionId: entry.CriterionID,
				Mark:        float32(mark),
				Comment:     entry.CommentEntry.Text,
			})
			if err != nil {
				log.Printf("Failed to save main criteria mark for criterion ID %d: %v", entry.CriterionID, err)
				dialog.ShowError(err, w)
				return
			}
			if resp.Error != "" {
				log.Printf("SetMainCriteriaMark error for criterion ID %d: %s", entry.CriterionID, resp.Error)
				dialog.ShowInformation("Ошибка", resp.Error, w)
				return
			}
		}
		log.Printf("All main criteria marks saved successfully")
		dialog.ShowInformation("Успех", "Оценка завершена", w)
		state.currentPage = "assistant_works"
		w.SetContent(CreateAssistantWorksPage(state, leftBackground))
	})

	// Нижняя панель с кнопками
	bottomButtons := container.NewHBox(backButton, layout.NewSpacer(), finalizeButton)
	bottomButtons = container.NewStack(canvas.NewRectangle(color.NRGBA{R: 20, G: 40, B: 80, A: 255}), bottomButtons)
	// Сборка интерфейса
	return container.NewStack(
		canvas.NewRectangle(color.NRGBA{R: 255, G: 255, B: 255, A: 255}),
		container.NewBorder(
			headerContent,
			bottomButtons,
			nil,
			nil,
			split,
		),
	)
}
