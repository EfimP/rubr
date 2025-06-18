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
	gradingpb "rubr/proto/grade"
	rubricpb "rubr/proto/rubric"
	workpb "rubr/proto/work"
	workassignmentpb "rubr/proto/workassignment"
	"strconv"
	"strings"
	"time"
)

type Work struct {
	ID          int32
	Date        time.Time
	Title       string
	StudentName string
	TaskID      int32
	Status      string // Добавляем статус
	Grade       string // Добавляем оценку
}

type Task struct {
	ID    int32
	Date  time.Time
	Title string
}

func CreateSeminaristWorksPage(state *AppState) fyne.CanvasObject {
	headerTextColor := color.White
	darkBlue := color.NRGBA{R: 20, G: 40, B: 80, A: 255}
	separatorColor := color.NRGBA{R: 200, G: 200, B: 200, A: 255}

	// Подключение к сервисам
	workConn, err := grpc.Dial("localhost:50053", grpc.WithInsecure())
	if err != nil {
		log.Printf("Failed to connect to workservice: %v", err)
		return container.NewVBox(widget.NewLabel("Ошибка подключения к серверу"))
	}
	defer workConn.Close()
	workClient := workpb.NewWorkServiceClient(workConn)

	workAssignmentConn, err := grpc.Dial("localhost:50054", grpc.WithInsecure())
	if err != nil {
		log.Printf("Failed to connect to workassignmentservice: %v", err)
		return container.NewVBox(widget.NewLabel("Ошибка подключения к сервису назначений"))
	}
	defer workAssignmentConn.Close()
	workAssignmentClient := workassignmentpb.NewWorkAssignmentServiceClient(workAssignmentConn)

	gradingConn, err := grpc.Dial("localhost:50057", grpc.WithInsecure())
	if err != nil {
		log.Printf("Failed to connect to gradingservice: %v", err)
		return container.NewVBox(widget.NewLabel("Ошибка подключения к сервису оценок"))
	}
	defer gradingConn.Close()
	gradingClient := gradingpb.NewGradingServiceClient(gradingConn)

	rubricConn, err := grpc.Dial("localhost:50055", grpc.WithInsecure())
	if err != nil {
		log.Printf("Failed to connect to rubricservice: %v", err)
		return container.NewVBox(widget.NewLabel("Ошибка подключения к сервису рубрик"))
	}
	defer rubricConn.Close()
	rubricClient := rubricpb.NewRubricServiceClient(rubricConn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Получение списка работ
	studentWorksResp, err := workClient.GetStudentWorksForSeminarist(ctx, &workpb.GetStudentWorksForSeminaristRequest{
		SeminaristId: state.userID,
	})
	if err != nil {
		log.Printf("Failed to load student works: %v", err)
		return container.NewVBox(widget.NewLabel("Ошибка загрузки работ студентов"))
	}
	if studentWorksResp.Error != "" {
		log.Println("Error from server:", studentWorksResp.Error)
		return container.NewVBox(widget.NewLabel("Ошибка: " + studentWorksResp.Error))
	}

	// Обработка списка работ
	studentWorks := make([]Work, len(studentWorksResp.Works))
	for i, w := range studentWorksResp.Works {
		createdAt, _ := time.Parse(time.RFC3339, w.CreatedAt)
		work := Work{
			ID:          w.Id,
			Date:        createdAt,
			Title:       w.Title,
			StudentName: w.StudentName,
			TaskID:      w.TaskId,
		}

		// Получение статуса через workassignmentpb.GetWorkDetails
		detailsResp, err := workAssignmentClient.GetWorkDetails(ctx, &workassignmentpb.GetWorkDetailsRequest{WorkId: w.Id})
		if err != nil {
			log.Printf("Не удалось получить детали работы %d: %v", w.Id, err)
			work.Status = "Ошибка"
		} else if detailsResp.Error != "" {
			log.Printf("Ошибка получения деталей работы %d: %s", w.Id, detailsResp.Error)
			work.Status = "Ошибка"
		} else {
			work.Status = detailsResp.Status
		}

		// Вычисление оценки, если статус 'graded by assistant' или 'graded by seminarist'
		if work.Status == "graded by assistant" || work.Status == "graded by seminarist" {
			// Получение оценок
			marksResp, err := gradingClient.GetCriteriaMarks(ctx, &gradingpb.GetCriteriaMarksRequest{WorkId: w.Id})
			if err != nil {
				log.Printf("Не удалось получить оценки для работы %d: %v", w.Id, err)
				work.Grade = "Ошибка"
			} else if marksResp.Error != "" {
				log.Printf("Ошибка получения оценок для работы %d: %s", w.Id, marksResp.Error)
				work.Grade = "Ошибка"
			} else {
				// Получение блокирующих критериев
				blockingResp, err := rubricClient.LoadTaskBlockingCriterias(ctx, &rubricpb.LoadTaskBlockingCriteriasRequest{TaskId: w.TaskId})
				if err != nil || blockingResp.Error != "" {
					log.Printf("Не удалось загрузить блокирующие критерии для задачи %d: %v", w.TaskId, err)
					work.Grade = "Ошибка"
				} else {
					// Получение основных критериев
					mainResp, err := rubricClient.LoadTaskMainCriterias(ctx, &rubricpb.LoadTaskMainCriteriasRequest{TaskId: w.TaskId})
					if err != nil || mainResp.Error != "" {
						log.Printf("Не удалось загрузить основные критерии для задачи %d: %v", w.TaskId, err)
						work.Grade = "Ошибка"
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
							work.Grade = fmt.Sprintf("%.2f", minBlockingMark)
						} else {
							// Вычисление оценки по основным критериям
							totalMark := float32(0)
							totalMaxMark := float32(0)
							for _, mark := range marksResp.Marks {
								for _, group := range mainResp.Groups {
									for _, crit := range group.Criteria {
										if mark.CriterionId == crit.Id {
											totalMark += mark.Mark * float32(crit.Weight)
										}
									}
								}
							}
							for _, group := range mainResp.Groups {
								for _, crit := range group.Criteria {
									totalMaxMark += float32(crit.Weight)
								}
							}
							if totalMaxMark > 0 {
								finalGrade := (totalMark / totalMaxMark) * 10
								work.Grade = fmt.Sprintf("%.2f", finalGrade)
							} else {
								work.Grade = "0.00"
							}
						}
					}
				}
			}
		}

		studentWorks[i] = work
	}

	// Получение задач
	tasksResp, err := workClient.GetTasksForSeminarist(ctx, &workpb.GetTasksForSeminaristRequest{
		SeminaristId: state.userID,
	})
	if err != nil {
		log.Printf("Failed to load tasks: %v", err)
		return container.NewVBox(widget.NewLabel("Ошибка загрузки задач"))
	}
	if tasksResp.Error != "" {
		log.Println("Error from server:", tasksResp.Error)
		return container.NewVBox(widget.NewLabel("Ошибка: " + tasksResp.Error))
	}

	tasks := make([]Task, len(tasksResp.Tasks))
	for i, t := range tasksResp.Tasks {
		deadline, _ := time.Parse(time.RFC3339, t.Deadline)
		tasks[i] = Task{
			ID:    t.Id,
			Date:  deadline,
			Title: t.Title,
		}
	}

	// Настройка заголовка
	logo := canvas.NewText("ВШЭ", headerTextColor)
	logo.TextStyle.Bold = true
	logoContainer := container.New(layout.NewMaxLayout(), logo)

	headerTitleText := canvas.NewText("Панель семинариста", headerTextColor)
	headerTitleText.TextStyle.Bold = true
	headerTitleText.Alignment = fyne.TextAlignCenter

	header := container.New(layout.NewBorderLayout(nil, nil, logoContainer, nil),
		logoContainer,
		container.NewCenter(headerTitleText),
	)
	headerBackground := canvas.NewRectangle(darkBlue)
	headerWithBackground := container.NewStack(headerBackground, header)

	// Кнопка выхода
	backButton := widget.NewButton("Выйти из аккаунта", func() {
		log.Println("Кнопка 'Выйти из аккаунта' нажата. Возврат на экран авторизации.")
		state.currentPage = "greeting"
		state.window.SetContent(createContent(state))
	})
	backButtonRow := container.New(layout.NewMaxLayout(), backButton)

	// Создание вкладок
	studentWorksContent := createWorksTable(studentWorks, separatorColor, state)
	studentWorksTab := container.NewTabItem("Работы студентов", studentWorksContent)

	tasksContent := createTasksTable(tasks, separatorColor, state)
	tasksTab := container.NewTabItem("Работы от лектора", tasksContent)

	tabs := container.NewAppTabs(studentWorksTab, tasksTab)

	centralContentPanel := container.NewVBox(
		backButtonRow,
		tabs,
	)

	contentBackground := canvas.NewRectangle(color.White)
	centralContentWithBackground := container.NewStack(contentBackground, centralContentPanel)

	return container.NewBorder(
		headerWithBackground,
		nil,
		nil,
		nil,
		centralContentWithBackground,
	)
}

