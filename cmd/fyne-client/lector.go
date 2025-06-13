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

func CreateLectorWorksPage(state *AppState, leftBackground *canvas.Image) fyne.CanvasObject {
	// Не создаем соединение здесь, передаем клиента в CreateWorkPage
	userIDint64, err := strconv.ParseInt(state.userID, 10, 32)
	if err != nil {
		log.Printf("Invalid user ID: %v", err)
		return container.NewVBox(widget.NewLabel("Ошибка: некорректный ID пользователя"))
	}
	userID := int32(userIDint64)

	conn, err := grpc.Dial("localhost:50053", grpc.WithInsecure())
	if err != nil {
		log.Printf("Failed to connect to gRPC: %v", err)
		return container.NewVBox(widget.NewLabel("Ошибка подключения к серверу"))
	}
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
		data = append(data, MyListItem{
			ID:      task.Id,
			Name:    task.Title,
			DueDate: task.Deadline,
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
			fmt.Printf("Изменение работы '%s' (ID: %d)\n", data[currentID].Name, data[currentID].ID)
		}

		deleteButton.OnTapped = func() {
			dialog.ShowConfirm(
				"Подтверждение удаления",
				fmt.Sprintf("Вы уверены, что хотите удалить работу '%s'?", data[currentID].Name),
				func(confirmed bool) {
					if confirmed {
						ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
						defer cancel()
						_, err := workClient.DeleteTask(ctx, &workpb.DeleteTaskRequest{TaskId: data[currentID].ID})
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
		CreateWorkPage(state, workClient, conn, nil)
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

func CreateWorkPage(state *AppState, workClient workpb.WorkServiceClient, conn *grpc.ClientConn, taskID *int32) {
	w := state.window
	var isNewWork bool
	if taskID == nil {
		isNewWork = true
	}

	headerTextColor := color.White
	logoText := canvas.NewText("ВШЭ", headerTextColor)
	logoText.TextStyle.Bold = true
	logoText.TextSize = 24
	logoText.Alignment = fyne.TextAlignCenter
	leftHeaderObject := container.NewStack(logoText)

	headerTitle := canvas.NewText("Создать работу", headerTextColor)
	headerTitle.TextStyle.Bold = true
	headerTitle.TextSize = 20
	headerTitle.Alignment = fyne.TextAlignCenter

	headerContent := container.New(layout.NewBorderLayout(nil, nil, leftHeaderObject, nil),
		leftHeaderObject,
		container.NewCenter(headerTitle),
	)

	titleEntry := widget.NewEntry()
	titleEntry.SetPlaceHolder("Название")

	descriptionEntry := widget.NewMultiLineEntry()
	descriptionEntry.SetPlaceHolder("Описание работы")
	scrollableDescription := container.NewVScroll(descriptionEntry)
	scrollableDescription.SetMinSize(fyne.NewSize(0, descriptionEntry.MinSize().Height*5))

	dateAndTimeEntry := widget.NewEntry()
	dateAndTimeEntry.SetPlaceHolder("Выберите дату и время дедлайна")
	dateAndTimeEntry.Disable()

	var selectedDateTime time.Time
	var isDateTimeSelected bool

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
		scrollableDescription,
	)

	nextButton := widget.NewButton("Далее", func() {
		if !isDateTimeSelected {
			dialog.ShowInformation("Ошибка", "Пожалуйста, выберите дедлайн", w)
			return
		}
		deadline := selectedDateTime.Format(time.RFC3339)

		userIDint64, err := strconv.ParseInt(state.userID, 10, 32)
		if err != nil {
			log.Printf("Invalid user ID: %v", err)
			dialog.ShowInformation("Ошибка", "Некорректный ID пользователя", w)
			return
		}
		userID := int32(userIDint64)

		if isNewWork {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			resp, err := workClient.CreateWork(ctx, &workpb.CreateWorkRequest{
				LectorId:     userID,
				GroupId:      1, // Замените на реальный group_id
				Title:        titleEntry.Text,
				Description:  descriptionEntry.Text,
				Deadline:     deadline,
				DisciplineId: 1, // Замените на реальный discipline_id
				ContentUrl:   "",
			})
			if err != nil {
				log.Printf("Failed to create work: %v", err)
				dialog.ShowError(err, w)
				return
			}
			if resp.Error != "" {
				log.Println("CreateWork error:", resp.Error)
				dialog.ShowInformation("Ошибка", resp.Error, w)
				return
			}
			taskID := resp.TaskId
			ShowBlockingCriteriaPage(state, workClient, conn, taskID)
		}
	})
	nextButtonContainer := container.New(layout.NewHBoxLayout(), layout.NewSpacer(), nextButton)

	contentBackground := canvas.NewRectangle(color.White)
	contentWithPadding := container.NewPadded(inputGrid)
	centralContent := container.NewStack(contentBackground, contentWithPadding)

	w.SetContent(container.NewStack(
		canvas.NewRectangle(color.NRGBA{R: 20, G: 40, B: 80, A: 255}),
		container.NewBorder(
			headerContent,
			nextButtonContainer,
			nil,
			nil,
			centralContent,
		),
	))
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
func ShowBlockingCriteriaPage(state *AppState, workClient workpb.WorkServiceClient, conn *grpc.ClientConn, taskID int32) {
	w := state.window

	rubricConn, err := grpc.Dial("localhost:50053", grpc.WithInsecure())
	if err != nil {
		log.Printf("Failed to connect to gRPC: %v", err)
		dialog.ShowError(err, w)
		return
	}
	defer rubricConn.Close()
	rubricClient := rubricpb.NewRubricServiceClient(rubricConn)

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
		CreateWorkPage(state, workClient, conn, &taskID)
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
	addCriterionEntry := func() {
		nameEntry := widget.NewEntry()
		nameEntry.SetPlaceHolder("Название")
		nameEntryContainer := container.NewMax(nameEntry)
		nameEntryContainer.Resize(fyne.NewSize(250, 60))

		descriptionEntry := widget.NewEntry()
		descriptionEntry.SetPlaceHolder("Описание")
		descriptionEntryContainer := container.NewMax(descriptionEntry)
		descriptionEntryContainer.Resize(fyne.NewSize(250, 60))

		commentEntry := widget.NewEntry()
		commentEntry.SetPlaceHolder("Комментарий")
		commentEntryContainer := container.NewMax(commentEntry)
		commentEntryContainer.Resize(fyne.NewSize(250, 60))

		evaluationEntry := widget.NewEntry()
		evaluationEntry.SetPlaceHolder("Оценка")
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
			NameEntry:        nameEntry,
			DescriptionEntry: descriptionEntry,
			CommentEntry:     commentEntry,
			EvaluationEntry:  evaluationEntry,
			Container:        criterionRow,
		})
	}

	deleteCriterionEntry := func() {
		if len(activeCriteria) > 0 {
			lastIndex := len(activeCriteria) - 1
			lastCriterion := activeCriteria[lastIndex]
			criteriaListContainer.Remove(lastCriterion.Container)
			activeCriteria = activeCriteria[:lastIndex]
			criteriaListContainer.Refresh()
		}
	}

	addCriterionEntry()

	listLabel := widget.NewLabelWithStyle("Список блокирующих критериев", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	scrollableCriteria := container.NewVScroll(criteriaListContainer)
	scrollableCriteria.SetMinSize(fyne.NewSize(0, 400))

	addButton := widget.NewButton("Добавить", func() { addCriterionEntry() })
	deleteButton := widget.NewButton("Удалить", func() { deleteCriterionEntry() })
	nextButton := widget.NewButton("Далее", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		for _, criterion := range activeCriteria {
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
		state.currentPage = "lector_works"
		w.SetContent(createContent(state))
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
