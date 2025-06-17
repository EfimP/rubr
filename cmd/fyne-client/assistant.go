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
	Status          string
	Grade           string
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

func CreateAssistantWorksPage(state *AppState) fyne.CanvasObject {
	userIDint64, err := strconv.ParseInt(state.userID, 10, 32)
	if err != nil {
		log.Printf("Некорректный ID пользователя: %v", err)
		return container.NewVBox(widget.NewLabel("Ошибка: некорректный ID пользователя"))
	}
	userID := int32(userIDint64)

	// Подключение к сервисам
	workConn, err := grpc.Dial("localhost:50054", grpc.WithInsecure())
	if err != nil {
		log.Printf("Не удалось подключиться к сервису назначений работ: %v", err)
		return container.NewVBox(widget.NewLabel("Ошибка подключения к сервису"))
	}
	defer workConn.Close()
	workClient := workassignmentpb.NewWorkAssignmentServiceClient(workConn)

	gradingConn, err := grpc.Dial("localhost:50057", grpc.WithInsecure())
	if err != nil {
		log.Printf("Не удалось подключиться к сервису оценок: %v", err)
		return container.NewVBox(widget.NewLabel("Ошибка подключения к сервису оценок"))
	}
	defer gradingConn.Close()
	gradingClient := gradingpb.NewGradingServiceClient(gradingConn)

	rubricConn, err := grpc.Dial("localhost:50055", grpc.WithInsecure())
	if err != nil {
		log.Printf("Не удалось подключиться к сервису рубрик: %v", err)
		return container.NewVBox(widget.NewLabel("Ошибка подключения к сервису рубрик"))
	}
	defer rubricConn.Close()
	rubricClient := rubricpb.NewRubricServiceClient(rubricConn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Получение списка работ
	worksResp, err := workClient.GetWorksForAssistant(ctx, &workassignmentpb.GetWorksForAssistantRequest{AssistantId: userID})
	if err != nil {
		log.Printf("Не удалось получить работы: %v", err)
		return container.NewVBox(widget.NewLabel("Ошибка загрузки работ"))
	}
	if worksResp.Error != "" {
		log.Printf("Ошибка от сервиса: %s", worksResp.Error)
		return container.NewVBox(widget.NewLabel(worksResp.Error))
	}

	// Обработка списка работ
	var data []AssistantWorkItem
	for _, work := range worksResp.Works {
		// Получение статуса через GetWorkDetails
		detailsResp, err := workClient.GetWorkDetails(ctx, &workassignmentpb.GetWorkDetailsRequest{WorkId: work.WorkId})
		if err != nil {
			log.Printf("Не удалось получить детали работы %d: %v", work.WorkId, err)
			continue
		}
		if detailsResp.Error != "" {
			log.Printf("Ошибка получения деталей работы %d: %s", work.WorkId, detailsResp.Error)
			continue
		}

		fullName := work.StudentName + " " + work.StudentSurname
		if work.StudentPatronymic != "" {
			fullName += " " + work.StudentPatronymic
		}

		item := AssistantWorkItem{
			WorkID:          work.WorkId,
			TaskID:          work.TaskId,
			TaskTitle:       work.TaskTitle,
			StudentEmail:    work.StudentEmail,
			StudentFullName: fullName,
			Status:          detailsResp.Status,
		}

		// Вычисление оценки, если статус 'graded by assistant' или 'graded by seminarist'
		if item.Status == "graded by assistant" || item.Status == "graded by seminarist" {
			// Получение оценок из student_criteria_marks
			marksResp, err := gradingClient.GetCriteriaMarks(ctx, &gradingpb.GetCriteriaMarksRequest{WorkId: work.WorkId})
			if err != nil {
				log.Printf("Не удалось получить оценки для работы %d: %v", work.WorkId, err)
				item.Grade = "Ошибка"
			} else if marksResp.Error != "" {
				log.Printf("Ошибка получения оценок для работы %d: %s", work.WorkId, marksResp.Error)
				item.Grade = "Ошибка"
			} else {
				// Получение блокирующих критериев
				blockingResp, err := rubricClient.LoadTaskBlockingCriterias(ctx, &rubricpb.LoadTaskBlockingCriteriasRequest{TaskId: work.TaskId})
				if err != nil || blockingResp.Error != "" {
					log.Printf("Не удалось загрузить блокирующие критерии для задачи %d: %v", work.TaskId, err)
					item.Grade = "Ошибка"
				} else {
					// Получение основных критериев
					mainResp, err := rubricClient.LoadTaskMainCriterias(ctx, &rubricpb.LoadTaskMainCriteriasRequest{TaskId: work.TaskId})
					if err != nil || mainResp.Error != "" {
						log.Printf("Не удалось загрузить основные критерии для задачи %d: %v", work.TaskId, err)
						item.Grade = "Ошибка"
					} else {
						// Проверка блокирующих критериев
						minBlockingMark := float32(0)
						hasBlockingMark := false
						for _, mark := range marksResp.Marks {
							for _, crit := range blockingResp.Criteria {
								if mark.CriterionId == crit.Id && mark.Mark > 0 {
									if !hasBlockingMark || mark.Mark < minBlockingMark {
										minBlockingMark = mark.Mark
										hasBlockingMark = true
									}
								}
							}
						}

						if hasBlockingMark {
							item.Grade = fmt.Sprintf("%.2f", minBlockingMark)
						} else {
							// Вычисление оценки по основным критериям
							totalMark := float32(0)
							totalMaxMark := float32(0)
							for _, mark := range marksResp.Marks {
								for _, group := range mainResp.Groups {
									for _, crit := range group.Criteria {
										if mark.CriterionId == crit.Id {
											totalMark += mark.Mark
										}
									}
								}
							}
							// Предполагаем, что каждый критерий имеет max_mark = 1
							for _, group := range mainResp.Groups {
								totalMaxMark += float32(len(group.Criteria))
							}
							if totalMaxMark > 0 {
								finalGrade := (totalMark / totalMaxMark) * 10
								item.Grade = fmt.Sprintf("%.2f", finalGrade)
							} else {
								item.Grade = "0.00"
							}
						}
					}
				}
			}
		}

		data = append(data, item)
	}

	// Настройка заголовка
	headerTextColor := color.White
	logoText := canvas.NewText("ВШЭ", headerTextColor)
	logoText.TextStyle.Bold = true
	logoText.TextSize = 24
	logoText.Alignment = fyne.TextAlignCenter
	leftHeaderObject := container.NewMax(logoText)

	headerTitle := canvas.NewText("Работы на проверку", headerTextColor)
	headerTitle.TextStyle.Bold = true
	headerTitle.TextSize = 20
	headerContent := container.New(layout.NewBorderLayout(nil, nil, leftHeaderObject, nil),
		leftHeaderObject,
		container.NewCenter(headerTitle),
	)
	headerWithBackground := container.NewMax(canvas.NewRectangle(color.NRGBA{R: 23, G: 44, B: 101, A: 255}), headerContent)

	// Вывод списка работ
	var myListWidget *widget.List
	createItem := func() fyne.CanvasObject {
		taskTitleLabel := widget.NewLabel("Задание: ")
		taskTitleLabel.TextStyle.Bold = true
		studentEmailLabel := widget.NewLabel("Email студента: ")
		studentNameLabel := widget.NewLabel("Студент: ")
		statusLabel := widget.NewLabel("Состояние: ")
		gradeLabel := widget.NewLabel("Оценка: ")
		return container.NewVBox(taskTitleLabel, studentEmailLabel, studentNameLabel, statusLabel, gradeLabel)
	}

	updateItem := func(id widget.ListItemID, item fyne.CanvasObject) {
		vbox := item.(*fyne.Container)
		taskTitleLabel := vbox.Objects[0].(*widget.Label)
		studentEmailLabel := vbox.Objects[1].(*widget.Label)
		studentNameLabel := vbox.Objects[2].(*widget.Label)
		// Исправляем индексы для statusLabel и gradeLabel
		statusLabel := vbox.Objects[3].(*widget.Label)
		gradeLabel := vbox.Objects[4].(*widget.Label)

		taskTitleLabel.SetText("Задание: " + data[id].TaskTitle)
		studentEmailLabel.SetText("Email студента: " + data[id].StudentEmail)
		studentNameLabel.SetText("Студент: " + data[id].StudentFullName)
		statusLabel.SetText("Состояние: " + data[id].Status)
		if data[id].Grade != "" {
			gradeLabel.SetText("Оценка: " + data[id].Grade)
		} else {
			gradeLabel.SetText("Оценка: -")
		}
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
		state.window.SetContent(CreateAssistantWorkDetailsPage(state, workID, taskID))
		myListWidget.UnselectAll()
	}

	// Сборка интерфейса
	listBackground := canvas.NewRectangle(color.White)
	listWithBackground := container.NewMax(listBackground, myListWidget)

	return container.NewStack(
		canvas.NewRectangle(color.White),
		container.NewBorder(
			headerWithBackground,
			nil,
			nil,
			nil,
			listWithBackground,
		),
	)
}
func CreateAssistantWorkDetailsPage(state *AppState, workID int32, taskID int32) fyne.CanvasObject {
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
		state.window.SetContent(CreateBlockingCriteriaGradingPage(state, workID, taskID))
	})

	backButton := widget.NewButton("Назад", func() {
		state.currentPage = "assistant_works"
		state.window.SetContent(CreateAssistantWorksPage(state))
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

func CreateBlockingCriteriaGradingPage(state *AppState, workID int32, taskID int32) fyne.CanvasObject {
	w := state.window

	// Подключение к gRPC сервисам
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rubricConn, err := grpc.Dial("localhost:50055", grpc.WithInsecure())
	if err != nil {
		log.Printf("Failed to connect to rubricservice: %v", err)
		return container.NewVBox(widget.NewLabel("Ошибка подключения к сервису"))
	}
	defer rubricConn.Close()
	rubricClient := rubricpb.NewRubricServiceClient(rubricConn)

	gradingConn, err := grpc.Dial("localhost:50057", grpc.WithInsecure())
	if err != nil {
		log.Printf("Failed to connect to gradingservice: %v", err)
		return container.NewVBox(widget.NewLabel("Ошибка подключения к сервису"))
	}
	defer gradingConn.Close()
	gradingClient := gradingpb.NewGradingServiceClient(gradingConn)

	// Загрузка существующих оценок
	marksResp, err := gradingClient.GetCriteriaMarks(ctx, &gradingpb.GetCriteriaMarksRequest{WorkId: workID})
	if err != nil {
		log.Printf("Не удалось загрузить оценки для работы %d: %v", workID, err)
		return container.NewVBox(widget.NewLabel("Ошибка загрузки оценок: " + err.Error()))
	}
	if marksResp.Error != "" {
		log.Printf("Ошибка загрузки оценок для работы %d: %s", workID, marksResp.Error)
		return container.NewVBox(widget.NewLabel("Ошибка загрузки оценок: " + marksResp.Error))
	}

	// Создаем карту для быстрого доступа к оценкам по criterion_id
	marksMap := make(map[int32]gradingpb.CriterionMark)
	for _, mark := range marksResp.Marks {
		marksMap[mark.CriterionId] = *mark
	}

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
		w.SetContent(CreateAssistantWorkDetailsPage(state, workID, taskID))
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
		// Редактируемое поле оценки
		evaluationEntry := widget.NewEntry()

		// Загружаем существующие данные
		if mark, exists := marksMap[crit.Id]; exists {
			commentEntry.SetText(mark.Comment)
			evaluationEntry.SetText(fmt.Sprintf("%.2f", mark.Mark))
			log.Printf("Загружена оценка для блокирующего критерия ID %d: mark=%.2f, comment=%s", crit.Id, mark.Mark, mark.Comment)
		} else {
			commentEntry.SetText(crit.Comment)
			evaluationEntry.SetText(strconv.FormatInt(crit.FinalMark, 10))
			log.Printf("Оценка не найдена для блокирующего критерия ID %d, установлена оценка по умолчанию 0.0 и комментарий лектора: %s", crit.Id, crit.Comment)
		}

		commentEntryContainer := container.NewMax(commentEntry)
		commentEntryContainer.Resize(fyne.NewSize(250, 60))
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
			log.Printf("Failed to connect to gradingservice: %v", err)
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
		w.SetContent(CreateMainCriteriaGradingPage(state, workID, taskID))
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

func CreateMainCriteriaGradingPage(state *AppState, workID int32, taskID int32) fyne.CanvasObject {
	w := state.window

	// Подключение к gRPC сервисам
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	rubricConn, err := grpc.Dial("localhost:50055", grpc.WithInsecure())
	if err != nil {
		log.Printf("Не удалось подключиться к RubricService: %v", err)
		return container.NewVBox(widget.NewLabel("Ошибка подключения к сервису"))
	}
	defer rubricConn.Close()
	rubricClient := rubricpb.NewRubricServiceClient(rubricConn)

	gradingConn, err := grpc.Dial("localhost:50057", grpc.WithInsecure())
	if err != nil {
		log.Printf("Не удалось подключиться к GradingService: %v", err)
		return container.NewVBox(widget.NewLabel("Ошибка подключения к сервису"))
	}
	defer gradingConn.Close()
	gradingClient := gradingpb.NewGradingServiceClient(gradingConn)

	// Загрузка существующих оценок
	marksResp, err := gradingClient.GetCriteriaMarks(ctx, &gradingpb.GetCriteriaMarksRequest{WorkId: workID})
	if err != nil {
		log.Printf("Не удалось загрузить оценки для работы %d: %v", workID, err)
		return container.NewVBox(widget.NewLabel("Ошибка загрузки оценок: " + err.Error()))
	}
	if marksResp.Error != "" {
		log.Printf("Ошибка загрузки оценок для работы %d: %s", workID, marksResp.Error)
		return container.NewVBox(widget.NewLabel("Ошибка загрузки оценок: " + marksResp.Error))
	}

	// Создаем карту для быстрого доступа к оценкам по criterion_id
	marksMap := make(map[int32]gradingpb.CriterionMark)
	for _, mark := range marksResp.Marks {
		marksMap[mark.CriterionId] = *mark
	}

	// Загрузка основных критериев
	resp, err := rubricClient.LoadTaskMainCriterias(ctx, &rubricpb.LoadTaskMainCriteriasRequest{TaskId: taskID})
	if err != nil || resp.Error != "" {
		errorMsg := resp.Error
		if err != nil {
			errorMsg = err.Error()
		}
		log.Printf("Не удалось загрузить основные критерии: %s", errorMsg)
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
			hasMark := false
			var currentMark string

			// Проверяем наличие оценки
			if mark, exists := marksMap[crit.Id]; exists {
				markStr := fmt.Sprintf("%.2f", mark.Mark)
				if contains(selectOptions, markStr) {
					currentMark = markStr
					hasMark = true
					log.Printf("Загружена оценка для основного критерия ID %d: mark=%.2f, comment=%s", crit.Id, mark.Mark, mark.Comment)
				}
			}

			selectWidget := widget.NewSelect(selectOptions, func(s string) {
				if !hasMark {
					// Устанавливаем комментарий лектора только если оценка не была ранее проставлена
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
				}
				// Если оценка уже была, комментарий не меняется автоматически
			})

			// Устанавливаем начальные значения
			if hasMark {
				selectWidget.SetSelected(currentMark)
				commentEntry.SetText(marksMap[crit.Id].Comment)
			}

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

	log.Printf("Инициализировано %d записей критериев", len(entries))

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
				log.Printf("Не найдена запись для критерия ID %d", crit.Id)
				return
			}

			// Создаем контейнер для отображения комментариев
			commentsContainer := container.NewVBox(
				widget.NewLabelWithStyle("Комментарии лектора:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
				widget.NewLabel(fmt.Sprintf("Для 0.0: %s", crit.Comment_000)),
				widget.NewLabel(fmt.Sprintf("Для 0.25: %s", crit.Comment_025)),
				widget.NewLabel(fmt.Sprintf("Для 0.50: %s", crit.Comment_050)),
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
		w.SetContent(CreateBlockingCriteriaGradingPage(state, workID, taskID))
	})

	finalizeButton := widget.NewButton("Завершить оценку", func() {
		// Создаем новый контекст и соединение
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		gradingConn, err := grpc.Dial("localhost:50057", grpc.WithInsecure())
		if err != nil {
			log.Printf("Не удалось подключиться к GradingService: %v", err)
			dialog.ShowError(err, w)
			return
		}
		defer gradingConn.Close()
		gradingClient := gradingpb.NewGradingServiceClient(gradingConn)

		// Проверка, что все критерии оценены
		evaluatedCriteria := make(map[int32]struct{})
		var missingCriteria []string
		for _, entry := range entries {
			if entry.Select.Selected == "" {
				info, exists := criteriaInfo[entry.CriterionID]
				if exists {
					missingCriteria = append(missingCriteria, fmt.Sprintf("- Группа: %s, Критерий: %s", info.groupName, info.critName))
					log.Printf("Критерий ID %d (Группа: %s, Имя: %s) не имеет выбранной оценки", entry.CriterionID, info.groupName, info.critName)
				} else {
					log.Printf("Критерий ID %d не имеет выбранной оценки (информация о группе отсутствует)", entry.CriterionID)
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

		// Сохранение оценок
		for _, entry := range entries {
			mark, err := strconv.ParseFloat(entry.Select.Selected, 32)
			if err != nil {
				log.Printf("Не удалось разобрать оценку для критерия ID %d: %v", entry.CriterionID, err)
				dialog.ShowInformation("Ошибка", "Некорректное значение оценки", w)
				return
			}
			start := time.Now()
			log.Printf("Сохранение оценки для критерия ID %d: mark=%f, comment=%s", entry.CriterionID, mark, entry.CommentEntry.Text)
			resp, err := gradingClient.SetMainCriteriaMark(ctx, &gradingpb.SetMainCriteriaMarkRequest{
				WorkId:      workID,
				CriterionId: entry.CriterionID,
				Mark:        float32(mark),
				Comment:     entry.CommentEntry.Text,
			})
			if err != nil {
				log.Printf("Не удалось сохранить оценку основного критерия для ID %d: %v (время выполнения: %v)", entry.CriterionID, err, time.Since(start))
				dialog.ShowError(err, w)
				return
			}
			if resp.Error != "" {
				log.Printf("Ошибка SetMainCriteriaMark для критерия ID %d: %s (время выполнения: %v)", entry.CriterionID, resp.Error, time.Since(start))
				dialog.ShowInformation("Ошибка", resp.Error, w)
				return
			}
			log.Printf("Оценка для критерия ID %d сохранена успешно (время выполнения: %v)", entry.CriterionID, time.Since(start))
		}

		// Обновление статуса работы
		start := time.Now()
		updateResp, err := gradingClient.UpdateWorkStatus(ctx, &gradingpb.UpdateWorkStatusRequest{
			WorkId: workID,
			Status: "graded by assistant",
		})
		if err != nil {
			log.Printf("Не удалось обновить статус работы %d: %v (время выполнения: %v)", workID, err, time.Since(start))
			dialog.ShowError(err, w)
			return
		}
		if updateResp.Error != "" {
			log.Printf("Ошибка UpdateWorkStatus для работы %d: %s (время выполнения: %v)", workID, updateResp.Error, time.Since(start))
			dialog.ShowInformation("Ошибка", updateResp.Error, w)
			return
		}

		log.Printf("Все оценки для работы %d сохранены, статус обновлен на 'graded by assistant' (время выполнения: %v)", workID, time.Since(start))
		dialog.ShowInformation("Успех", "Оценка завершена", w)
		state.currentPage = "assistant_works"
		w.SetContent(CreateAssistantWorksPage(state))
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

// Вспомогательная функция для проверки наличия строки в срезе
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