func createWorksTable(works []Work, separatorColor color.Color, state *AppState) fyne.CanvasObject {
	var tableContent []fyne.CanvasObject
	for i, work := range works {
		titleLabel := widget.NewLabel(work.Title)
		studentName := widget.NewLabel(work.StudentName)
		date := widget.NewLabel(work.Date.Format("02.01.2006"))
		statusLabel := widget.NewLabel("Статус: " + work.Status)
		gradeLabel := widget.NewLabel("Оценка: " + work.Grade)
		if work.Grade == "" {
			gradeLabel.SetText("Оценка: -")
		}

		detailsButton := widget.NewButton("Перейти", func(w Work) func() {
			return func() {
				log.Printf("Перейти к работе ID: %d, TaskID: %d", w.ID, w.TaskID)
				state.window.SetContent(CreateSeminaristBlockingCriteriaGradingPage(state, w.ID, w.TaskID))
			}
		}(work))

		row := container.NewHBox(
			titleLabel,
			studentName,
			date,
			statusLabel,
			gradeLabel,
			detailsButton,
		)

		tableContent = append(tableContent, row)

		if i < len(works)-1 {
			separator := canvas.NewLine(separatorColor)
			separator.StrokeWidth = 2
			separator.Position1 = fyne.NewPos(0, 0)
			separator.Position2 = fyne.NewPos(1920, 0)
			separatorContainer := container.New(layout.NewMaxLayout(), separator)
			tableContent = append(tableContent, separatorContainer)
		}
	}

	return container.NewVBox(tableContent...)
}
func createTasksTable(tasks []Task, separatorColor color.Color, state *AppState) fyne.CanvasObject {
	var tableContent []fyne.CanvasObject
	for i, task := range tasks {
		titleLabel := widget.NewLabel(task.Title)
		date := widget.NewLabel(task.Date.Format("02.01.2006"))

		detailsButton := widget.NewButton("Перейти", func(t Task) func() {
			return func() {
				log.Printf("Перейти к задаче ID: %d", t.ID)
				state.window.SetContent(CreateTaskDetailsPage(state, t.ID))
			}
		}(task))

		row := container.NewHBox(
			titleLabel,
			date,
			detailsButton,
		)

		tableContent = append(tableContent, row)

		if i < len(tasks)-1 {
			separator := canvas.NewLine(separatorColor)
			separator.StrokeWidth = 2
			separator.Position1 = fyne.NewPos(0, 0)
			separator.Position2 = fyne.NewPos(1920, 0)
			separatorContainer := container.New(layout.NewMaxLayout(), separator)
			tableContent = append(tableContent, separatorContainer)
		}
	}

	return container.NewVBox(tableContent...)
}

