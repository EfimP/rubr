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
	"google.golang.org/grpc"
	"image/color"
	"log"
	rubricpb "rubr/proto/rubric"
	workpb "rubr/proto/work"
	"strconv"
	"time"
)

type MyListItem struct {
	ID      int32
	Name    string
	DueDate string
}

type CriterionEntry struct {
	ID               int32
	NameEntry        *widget.Entry
	DescriptionEntry *widget.Entry
	CommentEntry     *widget.Entry
	EvaluationEntry  *widget.Entry
	Container        *fyne.Container
}

func showDateTimePickerDialog(parent fyne.Window, selectedTime *time.Time, isSelected *bool, displayEntry *widget.Entry) {
	var initialDate time.Time
	var currentHour, currentMinute string
	if *isSelected {
		initialDate = *selectedTime
	} else {
		initialDate = time.Now()
	}
	currentHour = fmt.Sprintf("%02d", initialDate.Hour())
	currentMinute = fmt.Sprintf("%02d", initialDate.Minute())

	dateFromCalendar := initialDate
	calendar := widget.NewCalendar(initialDate, func(t time.Time) {
		dateFromCalendar = t
	})

	hours := make([]string, 24)
	for i := 0; i < 24; i++ {
		hours[i] = fmt.Sprintf("%02d", i)
	}
	minutes := make([]string, 60)
	for i := 0; i < 60; i++ {
		minutes[i] = fmt.Sprintf("%02d", i)
	}

	hourSelect := widget.NewSelect(hours, func(s string) { currentHour = s })
	hourSelect.SetSelected(currentHour)
	minuteSelect := widget.NewSelect(minutes, func(s string) { currentMinute = s })
	minuteSelect.SetSelected(currentMinute)

	nowButton := widget.NewButton("Now", func() {
		currentTime := time.Now()
		currentHour = fmt.Sprintf("%02d", currentTime.Hour())
		currentMinute = fmt.Sprintf("%02d", currentTime.Minute())
		hourSelect.SetSelected(currentHour)
		minuteSelect.SetSelected(currentMinute)
	})

	timeLayout := container.New(layout.NewHBoxLayout(),
		widget.NewLabel("Time"),
		hourSelect,
		widget.NewLabel(":"),
		minuteSelect,
		layout.NewSpacer(),
		nowButton,
	)
	dialogContent := container.NewVBox(
		widget.NewLabelWithStyle("Choose date and time", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		calendar,
		widget.NewSeparator(),
		timeLayout,
	)

	d := dialog.NewCustomConfirm(
		"",
		"Ok",
		"Cancel",
		dialogContent,
		func(ok bool) {
			if ok {
				h, _ := strconv.Atoi(currentHour)
				m, _ := strconv.Atoi(currentMinute)
				finalTime := time.Date(
					dateFromCalendar.Year(),
					dateFromCalendar.Month(),
					dateFromCalendar.Day(),
					h,
					m,
					0, 0,
					dateFromCalendar.Location(),
				)
				*selectedTime = finalTime
				*isSelected = true
				displayEntry.SetText(finalTime.Format("02.01.2006 15:04"))
			}
		},
		parent,
	)
	d.Resize(fyne.NewSize(400, 500))
	d.Show()
}

func CreateLectorWorksPage(state *AppState) fyne.CanvasObject {
	userIDint64, err := strconv.ParseInt(state.userID, 10, 32)
	if err != nil {
		log.Printf("Invalid user ID: %v", err)
		return container.NewVBox(widget.NewLabel("Ошибка: некорректный ID пользователя"))
	}
	userID := int32(userIDint64)

	conn, err := grpc.Dial("localhost:50053", grpc.WithInsecure())
	if err != nil {
		log.Printf("Failed to connect to workservice: %v", err)
		return container.NewVBox(widget.NewLabel("Ошибка подключения к сервису работ"))
	}
	defer conn.Close()

	workClient := workpb.NewWorkServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := workClient.GetTasksForLector(ctx, &workpb.GetTasksForLectorRequest{LectorId: userID})
	if err != nil {
		log.Printf("Failed to get tasks: %v", err)
		return container.NewVBox(widget.NewLabel("Ошибка загрузки работ"))
	}
	if resp.Error != "" {
		log.Println("GetTasksForLector error:", resp.Error)
		return container.NewVBox(widget.NewLabel(resp.Error))
	}

	var data []MyListItem
	for _, task := range resp.Tasks {
		// Парсим дедлайн из RFC3339 и форматируем в читаемый вид
		var dueDate string
		if task.Deadline != "" {
			deadlineTime, err := time.Parse(time.RFC3339, task.Deadline)
			if err != nil {
				log.Printf("Failed to parse deadline for task %d: %v", task.Id, err)
				dueDate = "Неверный формат даты"
			} else {
				dueDate = deadlineTime.Format("02.01.2006 15:04")
			}
		} else {
			dueDate = "Не указан"
		}
		data = append(data, MyListItem{
			ID:      task.Id,
			Name:    task.Title,
			DueDate: dueDate,
		})
	}

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

	var myListWidget *widget.List
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

	updateItem := func(id widget.ListItemID, item fyne.CanvasObject) {
		cellContainer := item.(*fyne.Container)
		textGroup := cellContainer.Objects[0].(*fyne.Container)
		changeButton := cellContainer.Objects[2].(*widget.Button)
		deleteButton := cellContainer.Objects[3].(*widget.Button)

		dueDateLabel := textGroup.Objects[0].(*widget.Label)
		nameLabel := textGroup.Objects[1].(*widget.Label)

		dueDateLabel.SetText("Дата сдачи: " + data[id].DueDate)
		nameLabel.SetText(data[id].Name)

		currentID := id
		changeButton.OnTapped = func() {
			taskID := data[currentID].ID
			CreateWorkPage(state, &taskID)
		}

		deleteButton.OnTapped = func() {
			dialog.ShowConfirm(
				"Подтверждение удаления",
				fmt.Sprintf("Вы уверены, что хотите удалить работу '%s'?", data[currentID].Name),
				func(confirmed bool) {
					if confirmed {
						conn, err := grpc.Dial("localhost:50053", grpc.WithInsecure())
						if err != nil {
							log.Printf("Failed to connect to workservice: %v", err)
							return
						}
						defer conn.Close()

						workClient := workpb.NewWorkServiceClient(conn)
						ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
						defer cancel()
						_, err = workClient.DeleteTask(ctx, &workpb.DeleteTaskRequest{TaskId: data[currentID].ID})
						if err != nil {
							log.Printf("Failed to delete task: %v", err)
							return
						}
						data = append(data[:currentID], data[currentID+1:]...)
						myListWidget.Refresh()
					}
				},
				state.window,
			)
		}
	}

	myListWidget = widget.NewList(
		func() int { return len(data) },
		createItem,
		updateItem,
	)

	addButton := widget.NewButton("Добавить", func() {
		CreateWorkPage(state, nil)
	})
	addButtonContainer := container.New(layout.NewHBoxLayout(), layout.NewSpacer(), addButton)

	listBackground := canvas.NewRectangle(color.White)
	listWithBackground := container.NewMax(listBackground, myListWidget)

	return container.NewMax(
		canvas.NewRectangle(color.NRGBA{R: 20, G: 40, B: 80, A: 255}),
		container.NewBorder(
			headerContent,
			addButtonContainer,
			nil,
			nil,
			listWithBackground,
		),
	)
}
func CreateWorkPage(state *AppState, taskID *int32) {
	w := state.window
	var isNewWork bool
	if taskID == nil {
		isNewWork = true
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	userIDint64, err := strconv.ParseInt(state.userID, 10, 32)
	if err != nil {
		log.Printf("Invalid user ID: %v", err)
		dialog.ShowInformation("Ошибка", "Некорректный ID пользователя", w)
		return
	}
	userID := int32(userIDint64)

	workConn, err := grpc.Dial("localhost:50053", grpc.WithInsecure())
	if err != nil {
		log.Printf("Failed to connect to workservice: %v", err)
		dialog.ShowError(err, w)
		return
	}
	defer workConn.Close()

	workClient := workpb.NewWorkServiceClient(workConn)

	groupsResp, err := workClient.GetGroups(ctx, &workpb.GetGroupsRequest{LectorId: userID})
	if err != nil {
		log.Printf("Failed to get groups: %v", err)
		dialog.ShowError(err, w)
		return
	}
	if len(groupsResp.Groups) == 0 {
		dialog.ShowInformation("Ошибка", "Нет доступных групп для этого лектора", w)
		return
	}

	disciplinesResp, err := workClient.GetDisciplines(ctx, &workpb.GetDisciplinesRequest{LectorId: userID})
	if err != nil {
		log.Printf("Failed to get disciplines: %v", err)
		dialog.ShowError(err, w)
		return
	}
	if len(disciplinesResp.Disciplines) == 0 {
		dialog.ShowInformation("Ошибка", "Нет доступных дисциплин для этого лектора", w)
		return
	}

	var existingTask *workpb.GetTaskDetailsResponse
	if !isNewWork {
		resp, err := workClient.GetTaskDetails(ctx, &workpb.GetTaskDetailsRequest{TaskId: *taskID})
		if err != nil {
			log.Printf("Failed to get task details: %v", err)
			dialog.ShowError(err, w)
			return
		}
		if resp.Error != "" {
			dialog.ShowInformation("Ошибка", resp.Error, w)
			return
		}
		existingTask = resp
	}

	groupOptions := make([]string, len(groupsResp.Groups))
	groupIDs := make(map[string]int32)
	for i, group := range groupsResp.Groups {
		groupOptions[i] = group.Name
		groupIDs[group.Name] = group.Id
	}

	disciplineOptions := make([]string, len(disciplinesResp.Disciplines))
	disciplineIDs := make(map[string]int32)
	for i, discipline := range disciplinesResp.Disciplines {
		disciplineOptions[i] = discipline.Name
		disciplineIDs[discipline.Name] = discipline.Id
	}

	headerTextColor := color.White
	logoText := canvas.NewText("ВШЭ", headerTextColor)
	logoText.TextStyle.Bold = true
	logoText.TextSize = 24
	logoText.Alignment = fyne.TextAlignCenter
	leftHeaderObject := container.NewStack(logoText)

	headerTitleText := "Создать работу"
	if !isNewWork {
		headerTitleText = "Изменить работу"
	}
	headerTitle := canvas.NewText(headerTitleText, headerTextColor)
	headerTitle.TextStyle.Bold = true
	headerTitle.TextSize = 20
	headerTitle.Alignment = fyne.TextAlignCenter

	headerContent := container.New(layout.NewBorderLayout(nil, nil, leftHeaderObject, nil),
		leftHeaderObject,
		container.NewCenter(headerTitle),
	)

	// Add Back button
	backButton := widget.NewButton("Назад", func() {
		state.currentPage = "lector_works"
		w.SetContent(CreateLectorWorksPage(state))
	})

	titleEntry := widget.NewEntry()
	if existingTask != nil {
		titleEntry.SetText(existingTask.Title)
	} else {
		titleEntry.SetPlaceHolder("Название")
	}

	descriptionEntry := widget.NewMultiLineEntry()
	if existingTask != nil {
		descriptionEntry.SetText(existingTask.Description)
	} else {
		descriptionEntry.SetPlaceHolder("Описание работы")
	}
	scrollableDescription := container.NewVScroll(descriptionEntry)
	scrollableDescription.SetMinSize(fyne.NewSize(0, descriptionEntry.MinSize().Height*5))

	groupSelect := widget.NewSelect(groupOptions, func(string) {})
	if len(groupOptions) > 0 {
		if existingTask != nil {
			groupSelect.SetSelected(existingTask.GroupName)
		} else {
			groupSelect.SetSelected(groupOptions[0])
		}
	}

	disciplineSelect := widget.NewSelect(disciplineOptions, func(string) {})
	if len(disciplineOptions) > 0 {
		if existingTask != nil {
			disciplineSelect.SetSelected(existingTask.DisciplineName)
		} else {
			disciplineSelect.SetSelected(disciplineOptions[0])
		}
	}

	dateAndTimeEntry := widget.NewEntry()
	dateAndTimeEntry.SetPlaceHolder("Выберите дату и время дедлайна")
	dateAndTimeEntry.Disable()

	var selectedDateTime time.Time
	var isDateTimeSelected bool
	if existingTask != nil {
		deadlineTime, err := time.Parse(time.RFC3339, existingTask.Deadline)
		if err != nil {
			log.Printf("Failed to parse deadline: %v", err)
			dialog.ShowError(err, w)
			return
		}
		selectedDateTime = deadlineTime
		isDateTimeSelected = true
		dateAndTimeEntry.SetText(deadlineTime.Format("02.01.2006 15:04"))
	}

	selectDateTimeButton := widget.NewButton("Выбрать", func() {
		showDateTimePickerDialog(w, &selectedDateTime, &isDateTimeSelected, dateAndTimeEntry)
	})

	dateTimeInputContainer := container.New(layout.NewHBoxLayout(),
		dateAndTimeEntry,
		selectDateTimeButton,
	)

	inputGrid := container.New(layout.NewVBoxLayout(),
		container.New(layout.NewHBoxLayout(),
			container.New(layout.NewGridLayoutWithColumns(1), titleEntry),
			layout.NewSpacer(),
			dateTimeInputContainer,
		),
		container.NewPadded(widget.NewLabel("Группа")),
		groupSelect,
		container.NewPadded(widget.NewLabel("Дисциплина")),
		disciplineSelect,
		scrollableDescription,
	)

	nextButton := widget.NewButton("Далее", func() {
		if !isDateTimeSelected {
			dialog.ShowInformation("Ошибка", "Пожалуйста, выберите дедлайн", w)
			return
		}
		if groupSelect.Selected == "" || disciplineSelect.Selected == "" {
			dialog.ShowInformation("Ошибка", "Пожалуйста, выберите группу и дисциплину", w)
			return
		}
		deadline := selectedDateTime.Format(time.RFC3339)

		workConn, err := grpc.Dial("localhost:50053", grpc.WithInsecure())
		if err != nil {
			log.Printf("Failed to connect to workservice: %v", err)
			dialog.ShowError(err, w)
			return
		}
		defer workConn.Close()
		workClient := workpb.NewWorkServiceClient(workConn)

		if isNewWork {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			resp, err := workClient.CreateWork(ctx, &workpb.CreateWorkRequest{
				LectorId:     userID,
				GroupId:      groupIDs[groupSelect.Selected],
				Title:        titleEntry.Text,
				Description:  descriptionEntry.Text,
				Deadline:     deadline,
				DisciplineId: disciplineIDs[disciplineSelect.Selected],
				ContentUrl:   "",
			})
			if err != nil {
				log.Printf("Failed to create work: %v", err)
				dialog.ShowError(err, w)
				return
			}
			if resp.Error != "" {
				dialog.ShowInformation("Ошибка", resp.Error, w)
				return
			}
			taskID := resp.TaskId
			CreateBlockingCriteriaPage(state, taskID)
		} else {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			respTitle, err := workClient.SetTaskTitle(ctx, &workpb.SetTaskTitleRequest{
				TaskId: *taskID,
				Title:  titleEntry.Text,
			})
			if err != nil {
				log.Printf("Failed to set task title: %v", err)
				dialog.ShowError(err, w)
				return
			}
			if respTitle.Error != "" {
				dialog.ShowInformation("Ошибка", respTitle.Error, w)
				return
			}

			respDesc, err := workClient.SetTaskDescription(ctx, &workpb.SetTaskDescriptionRequest{
				TaskId:      *taskID,
				Description: descriptionEntry.Text,
			})
			if err != nil {
				log.Printf("Failed to set task description: %v", err)
				dialog.ShowError(err, w)
				return
			}
			if respDesc.Error != "" {
				dialog.ShowInformation("Ошибка", respDesc.Error, w)
				return
			}

			respDeadline, err := workClient.SetTaskDeadline(ctx, &workpb.SetTaskDeadlineRequest{
				TaskId:   *taskID,
				Deadline: deadline,
			})
			if err != nil {
				log.Printf("Failed to set task deadline: %v", err)
				dialog.ShowError(err, w)
				return
			}
			if respDeadline.Error != "" {
				dialog.ShowInformation("Ошибка", respDeadline.Error, w)
				return
			}

			respGroupDisc, err := workClient.UpdateTaskGroupAndDiscipline(ctx, &workpb.UpdateTaskGroupAndDisciplineRequest{
				TaskId:       *taskID,
				GroupId:      groupIDs[groupSelect.Selected],
				DisciplineId: disciplineIDs[disciplineSelect.Selected],
			})
			if err != nil {
				log.Printf("Failed to update task group and discipline: %v", err)
				dialog.ShowError(err, w)
				return
			}
			if respGroupDisc.Error != "" {
				dialog.ShowInformation("Ошибка", respGroupDisc.Error, w)
				return
			}

			CreateBlockingCriteriaPage(state, *taskID)
		}
	})
	buttonsContainer := container.New(layout.NewHBoxLayout(), backButton, layout.NewSpacer(), nextButton)

	contentBackground := canvas.NewRectangle(color.White)
	contentWithPadding := container.NewPadded(inputGrid)
	centralContent := container.NewStack(contentBackground, contentWithPadding)

	w.SetContent(container.NewStack(
		canvas.NewRectangle(color.NRGBA{R: 20, G: 40, B: 80, A: 255}),
		container.NewBorder(
			headerContent,
			buttonsContainer,
			nil,
			nil,
			centralContent,
		),
	))
}

func CreateBlockingCriteriaPage(state *AppState, taskID int32) {
	w := state.window

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rubricConn, err := grpc.Dial("localhost:50055", grpc.WithInsecure())
	if err != nil {
		log.Printf("Failed to connect to rubricservice: %v", err)
		dialog.ShowError(err, w)
		return
	}
	defer rubricConn.Close()

	rubricClient := rubricpb.NewRubricServiceClient(rubricConn)

	resp, err := rubricClient.LoadTaskBlockingCriterias(ctx, &rubricpb.LoadTaskBlockingCriteriasRequest{TaskId: taskID})
	if err != nil {
		log.Printf("Failed to load blocking criteria: %v", err)
		dialog.ShowError(err, w)
		return
	}
	if resp.Error != "" {
		dialog.ShowInformation("Ошибка", resp.Error, w)
		return
	}

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
		CreateWorkPage(state, &taskID)
	})
	backButtonContainer := container.NewHBox(layout.NewSpacer(), backButton)

	criteriaListContainer := container.NewVBox()
	columnHeaders := container.New(layout.NewGridLayoutWithColumns(4),
		container.NewPadded(widget.NewLabelWithStyle("Название критерия", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
		container.NewPadded(widget.NewLabelWithStyle("Описание критерия", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
		container.NewPadded(widget.NewLabelWithStyle("Комментарий", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
		container.NewPadded(widget.NewLabelWithStyle("Финальная оценка", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
	)

	var activeCriteria []*CriterionEntry
	addCriterionEntry := func(crit *rubricpb.BlockingCriteria) {
		nameEntry := widget.NewEntry()
		var id int32
		if crit != nil {
			nameEntry.SetText(crit.Name)
			id = crit.Id
		} else {
			nameEntry.SetPlaceHolder("Название")
			id = 0
		}
		nameEntryContainer := container.NewMax(nameEntry)
		nameEntryContainer.Resize(fyne.NewSize(250, 60))

		descriptionEntry := widget.NewEntry()
		if crit != nil {
			descriptionEntry.SetText(crit.Description)
		} else {
			descriptionEntry.SetPlaceHolder("Описание")
		}
		descriptionEntryContainer := container.NewMax(descriptionEntry)
		descriptionEntryContainer.Resize(fyne.NewSize(250, 60))

		commentEntry := widget.NewEntry()
		if crit != nil {
			commentEntry.SetText(crit.Comment)
		} else {
			commentEntry.SetPlaceHolder("Комментарий")
		}
		commentEntryContainer := container.NewMax(commentEntry)
		commentEntryContainer.Resize(fyne.NewSize(250, 60))

		evaluationEntry := widget.NewEntry()
		if crit != nil {
			evaluationEntry.SetText(strconv.FormatInt(crit.FinalMark, 10))
		} else {
			evaluationEntry.SetPlaceHolder("Оценка")
		}
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

		activeCriteria = append(activeCriteria, &CriterionEntry{
			ID:               id,
			NameEntry:        nameEntry,
			DescriptionEntry: descriptionEntry,
			CommentEntry:     commentEntry,
			EvaluationEntry:  evaluationEntry,
			Container:        criterionRow,
		})
	}

	deleteCriterionEntry := func() {
		if len(activeCriteria) == 0 {
			dialog.ShowInformation("Ошибка", "Нет критериев для удаления", w)
			return
		}

		criteriaNames := make([]string, len(activeCriteria))
		nameToIndex := make(map[string]int)
		for i, crit := range activeCriteria {
			name := crit.NameEntry.Text
			if name == "" {
				name = "Критерий без названия"
			}
			criteriaNames[i] = name
			nameToIndex[name] = i
		}

		// Создаем выпадающий список для выбора критерия
		selectEntry := widget.NewSelect(criteriaNames, func(selected string) {})
		selectEntry.PlaceHolder = "Выберите критерий для удаления"
		selectEntryContainer := container.NewVBox(
			widget.NewLabel("Выберите критерий:"),
			selectEntry,
		)

		deleteDialog := dialog.NewCustomConfirm(
			"Удаление критерия",
			"Удалить",
			"Отмена",
			selectEntryContainer,
			func(confirmed bool) {
				if !confirmed || selectEntry.Selected == "" {
					return
				}

				// Находим выбранный критерий
				index, ok := nameToIndex[selectEntry.Selected]
				if !ok {
					dialog.ShowInformation("Ошибка", "Выбранный критерий не найден", w)
					return
				}
				selectedCriterion := activeCriteria[index]

				// Если критерий существует в базе (ID != 0), удаляем его
				if selectedCriterion.ID != 0 {
					rubricConn, err := grpc.Dial("localhost:50055", grpc.WithInsecure())
					if err != nil {
						log.Printf("Failed to connect to rubricservice: %v", err)
						dialog.ShowError(err, w)
						return
					}
					defer rubricConn.Close()
					rubricClient := rubricpb.NewRubricServiceClient(rubricConn)

					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()

					resp, err := rubricClient.DeleteBlockingCriteria(ctx, &rubricpb.DeleteBlockingCriteriaRequest{CriteriaId: selectedCriterion.ID})
					if err != nil {
						log.Printf("Failed to delete blocking criteria: %v", err)
						dialog.ShowError(err, w)
						return
					}
					if resp.Error != "" {
						log.Println("DeleteBlockingCriteria error:", resp.Error)
						dialog.ShowInformation("Ошибка", resp.Error, w)
						return
					}
				}

				// Удаляем критерий из интерфейса
				criteriaListContainer.Remove(selectedCriterion.Container)
				activeCriteria = append(activeCriteria[:index], activeCriteria[index+1:]...)
				criteriaListContainer.Refresh()
			},
			w,
		)
		deleteDialog.Show()
	}

	for _, crit := range resp.Criteria {
		addCriterionEntry(crit)
	}
	if len(resp.Criteria) == 0 {
		addCriterionEntry(nil)
	}

	listLabel := widget.NewLabelWithStyle("Список блокирующих критериев", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	scrollableCriteria := container.NewVScroll(criteriaListContainer)
	scrollableCriteria.SetMinSize(fyne.NewSize(0, 400))

	addButton := widget.NewButton("Добавить", func() { addCriterionEntry(nil) })
	deleteButton := widget.NewButton("Удалить", func() { deleteCriterionEntry() })
	nextButton := widget.NewButton("Далее", func() {
		rubricConn, err := grpc.Dial("localhost:50055", grpc.WithInsecure())
		if err != nil {
			log.Printf("Failed to connect to rubricservice: %v", err)
			dialog.ShowError(err, w)
			return
		}
		defer rubricConn.Close()
		rubricClient := rubricpb.NewRubricServiceClient(rubricConn)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Удаляем все существующие блокирующие критерии для task_id
		respDelete, err := rubricClient.DeleteTaskBlockingCriterias(ctx, &rubricpb.DeleteTaskBlockingCriteriasRequest{TaskId: taskID})
		if err != nil {
			log.Printf("Failed to delete existing blocking criteria: %v", err)
			dialog.ShowError(err, w)
			return
		}
		if respDelete.Error != "" {
			log.Println("DeleteTaskBlockingCriterias error:", respDelete.Error)
			dialog.ShowInformation("Ошибка", respDelete.Error, w)
			return
		}

		// Создаем новые критерии
		for _, criterion := range activeCriteria {
			if criterion.NameEntry.Text == "" || criterion.EvaluationEntry.Text == "" {
				dialog.ShowInformation("Ошибка", "Название и оценка обязательны для всех критериев", w)
				return
			}
			finalMark, err := strconv.ParseInt(criterion.EvaluationEntry.Text, 10, 64)
			if err != nil {
				dialog.ShowInformation("Ошибка", "Оценка должна быть числом", w)
				return
			}
			resp, err := rubricClient.CreateNewBlockingCriteria(ctx, &rubricpb.CreateNewBlockingCriteriaRequest{
				TaskId:      taskID,
				Name:        criterion.NameEntry.Text,
				Description: criterion.DescriptionEntry.Text,
				Comment:     criterion.CommentEntry.Text,
				FinalMark:   finalMark,
			})
			if err != nil {
				log.Printf("Failed to create blocking criteria: %v", err)
				dialog.ShowError(err, w)
				return
			}
			if resp.Error != "" {
				log.Println("CreateNewBlockingCriteria error:", resp.Error)
				dialog.ShowInformation("Ошибка", resp.Error, w)
				return
			}
		}
		CreateMainCriteriaPage(state, taskID)
	})

	bottomButtons := container.New(layout.NewHBoxLayout(),
		addButton,
		deleteButton,
		layout.NewSpacer(),
		nextButton,
	)
	bottomButtonsWithPadding := container.NewPadded(bottomButtons)

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

	w.SetContent(container.NewStack(
		canvas.NewRectangle(color.NRGBA{R: 20, G: 40, B: 80, A: 255}),
		container.NewBorder(
			container.NewVBox(headerContent, backButtonContainer),
			bottomButtonsWithPadding,
			nil,
			nil,
			centralContent,
		),
	))
}

func CreateMainCriteriaPage(state *AppState, taskID int32) {
	w := state.window

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rubricConn, err := grpc.Dial("localhost:50055", grpc.WithInsecure())
	if err != nil {
		log.Printf("Failed to connect to rubricservice: %v", err)
		dialog.ShowError(err, w)
		return
	}
	defer rubricConn.Close()

	rubricClient := rubricpb.NewRubricServiceClient(rubricConn)

	resp, err := rubricClient.LoadTaskMainCriterias(ctx, &rubricpb.LoadTaskMainCriteriasRequest{TaskId: taskID})
	if err != nil {
		log.Printf("Failed to load main criteria: %v", err)
		dialog.ShowError(err, w)
		return
	}
	if resp.Error != "" {
		dialog.ShowInformation("Ошибка", resp.Error, w)
		return
	}

	type Group struct {
		Name        string
		GroupID     int32
		Criteria    []string
		CriteriaIDs []int32
	}
	groups := make([]Group, len(resp.Groups))
	for i, g := range resp.Groups {
		groups[i] = Group{
			Name:        g.GroupName,
			GroupID:     g.Id,
			Criteria:    make([]string, len(g.Criteria)),
			CriteriaIDs: make([]int32, len(g.Criteria)),
		}
		for j, crit := range g.Criteria {
			groups[i].Criteria[j] = crit.Name
			groups[i].CriteriaIDs[j] = crit.Id
		}
	}

	selectedGroupIndex := -1

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

	backButton := widget.NewButton("Назад", func() {
		state.currentPage = "lector_blocking"
		CreateBlockingCriteriaPage(state, taskID)
	})
	backButtonContainer := container.NewHBox(layout.NewSpacer(), backButton)

	contentContainer := container.New(layout.NewMaxLayout(), widget.NewLabel("Выберите группу и критерий"))

	createButton := widget.NewButton("Создать", func() {
		state.currentPage = "lector_works"
		w.SetContent(createContent(state))
	})
	createButtonColored := container.NewStack(createButton, canvas.NewRectangle(color.NRGBA{R: 20, G: 40, B: 80, A: 192}))

	mainContent := container.NewBorder(nil, createButtonColored, nil, nil, contentContainer)

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

	addGroupButton := widget.NewButton("Добавить группу", func() {
		entry := widget.NewEntry()
		dialog.ShowForm("Новая группа", "Создать", "Отмена", []*widget.FormItem{
			{Text: "Название группы", Widget: entry},
		}, func(b bool) {
			if b && entry.Text != "" {
				rubricConn, err := grpc.Dial("localhost:50055", grpc.WithInsecure())
				if err != nil {
					log.Printf("Failed to connect to rubricservice: %v", err)
					dialog.ShowError(err, w)
					return
				}
				defer rubricConn.Close()
				rubricClient := rubricpb.NewRubricServiceClient(rubricConn)

				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				resp, err := rubricClient.CreateCriteriaGroup(ctx, &rubricpb.CreateCriteriaGroupRequest{
					TaskId:    taskID,
					GroupName: entry.Text,
				})
				if err != nil {
					log.Printf("Failed to create criteria group: %v", err)
					dialog.ShowError(err, w)
					return
				}
				if resp.Error != "" {
					dialog.ShowInformation("Ошибка", resp.Error, w)
					return
				}
				groups = append(groups, Group{Name: entry.Text, GroupID: resp.GroupId, Criteria: []string{}, CriteriaIDs: []int32{}})
				groupList.Refresh()
			}
		}, w)
	})

	deleteGroupButton := widget.NewButton("Удалить группу", func() {
		if len(groups) == 0 {
			dialog.ShowInformation("Ошибка", "Нет групп для удаления", w)
			return
		}

		groupNames := make([]string, len(groups))
		nameToIndex := make(map[string]int)
		for i, group := range groups {
			groupNames[i] = group.Name
			nameToIndex[group.Name] = i
		}

		selectEntry := widget.NewSelect(groupNames, func(selected string) {})
		selectEntry.PlaceHolder = "Выберите группу для удаления"
		selectEntryContainer := container.NewVBox(
			widget.NewLabel("Выберите группу:"),
			selectEntry,
		)

		deleteDialog := dialog.NewCustomConfirm(
			"Удаление группы",
			"Удалить",
			"Отмена",
			selectEntryContainer,
			func(confirmed bool) {
				if !confirmed || selectEntry.Selected == "" {
					return
				}

				index, ok := nameToIndex[selectEntry.Selected]
				if !ok {
					dialog.ShowInformation("Ошибка", "Выбранная группа не найдена", w)
					return
				}

				rubricConn, err := grpc.Dial("localhost:50055", grpc.WithInsecure())
				if err != nil {
					log.Printf("Failed to connect to rubricservice: %v", err)
					dialog.ShowError(err, w)
					return
				}
				defer rubricConn.Close()
				rubricClient := rubricpb.NewRubricServiceClient(rubricConn)

				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				resp, err := rubricClient.DeleteCriteriaGroup(ctx, &rubricpb.DeleteCriteriaGroupRequest{GroupId: groups[index].GroupID})
				if err != nil {
					log.Printf("Failed to delete criteria group: %v", err)
					dialog.ShowError(err, w)
					return
				}
				if resp.Error != "" {
					log.Println("DeleteCriteriaGroup error:", resp.Error)
					dialog.ShowInformation("Ошибка", resp.Error, w)
					return
				}

				groups = append(groups[:index], groups[index+1:]...)
				if selectedGroupIndex == index {
					selectedGroupIndex = -1
					contentContainer.Objects = []fyne.CanvasObject{widget.NewLabel("Выберите группу и критерий")}
					contentContainer.Refresh()
				} else if selectedGroupIndex > index {
					selectedGroupIndex--
				}
				groupList.Refresh()
			},
			w,
		)
		deleteDialog.Show()
	})

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

	addCriterionButton := widget.NewButton("Добавить критерий", func() {
		if selectedGroupIndex < 0 {
			dialog.ShowInformation("Ошибка", "Сначала выберите группу", w)
			return
		}
		entry := widget.NewEntry()
		dialog.ShowForm("Новый критерий", "Создать", "Отмена", []*widget.FormItem{
			{Text: "Название критерия", Widget: entry},
		}, func(b bool) {
			if b && entry.Text != "" {
				rubricConn, err := grpc.Dial("localhost:50055", grpc.WithInsecure())
				if err != nil {
					log.Printf("Failed to connect to rubricservice: %v", err)
					dialog.ShowError(err, w)
					return
				}
				defer rubricConn.Close()
				rubricClient := rubricpb.NewRubricServiceClient(rubricConn)

				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				resp, err := rubricClient.CreateCriterion(ctx, &rubricpb.CreateCriterionRequest{
					GroupId: groups[selectedGroupIndex].GroupID,
					Name:    entry.Text,
				})
				if err != nil {
					log.Printf("Failed to create criterion: %v", err)
					dialog.ShowError(err, w)
					return
				}
				if resp.Error != "" {
					dialog.ShowInformation("Ошибка", resp.Error, w)
					return
				}
				groups[selectedGroupIndex].Criteria = append(groups[selectedGroupIndex].Criteria, entry.Text)
				groups[selectedGroupIndex].CriteriaIDs = append(groups[selectedGroupIndex].CriteriaIDs, resp.CriterionId)
				criteriaList.Refresh()
			}
		}, w)
	})

	deleteCriterionButton := widget.NewButton("Удалить критерий", func() {
		if selectedGroupIndex < 0 {
			dialog.ShowInformation("Ошибка", "Сначала выберите группу", w)
			return
		}
		if len(groups[selectedGroupIndex].Criteria) == 0 {
			dialog.ShowInformation("Ошибка", "Нет критериев для удаления", w)
			return
		}

		criteriaNames := make([]string, len(groups[selectedGroupIndex].Criteria))
		nameToIndex := make(map[string]int)
		for i, name := range groups[selectedGroupIndex].Criteria {
			criteriaNames[i] = name
			nameToIndex[name] = i
		}

		selectEntry := widget.NewSelect(criteriaNames, func(selected string) {})
		selectEntry.PlaceHolder = "Выберите критерий для удаления"
		selectEntryContainer := container.NewVBox(
			widget.NewLabel("Выберите критерий:"),
			selectEntry,
		)

		deleteDialog := dialog.NewCustomConfirm(
			"Удаление критерия",
			"Удалить",
			"Отмена",
			selectEntryContainer,
			func(confirmed bool) {
				if !confirmed || selectEntry.Selected == "" {
					return
				}

				nameIndex, ok := nameToIndex[selectEntry.Selected]
				if !ok {
					dialog.ShowInformation("Ошибка", "Выбранный критерий не найден", w)
					return
				}

				rubricConn, err := grpc.Dial("localhost:50055", grpc.WithInsecure())
				if err != nil {
					log.Printf("Failed to connect to rubricservice: %v", err)
					dialog.ShowError(err, w)
					return
				}
				defer rubricConn.Close()
				rubricClient := rubricpb.NewRubricServiceClient(rubricConn)

				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				resp, err := rubricClient.DeleteCriterion(ctx, &rubricpb.DeleteCriterionRequest{CriterionId: groups[selectedGroupIndex].CriteriaIDs[nameIndex]})
				if err != nil {
					log.Printf("Failed to delete criterion: %v", err)
					dialog.ShowError(err, w)
					return
				}
				if resp.Error != "" {
					log.Println("DeleteCriterion error:", resp.Error)
					dialog.ShowInformation("Ошибка", resp.Error, w)
					return
				}

				groups[selectedGroupIndex].Criteria = append(groups[selectedGroupIndex].Criteria[:nameIndex], groups[selectedGroupIndex].Criteria[nameIndex+1:]...)
				groups[selectedGroupIndex].CriteriaIDs = append(groups[selectedGroupIndex].CriteriaIDs[:nameIndex], groups[selectedGroupIndex].CriteriaIDs[nameIndex+1:]...)
				criteriaList.Refresh()
			},
			w,
		)
		deleteDialog.Show()
	})

	groupList.OnSelected = func(id widget.ListItemID) {
		selectedGroupIndex = id
		criteriaList.Refresh()
		contentContainer.Objects = []fyne.CanvasObject{widget.NewLabel("Выберите критерий")}
		contentContainer.Refresh()
	}

	criteriaList.OnSelected = func(id widget.ListItemID) {
		if selectedGroupIndex >= 0 && selectedGroupIndex < len(groups) && id >= 0 && id < len(groups[selectedGroupIndex].Criteria) {
			criterionId := groups[selectedGroupIndex].CriteriaIDs[id]

			var crit *rubricpb.MainCriteria
			for _, group := range resp.Groups {
				if group.Id == groups[selectedGroupIndex].GroupID {
					for _, c := range group.Criteria {
						if c.Id == criterionId {
							crit = c
							break
						}
					}
					break
				}
			}

			rubricConn, err := grpc.Dial("localhost:50055", grpc.WithInsecure())
			if err != nil {
				log.Printf("Failed to connect to rubricservice: %v", err)
				dialog.ShowError(err, w)
				return
			}
			defer rubricConn.Close()

			weightEntry := widget.NewEntry()
			if crit != nil {
				weightEntry.SetText(strconv.FormatInt(crit.Weight, 10))
			} else {
				weightEntry.SetPlaceHolder("Вес критерия")
			}

			entries := make(map[string]*widget.Entry)
			scores := []string{"000", "025", "050", "075", "100"}
			formItems := []*widget.FormItem{}
			for _, score := range scores {
				entry := widget.NewEntry()
				if crit != nil {
					switch score {
					case "000":
						entry.SetText(crit.Comment_000)
					case "025":
						entry.SetText(crit.Comment_025)
					case "050":
						entry.SetText(crit.Comment_050)
					case "075":
						entry.SetText(crit.Comment_075)
					case "100":
						entry.SetText(crit.Comment_100)
					}
				} else {
					entry.SetPlaceHolder(fmt.Sprintf("Комментарий для %s", score))
				}
				entries[score] = entry
				formItems = append(formItems, &widget.FormItem{
					Text:   fmt.Sprintf("Оценка %s", score),
					Widget: entry,
				})
			}
			formItems = append(formItems, &widget.FormItem{
				Text:   "Вес",
				Widget: weightEntry,
			})

			saveButton := widget.NewButton("Сохранить", func() {
				rubricConn, err := grpc.Dial("localhost:50055", grpc.WithInsecure())
				if err != nil {
					log.Printf("Failed to connect to rubricservice: %v", err)
					dialog.ShowError(err, w)
					return
				}
				defer rubricConn.Close()
				rubricClient := rubricpb.NewRubricServiceClient(rubricConn)

				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				if weightEntry.Text != "" {
					weight, err := strconv.Atoi(weightEntry.Text)
					if err != nil {
						dialog.ShowInformation("Ошибка", "Вес должен быть числом", w)
						return
					}
					respWeight, err := rubricClient.UpdateCriterionWeight(ctx, &rubricpb.UpdateCriterionWeightRequest{
						CriterionId: criterionId,
						Weight:      int32(weight),
					})
					if err != nil {
						log.Printf("Failed to update criterion weight: %v", err)
						dialog.ShowError(err, w)
						return
					}
					if respWeight.Error != "" {
						dialog.ShowInformation("Ошибка", respWeight.Error, w)
						return
					}
				}

				for score, entry := range entries {
					comment := entry.Text
					if comment == "" {
						comment = "Не выставляется"
					}
					respComment, err := rubricClient.UpdateCriterionComment(ctx, &rubricpb.UpdateCriterionCommentRequest{
						CriterionId: criterionId,
						Mark:        score,
						Comment:     comment,
					})
					if err != nil {
						log.Printf("Failed to update criterion comment for score %s: %v", score, err)
						dialog.ShowError(err, w)
						return
					}
					if respComment.Error != "" {
						dialog.ShowInformation("Ошибка", respComment.Error, w)
						return
					}
				}
				dialog.ShowInformation("Успех", "Данные сохранены", w)
				CreateMainCriteriaPage(state, taskID) // Перезагрузка страницы
			})

			content := container.NewVBox(
				widget.NewForm(formItems...),
				saveButton,
			)

			contentContainer.Objects = []fyne.CanvasObject{content}
			contentContainer.Refresh()
		}
	}

	groupContainer := container.NewVBox(groupList, addGroupButton, deleteGroupButton)
	criteriaContainer := container.NewVBox(criteriaList, addCriterionButton, deleteCriterionButton)
	leftPanel := container.NewHSplit(groupContainer, criteriaContainer)
	leftPanel.SetOffset(0.5)
	split := container.NewHSplit(leftPanel, mainContent)

	w.SetContent(container.NewStack(
		canvas.NewRectangle(color.NRGBA{R: 255, G: 255, B: 255, A: 255}),
		container.NewBorder(
			container.NewVBox(headerContent, backButtonContainer),
			nil,
			nil,
			nil,
			split,
		),
	))
}