func CreateTaskDetailsPage(state *AppState, taskID int32) fyne.CanvasObject {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.Dial("localhost:50053", grpc.WithInsecure())
	if err != nil {
		log.Printf("Не удалось подключиться к сервису: %v", err)
		return container.NewVBox(widget.NewLabel("Ошибка подключения к сервису"))
	}
	defer conn.Close()

	client := workpb.NewWorkServiceClient(conn)
	resp, err := client.GetTaskDetails(ctx, &workpb.GetTaskDetailsRequest{TaskId: taskID})
	if err != nil {
		log.Printf("Не удалось получить детали работы: %v", err)
		return container.NewVBox(widget.NewLabel("Ошибка загрузки деталей работы"))
	}
	if resp.Error != "" {
		log.Printf("Ошибка от сервиса: %s", resp.Error)
		return container.NewVBox(widget.NewLabel(resp.Error))
	}

	headerTextColor := color.White
	darkBlue := color.NRGBA{R: 20, G: 40, B: 80, A: 255}

	logo := canvas.NewText("ВШЭ", headerTextColor)
	logo.TextStyle.Bold = true
	logo.TextSize = 24
	logoContainer := container.New(layout.NewMaxLayout(), logo)

	headerTitleText := canvas.NewText("Задание", headerTextColor)
	headerTitleText.TextStyle.Bold = true
	headerTitleText.TextSize = 20
	headerTitleText.Alignment = fyne.TextAlignCenter

	header := container.New(layout.NewBorderLayout(nil, nil, logoContainer, nil),
		logoContainer,
		container.NewCenter(headerTitleText),
	)
	headerBackground := canvas.NewRectangle(darkBlue)
	headerWithBackground := container.NewStack(headerBackground, header)

	titleLabel := widget.NewLabel("Название: " + resp.Title)
	titleLabel.TextStyle.Bold = true

	descriptionEntry := widget.NewMultiLineEntry()
	descriptionEntry.SetText(resp.Description)
	descriptionEntry.Disable()
	scrollableDescription := container.NewVScroll(descriptionEntry)
	scrollableDescription.SetMinSize(fyne.NewSize(0, descriptionEntry.MinSize().Height*5))

	deadlineTime, err := time.Parse(time.RFC3339, resp.Deadline)
	if err != nil {
		log.Printf("Ошибка парсинга дедлайна: %v", err)
		return container.NewVBox(widget.NewLabel("Ошибка обработки дедлайна"))
	}
	deadlineLabel := widget.NewLabel("Дедлайн: " + deadlineTime.Format("02.01.2006"))

	lectorName := resp.LectorName
	if resp.LectorPatronymic != "" {
		lectorName += " " + resp.LectorPatronymic
	}
	lectorLabel := widget.NewLabel("Лектор: " + lectorName + " " + resp.LectorSurname)

	assignButton := widget.NewButton("Назначить ассистентов", func() {
		log.Println("Открытие экрана назначения ассистентов")
		state.window.SetContent(CreateAssistantAssignmentPage(state, taskID))
	})

	backButton := widget.NewButton("Назад", func() {
		log.Println("Кнопка 'Назад' нажата. Возврат на панель семинариста.")
		state.currentPage = "seminarist_works"
		state.window.SetContent(createContent(state))
	})

	buttonsContainer := container.NewHBox(backButton, layout.NewSpacer(), assignButton)

	inputGrid := container.NewVBox(
		titleLabel,
		scrollableDescription,
		deadlineLabel,
		lectorLabel,
	)

	contentBackground := canvas.NewRectangle(color.White)
	contentWithPadding := container.NewPadded(inputGrid)
	centralContent := container.NewStack(contentBackground, contentWithPadding)

	return container.NewBorder(
		headerWithBackground,
		buttonsContainer,
		nil,
		nil,
		centralContent,
	)
}

func CreateAssistantAssignmentPage(state *AppState, taskID int32) fyne.CanvasObject {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.Dial("localhost:50053", grpc.WithInsecure())
	if err != nil {
		log.Printf("Не удалось подключиться к сервису: %v", err)
		return container.NewVBox(widget.NewLabel("Ошибка подключения к сервису"))
	}
	defer conn.Close()

	client := workpb.NewWorkServiceClient(conn)

	startTime := time.Now()
	log.Printf("Start loading data for taskID: %d", taskID)

	// Получение работ студентов
	worksResp, err := client.GetStudentWorksByTask(ctx, &workpb.GetStudentWorksByTaskRequest{TaskId: taskID})
	if err != nil {
		log.Printf("Не удалось получить работы студентов: %v", err)
		return container.NewVBox(widget.NewLabel("Ошибка загрузки работ студентов"))
	}
	if worksResp.Error != "" {
		log.Printf("Ошибка от сервиса: %s", worksResp.Error)
		return container.NewVBox(widget.NewLabel(worksResp.Error))
	}
	log.Printf("Loaded %d student works in %v", len(worksResp.Works), time.Since(startTime))

	// Получение discipline_id задачи
	taskDetailsResp, err := client.GetTaskDetails(ctx, &workpb.GetTaskDetailsRequest{TaskId: taskID})
	if err != nil {
		log.Printf("Не удалось получить детали задачи: %v", err)
		return container.NewVBox(widget.NewLabel("Ошибка загрузки деталей задачи"))
	}
	if taskDetailsResp.Error != "" {
		log.Printf("Ошибка от сервиса: %s", taskDetailsResp.Error)
		return container.NewVBox(widget.NewLabel(taskDetailsResp.Error))
	}
	log.Printf("Loaded task details in %v", time.Since(startTime))

	// Получение ассистентов
	assistantsResp, err := client.GetAssistantsByDiscipline(ctx, &workpb.GetAssistantsByDisciplineRequest{DisciplineId: taskDetailsResp.DisciplineId})
	if err != nil {
		log.Printf("Не удалось получить ассистентов: %v", err)
		return container.NewVBox(widget.NewLabel("Ошибка загрузки ассистентов"))
	}
	if assistantsResp.Error != "" {
		log.Printf("Ошибка от сервиса: %s", assistantsResp.Error)
		return container.NewVBox(widget.NewLabel(assistantsResp.Error))
	}
	log.Printf("Loaded %d assistants in %v", len(assistantsResp.Assistants), time.Since(startTime))

	headerTextColor := color.White
	darkBlue := color.NRGBA{R: 20, G: 40, B: 80, A: 255}
	separatorColor := color.NRGBA{R: 200, G: 200, B: 200, A: 255}

	logo := canvas.NewText("ВШЭ", headerTextColor)
	logo.TextStyle.Bold = true
	logo.TextSize = 28
	logoContainer := container.New(layout.NewMaxLayout(), logo)

	headerTitleText := canvas.NewText("Назначение ассистентов", headerTextColor)
	headerTitleText.TextStyle.Bold = true
	headerTitleText.Alignment = fyne.TextAlignCenter

	header := container.New(layout.NewBorderLayout(nil, nil, logoContainer, nil),
		logoContainer,
		container.NewCenter(headerTitleText),
	)
	headerBackground := canvas.NewRectangle(darkBlue)
	headerWithBackground := container.NewStack(headerBackground, header)

	backButton := widget.NewButton("Назад", func() {
		log.Println("Кнопка 'Назад' нажата. Возврат на экран задания.")
		state.window.SetContent(CreateTaskDetailsPage(state, taskID))
	})
	backButtonRow := container.New(layout.NewMaxLayout(), backButton)

	sidebarBackground := canvas.NewRectangle(darkBlue)
	sidebar := container.New(layout.NewMaxLayout(), sidebarBackground, container.NewVBox(layout.NewSpacer(), layout.NewSpacer()))

	// Подготовка данных для списка
	works := worksResp.Works
	assistants := assistantsResp.Assistants
	assistantNames := make([]string, len(assistants))
	assistantIDs := make([]int32, len(assistants))
	for i, a := range assistants {
		name := a.Name
		if a.Patronymic != "" {
			name += " " + a.Patronymic
		}
		name += " " + a.Surname
		assistantNames[i] = name
		assistantIDs[i] = a.Id
	}

	// Срез для хранения выпадающих списков
	selectWidgets := make([]*widget.Select, len(works))

	// Создание списка студентов
	var listContent []fyne.CanvasObject
	for i, work := range works {
		// Формирование ФИО студента
		name := work.StudentName
		if work.StudentPatronymic != "" {
			name += " " + work.StudentPatronymic
		}
		name += " " + work.StudentSurname
		nameLabel := widget.NewLabel(name)
		nameLabel.Wrapping = fyne.TextWrapWord
		nameLabel.Resize(fyne.NewSize(500, nameLabel.MinSize().Height))

		emailLabel := widget.NewLabel(work.StudentEmail)
		emailLabel.Wrapping = fyne.TextWrapWord
		emailLabel.Resize(fyne.NewSize(500, emailLabel.MinSize().Height))

		selectAssistant := widget.NewSelect(assistantNames, func(selected string) {
			if selected != "" {
				log.Println(time.Now().Format("15:04:05"), "Выбран ассистент для", name, ":", selected)
			}
		})
		if len(assistantNames) > 0 {
			selectAssistant.SetSelectedIndex(0)
		}
		selectWidgets[i] = selectAssistant
		selectAssistant.Resize(fyne.NewSize(300, selectAssistant.MinSize().Height))

		row := container.NewGridWithColumns(3,
			nameLabel,
			emailLabel,
			selectAssistant,
		)

		listContent = append(listContent, row)

		if i < len(works)-1 {
			separator := canvas.NewLine(separatorColor)
			separator.StrokeWidth = 2
			separator.Position1 = fyne.NewPos(0, 0)
			separator.Position2 = fyne.NewPos(1920, 0)
			separatorContainer := container.New(layout.NewMaxLayout(), separator)
			listContent = append(listContent, separatorContainer)
		}
	}

	// Создание прокручиваемого списка
	studentList := container.NewVBox(listContent...)
	listScroll := container.NewVScroll(studentList)
	listScroll.SetMinSize(fyne.NewSize(1400, 560))

	nextButton := widget.NewButton("Далее", func() {
		allAssigned := true
		var unassignedStudents []string
		assignments := make([]*workpb.AssignAssistantsToWorksRequest_Assignment, len(works))
		for i := 0; i < len(works); i++ {
			selectAssistant := selectWidgets[i]
			if selectAssistant.Selected == "" {
				allAssigned = false
				name := works[i].StudentName
				if works[i].StudentPatronymic != "" {
					name += " " + works[i].StudentPatronymic
				}
				name += " " + works[i].StudentSurname
				unassignedStudents = append(unassignedStudents, name)
			} else {
				// Найти ID ассистента
				var assistantID int32
				for j, name := range assistantNames {
					if name == selectAssistant.Selected {
						assistantID = assistantIDs[j]
						break
					}
				}
				assignments[i] = &workpb.AssignAssistantsToWorksRequest_Assignment{
					WorkId:      works[i].Id,
					AssistantId: assistantID,
				}
			}
		}

		if !allAssigned {
			popupContent := container.NewVBox(
				widget.NewLabel("Не все ассистенты назначены!"),
				widget.NewLabel("Назначьте ассистентов для: "+strings.Join(unassignedStudents, ", ")),
				widget.NewButton("OK", func() {}),
			)
			popup := widget.NewPopUp(popupContent, state.window.Canvas())
			popup.Show()
		} else {
			// Создаём новый контекст для AssignAssistantsToWorks
			assignCtx, assignCancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer assignCancel()

			// Создаём новое gRPC-соединение для AssignAssistantsToWorks
			assignConn, err := grpc.Dial("localhost:50053", grpc.WithInsecure())
			if err != nil {
				log.Printf("Не удалось подключиться к сервису для назначения ассистентов: %v", err)
				popupContent := container.NewVBox(
					widget.NewLabel("Ошибка подключения к сервису: "+err.Error()),
					widget.NewButton("OK", func() {}),
				)
				popup := widget.NewPopUp(popupContent, state.window.Canvas())
				popup.Show()
				return
			}
			defer assignConn.Close()

			assignClient := workpb.NewWorkServiceClient(assignConn)

			log.Printf("Sending %d assignments", len(assignments))
			startAssign := time.Now()

			// Отправка назначений
			_, err = assignClient.AssignAssistantsToWorks(assignCtx, &workpb.AssignAssistantsToWorksRequest{
				Assignments: assignments,
			})
			if err != nil {
				log.Printf("Ошибка при назначении ассистентов: %v", err)
				popupContent := container.NewVBox(
					widget.NewLabel("Ошибка при назначении ассистентов: "+err.Error()),
					widget.NewButton("OK", func() {}),
				)
				popup := widget.NewPopUp(popupContent, state.window.Canvas())
				popup.Show()
				return
			}

			log.Printf("Assignments completed in %v", time.Since(startAssign))

			state.currentPage = "seminarist_works"
			state.window.SetContent(CreateSeminaristWorksPage(state))
		}
	})

	// Контейнер для списка и кнопки
	assignmentContent := container.NewVBox(listScroll, nextButton)
	centralContent := container.NewHBox(sidebar, assignmentContent)
	centralContentPanel := container.NewVBox(backButtonRow, centralContent)
	contentBackground := canvas.NewRectangle(color.White)
	centralContentWithBackground := container.NewStack(contentBackground, centralContentPanel)

	return container.NewBorder(
		headerWithBackground,
		nil,
		nil,
		nil,
		centralContentWithBackground,
	)
}

func CreateSeminaristBlockingCriteriaGradingPage(state *AppState, workID int32, taskID int32) fyne.CanvasObject {
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
		state.currentPage = "seminarist_works"
		w.SetContent(CreateSeminaristWorksPage(state))
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
			log.Printf("Оценка не найдена для блокирующего критерия ID %d, установлена оценка по умолчанию %d и комментарий лектора: %s", crit.Id, crit.FinalMark, crit.Comment)
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
		w.SetContent(CreateSeminaristMainCriteriaGradingPage(state, workID, taskID))
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

func CreateSeminaristMainCriteriaGradingPage(state *AppState, workID int32, taskID int32) fyne.CanvasObject {
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

	workConn, err := grpc.Dial("localhost:50053", grpc.WithInsecure())
	if err != nil {
		log.Printf("Не удалось подключиться к WorkService: %v", err)
		return container.NewVBox(widget.NewLabel("Ошибка подключения к сервису"))
	}
	defer workConn.Close()

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
		weight    int64
	}
	criteriaInfo := make(map[int32]criterionInfo)
	totalCriteriaCount := 0
	totalMaxScore := float64(0)

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
				weight:    crit.Weight,
			}
			totalCriteriaCount++
			totalMaxScore += float64(crit.Weight)
		}
	}

	log.Printf("Инициализировано %d записей критериев с максимальным баллом %.2f", len(entries), totalMaxScore)

	contentContainer := container.New(layout.NewMaxLayout(), widget.NewLabel("Выберите группу и критерий"))

	// Вычисление текущего балла
	calculateCurrentScore := func() (float64, string) {
		currentScore := float64(0)
		for _, entry := range entries {
			if entry.Select.Selected != "" {
				mark, err := strconv.ParseFloat(entry.Select.Selected, 64)
				if err == nil {
					currentScore += mark * float64(criteriaInfo[entry.CriterionID].weight)
				}
			}
		}
		// Оценка
		grade := (currentScore / totalMaxScore) * 10
		return currentScore, fmt.Sprintf("%.2f", grade)
	}

	// Метка для отображения текущего балла и оценки
	currentScore, currentGrade := calculateCurrentScore()
	scoreLabel := widget.NewLabel(fmt.Sprintf("Текущий основной балл: %.2f/%.2f\nОценка: %s", currentScore, totalMaxScore, currentGrade))

	// Обновление метки при изменении оценок
	updateScoreLabel := func() {
		currentScore, currentGrade := calculateCurrentScore()
		scoreLabel.SetText(fmt.Sprintf("Текущий балл: %.2f/%.2f\nОценка: %s%%", currentScore, totalMaxScore, currentGrade))
	}

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
				item.(*widget.Label).SetText(fmt.Sprintf("%s (Вес: %d)", resp.Groups[selectedGroupIndex].Criteria[id].Name, resp.Groups[selectedGroupIndex].Criteria[id].Weight))
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

			// Находим соответствующий entry
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

			// Обновляем обработчик выбора оценки
			selectWidget := selectedEntry.Select
			selectWidget.OnChanged = func(s string) {
				if s != "" {
					updateScoreLabel()
				}
			}

			// Создаем контейнер для отображения комментариев
			commentsContainer := container.NewVBox(
				widget.NewLabelWithStyle(fmt.Sprintf("Комментарии лектора (Вес: %d):", crit.Weight), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
				widget.NewLabel(fmt.Sprintf("Для 0.0: %s", crit.Comment_000)),
				widget.NewLabel(fmt.Sprintf("Для 0.25: %s", crit.Comment_025)),
				widget.NewLabel(fmt.Sprintf("Для 0.50: %s", crit.Comment_050)),
				widget.NewLabel(fmt.Sprintf("Для 0.75: %s", crit.Comment_075)),
				widget.NewLabel(fmt.Sprintf("Для 1.00: %s", crit.Comment_100)),
			)

			content := container.NewVBox(
				widget.NewLabel(fmt.Sprintf("Критерий: %s (Вес: %d)", crit.Name, crit.Weight)),
				container.NewHBox(widget.NewLabel("Оценка:"), selectWidget),
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
		state.currentPage = "seminarist_blocking"
		w.SetContent(CreateSeminaristBlockingCriteriaGradingPage(state, workID, taskID))
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

		// Обновление статуса работы и очистка assistant_id
		workConn, err := grpc.Dial("localhost:50053", grpc.WithInsecure())
		if err != nil {
			log.Printf("Не удалось подключиться к WorkService: %v", err)
			dialog.ShowError(err, w)
			return
		}
		defer workConn.Close()
		workClient := workpb.NewWorkServiceClient(workConn)

		start := time.Now()
		_, err = workClient.UpdateWork(ctx, &workpb.UpdateWorkRequest{
			WorkId:      workID,
			Status:      "graded by seminarist",
			AssistantId: 0, // Очистка assistant_id
		})
		if err != nil {
			log.Printf("Не удалось обновить статус работы %d: %v (время выполнения: %v)", workID, err, time.Since(start))
			dialog.ShowError(err, w)
			return
		}

		log.Printf("Статус работы %d обновлен на 'graded by seminarist', assistant_id очищен (время выполнения: %v)", workID, time.Since(start))
		dialog.ShowInformation("Успех", "Оценка завершена", w)
		state.currentPage = "seminarist_works"
		w.SetContent(CreateSeminaristWorksPage(state))
	})

	// Нижняя панель с меткой балла и кнопками
	bottomButtons := container.NewHBox(
		backButton,
		layout.NewSpacer(),
		scoreLabel,
		finalizeButton,
	)
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
