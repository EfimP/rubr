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
	"net/http"
	"net/url"
	"strconv"
	"time"

	gradingpb "rubr/proto/grade"
	pbGrade "rubr/proto/grade"
	rubricpb "rubr/proto/rubric"
	pbWork "rubr/proto/work"
	workassignmentpb "rubr/proto/workassignment"
)

// Структура для хранения предмета
type Subject struct {
	Name    string
	Grades  []float32
	Average float32
	Details string
}

// Структура для хранения задания
type Assignment struct {
	Title       string
	Description string
	Deadline    time.Time
	Submission  time.Time
	FilePath    string
}

// Структура для хранения работы
var WorkID int32
var TaskID int32
var prevPage string

func getStudentWorks(studentID int32) (*pbWork.ListWorksForStudentResponse, error) {
	connWork, err := grpc.Dial("89.169.39.161:50053", grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer connWork.Close()

	workClient := pbWork.NewWorkServiceClient(connWork)
	return workClient.ListWorksForStudent(context.Background(), &pbWork.ListWorksForStudentRequest{StudentId: studentID})
}

func СreateStudentGradesPage(state *AppState) fyne.CanvasObject {
	myWindow := state.window

	// Цвета
	headerTextColor := color.White
	darkBlue := color.NRGBA{R: 20, G: 40, B: 80, A: 255}

	// Верхняя панель (Header)
	logo := canvas.NewText("ВШЭ", headerTextColor)
	logo.TextStyle.Bold = true
	logoContainer := container.New(layout.NewMaxLayout(), logo)

	headerTitleText := canvas.NewText("ОЦЕНКИ СТУДЕНТА", headerTextColor)
	headerTitleText.TextStyle.Bold = true
	headerTitleText.Alignment = fyne.TextAlignCenter

	header := container.New(layout.NewBorderLayout(nil, nil, logoContainer, nil),
		logoContainer,
		container.NewCenter(headerTitleText),
	)
	headerBackground := canvas.NewRectangle(darkBlue)
	headerWithBackground := container.NewStack(headerBackground, header)

	// Кнопка "назад"
	backButton := widget.NewButton("назад", func() {
		dialog.ShowConfirm(
			"Подтверждение",
			"Выйти из приложения?",
			func(ok bool) {
				if ok {
					state.currentPage = "greeting"
					myWindow.SetContent(createContent(state))
					return
				}
			},
			myWindow,
		)
	})
	backButtonRow := container.New(layout.NewMaxLayout(), backButton)

	// Подключение к WorkService
	connWork, err := grpc.Dial("89.169.39.161:50053", grpc.WithInsecure())
	if err != nil {
		log.Printf("Не удалось подключиться к WorkService: %v", err)
		return container.NewVBox(widget.NewLabel("Ошибка подключения к серверу работ"))
	}
	defer connWork.Close()
	workClient := pbWork.NewWorkServiceClient(connWork)

	// Подключение к GradingService
	connGrade, err := grpc.Dial("89.169.39.161:50057", grpc.WithInsecure())
	if err != nil {
		log.Printf("Не удалось подключиться к GradingService: %v", err)
		return container.NewVBox(widget.NewLabel("Ошибка подключения к серверу оценок"))
	}
	defer connGrade.Close()
	gradingClient := pbGrade.NewGradingServiceClient(connGrade)

	// Получение ID студента
	userIDint64, err := strconv.ParseInt(state.userID, 10, 32)
	if err != nil {
		log.Printf("Некорректный ID пользователя: %v", err)
		return container.NewVBox(widget.NewLabel("Ошибка: некорректный ID пользователя"))
	}
	userID := int32(userIDint64)

	// Получение дисциплин студента
	discResp, err := workClient.GetStudentDisciplines(context.Background(), &pbWork.GetStudentDisciplinesRequest{StudentId: userID})
	if err != nil || discResp.Error != "" {
		log.Printf("Ошибка получения дисциплин для student_id %d: %v", userID, err)
		return container.NewVBox(widget.NewLabel("Ошибка загрузки дисциплин"))
	}

	var tableContent []fyne.CanvasObject
	for _, discipline := range discResp.Disciplines {
		// Получение работ по дисциплине
		worksResp, err := workClient.GetStudentWorksByDiscipline(context.Background(), &pbWork.GetStudentWorksByDisciplineRequest{
			StudentId:    userID,
			DisciplineId: discipline.Id,
		})
		if err != nil || worksResp.Error != "" {
			log.Printf("Ошибка получения работ для discipline_id %d: %v", discipline.Id, err)
			continue
		}
		log.Printf("%s: %v", discipline.Name, worksResp.Works)

		rubricConn, err := grpc.Dial("89.169.39.161:50055", grpc.WithInsecure())
		if err != nil {
			log.Printf("Не удалось подключиться к сервису рубрик: %v", err)
			return container.NewVBox(widget.NewLabel("Ошибка подключения к сервису рубрик"))
		}
		defer rubricConn.Close()
		rubricClient := rubricpb.NewRubricServiceClient(rubricConn)

		// Контейнер для оценок
		var gradeRows []fyne.CanvasObject
		var grades []float32
		for _, work := range worksResp.Works {
			if work.Status == "graded by assistant" || work.Status == "graded by seminarist" {
				// Получение оценок
				marksResp, err := gradingClient.GetCriteriaMarks(context.Background(), &gradingpb.GetCriteriaMarksRequest{WorkId: work.Id})
				if err != nil || marksResp.Error != "" {
					log.Printf("Ошибка получения оценок для работы %d: %v", work.Id, err)
					continue
				}

				// Вычисление итоговой оценки
				blockingResp, err := rubricClient.LoadTaskBlockingCriterias(context.Background(), &rubricpb.LoadTaskBlockingCriteriasRequest{TaskId: work.Id})
				if err != nil || blockingResp.Error != "" {
					log.Printf("Ошибка загрузки блокирующих критериев для работы %d: %v", work.Id, err)
					continue
				}

				mainResp, err := rubricClient.LoadTaskMainCriterias(context.Background(), &rubricpb.LoadTaskMainCriteriasRequest{TaskId: work.Id})
				if err != nil || mainResp.Error != "" {
					log.Printf("Ошибка загрузки основных критериев для работы %d: %v", work.Id, err)
					continue
				}

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

				var finalGrade float32
				if hasBlockingMark {
					finalGrade = minBlockingMark
				} else {
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
					for _, group := range mainResp.Groups {
						totalMaxMark += float32(len(group.Criteria))
					}
					if totalMaxMark > 0 {
						finalGrade = (totalMark / totalMaxMark) * 10
					} else {
						finalGrade = 0
					}
				}

				gradeLabel := widget.NewLabel(fmt.Sprintf("%.2f", finalGrade))
				gradeRows = append(gradeRows, gradeLabel)
				grades = append(grades, finalGrade)
			}
		}

		// Вычисление средней оценки
		var averageGrade float32
		if len(grades) > 0 {
			sum := float32(0)
			for _, grade := range grades {
				sum += grade
			}
			averageGrade = sum / float32(len(grades))
		}

		// Создание строки таблицы
		discLabel := widget.NewLabel(discipline.Name)
		gradesContainer := container.NewHBox(gradeRows...)
		avgLabel := widget.NewLabel(fmt.Sprintf("%.2f", averageGrade))

		row := container.NewHBox(discLabel, gradesContainer, avgLabel)
		tableContent = append(tableContent, row)

		// Добавление разделителя
		if len(discResp.Disciplines) > 1 && &discipline != &discResp.Disciplines[len(discResp.Disciplines)-1] {
			separator := canvas.NewLine(color.Black)
			separator.StrokeWidth = 2
			separator.Position1 = fyne.NewPos(0, 0)
			separator.Position2 = fyne.NewPos(1600, 0)
			separatorContainer := container.New(layout.NewMaxLayout(), separator)
			tableContent = append(tableContent, separatorContainer)
		}
	}

	mainContent := container.NewVBox(tableContent...)

	worksButton := widget.NewButton("Список работ", func() {
		log.Println("Переход на страницу списка работ")
		state.currentPage = "student_works"
		myWindow.SetContent(createContent(state))
	})
	worksButtonContainer := container.New(layout.NewMaxLayout(), worksButton)
	worksButtonPlacement := container.New(layout.NewHBoxLayout(), layout.NewSpacer(), worksButtonContainer)

	centralContentPanel := container.NewVBox(backButtonRow, mainContent, worksButtonPlacement)

	contentBackground := canvas.NewRectangle(color.White)
	centralContentWithBackground := container.NewStack(contentBackground, centralContentPanel)

	return container.NewBorder(headerWithBackground, nil, nil, nil, centralContentWithBackground)
}

func СreateStudentWorksPage(state *AppState) fyne.CanvasObject {
	myWindow := state.window

	// Цвета
	headerTextColor := color.White
	darkBlue := color.NRGBA{R: 20, G: 40, B: 80, A: 255}
	separatorColor := color.NRGBA{R: 200, G: 200, B: 200, A: 255}

	// Верхняя панель (Header)
	logo := canvas.NewText("ВШЭ", headerTextColor)
	logo.TextStyle.Bold = true
	logoContainer := container.New(layout.NewMaxLayout(), logo)

	headerTitleText := canvas.NewText("Список работ", headerTextColor)
	headerTitleText.TextStyle.Bold = true
	headerTitleText.Alignment = fyne.TextAlignCenter

	header := container.New(layout.NewBorderLayout(nil, nil, logoContainer, nil),
		logoContainer,
		container.NewCenter(headerTitleText),
	)
	headerBackground := canvas.NewRectangle(darkBlue)
	headerWithBackground := container.NewStack(headerBackground, header)

	// Кнопка "назад"
	backButton := widget.NewButton("назад", func() {
		log.Println("Переход на страницу списка работ")
		state.currentPage = "student_grades"
		myWindow.SetContent(createContent(state))
	})
	backButtonRow := container.New(layout.NewMaxLayout(), backButton)

	// Контейнер для таблицы
	var tableContent []fyne.CanvasObject

	// Подключение к WorkService (порт 50053) для получения списка работ
	connWork, err := grpc.Dial("89.169.39.161:50053", grpc.WithInsecure())
	if err != nil {
		log.Printf("Не удалось подключиться к WorkService: %v", err)
		return container.NewVBox(widget.NewLabel("Ошибка подключения к серверу работ"))
	}
	defer connWork.Close()

	workClient := pbWork.NewWorkServiceClient(connWork)
	userIDint64, err := strconv.ParseInt(state.userID, 10, 32)
	if err != nil {
		log.Printf("Invalid user ID: %v", err)
	}
	studentID32 := int32(userIDint64)
	respTask, err := workClient.ListTasksForStudent(context.Background(), &pbWork.ListTasksForStudentRequest{StudentId: studentID32})
	if err != nil {
		log.Printf("Не удалось получить список работ: %v", err)
		return container.NewVBox(widget.NewLabel("Ошибка загрузки работ"))
	}
	if respTask.Error != "" {
		log.Printf("Ошибка от сервера: %s", respTask.Error)
		return container.NewVBox(widget.NewLabel(fmt.Sprintf("Ошибка: %s", respTask.Error)))
	}

	log.Printf("%v", len(respTask.Tasks))
	for i, task := range respTask.Tasks {
		log.Printf("%v", task.Id)
		deadlineTime, err := time.Parse(time.RFC3339, task.Deadline)
		if err != nil {
			log.Printf("Ошибка парсинга дедлайна: %v", err)
			return container.NewVBox(widget.NewLabel("Ошибка обработки дедлайна"))
		}
		deadlineLabel := widget.NewLabel(deadlineTime.Format("02.01.2006 15:04"))
		titleLabel := widget.NewLabel(task.Title)
		statusLabel := widget.NewLabel(task.Status)
		detailsButton := widget.NewButton("Подробнее", func(w *pbWork.Tasks) func() {
			return func() {
				TaskID = task.Id
				prevPage = "student_works"
				state.currentPage = "student_assignment"
				myWindow.SetContent(createContent(state))
			}
		}(task))

		row := container.NewHBox(deadlineLabel, titleLabel, statusLabel, detailsButton)
		tableContent = append(tableContent, row)

		if i < len(respTask.Tasks)-1 {
			separator := canvas.NewLine(separatorColor)
			separator.StrokeWidth = 2
			separator.Position1 = fyne.NewPos(0, 0)
			separator.Position2 = fyne.NewPos(1920, 0)
			separatorContainer := container.New(layout.NewMaxLayout(), separator)
			tableContent = append(tableContent, separatorContainer)
		}
	}

	mainContent := container.NewVBox(tableContent...)

	centralContentPanel := container.NewVBox(backButtonRow, mainContent)

	contentBackground := canvas.NewRectangle(color.White)
	centralContentWithBackground := container.NewStack(contentBackground, centralContentPanel)

	return container.NewBorder(headerWithBackground, nil, nil, nil, centralContentWithBackground)
}

func CreateStudentWorkDetailsPage(state *AppState) fyne.CanvasObject {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.Dial("89.169.39.161:50054", grpc.WithInsecure())
	if err != nil {
		log.Printf("Не удалось подключиться к сервису: %v", err)
		return container.NewVBox(widget.NewLabel("Ошибка подключения к сервису"))
	}
	defer conn.Close()

	client := workassignmentpb.NewWorkAssignmentServiceClient(conn)
	resp, err := client.GetTaskDetails(ctx, &workassignmentpb.GetTaskDetailsRequest{TaskId: TaskID})
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

	downloadButton := widget.NewButton("Загрузить работу", func() {
		w := state.window
		fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil || reader == nil {
				log.Printf("Ошибка при выборе файла: %v", err)
				dialog.ShowError(fmt.Errorf("Не удалось открыть файл: %v", err), w)
				return
			}
			defer reader.Close()

			fileName := reader.URI().Name()

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			conn, err := grpc.Dial("89.169.39.161:50054", grpc.WithInsecure())
			if err != nil {
				log.Printf("Не удалось подключиться к сервису: %v", err)
				dialog.ShowError(fmt.Errorf("Ошибка подключения к сервису: %v", err), w)
				return
			}
			defer conn.Close()

			client := workassignmentpb.NewWorkAssignmentServiceClient(conn)

			// Создание работы
			userIDint64, err := strconv.ParseInt(state.userID, 10, 32)
			if err != nil {
				log.Printf("Invalid user ID: %v", err)
				dialog.ShowError(fmt.Errorf("Неверный ID пользователя: %v", err), w)
				return
			}
			studentID32 := int32(userIDint64)
			createResp, err := client.CreateWork(ctx, &workassignmentpb.CreateWorkRequest{
				StudentId: studentID32,
				TaskId:    TaskID,
			})
			if err != nil || createResp.Error != "" {
				log.Printf("Ошибка создания работы: %v, %s", err, createResp.Error)
				dialog.ShowError(fmt.Errorf("Не удалось создать работу: %v", err), w)
				return
			}
			workID := createResp.WorkId

			// Запрос pre-signed URL
			urlResp, err := client.GenerateUploadURL(ctx, &workassignmentpb.GenerateUploadURLRequest{
				WorkId:   workID,
				FileName: fileName,
			})
			if err != nil || urlResp.Error != "" {
				log.Printf("Ошибка получения URL: %v, %s", err, urlResp.Error)
				dialog.ShowError(fmt.Errorf("Не удалось получить URL: %v", err), w)
				return
			}

			// Загрузка файла в S3 через pre-signed URL
			httpClient := &http.Client{}
			req, err := http.NewRequestWithContext(ctx, "PUT", urlResp.Url, reader)
			if err != nil {
				log.Printf("Ошибка создания HTTP-запроса: %v", err)
				dialog.ShowError(fmt.Errorf("Ошибка загрузки файла: %v", err), w)
				return
			}
			resp, err := httpClient.Do(req)
			if err != nil || resp.StatusCode != http.StatusOK {
				log.Printf("Ошибка отправки файла в S3: %v, статус: %d", err, resp.StatusCode)
				dialog.ShowError(fmt.Errorf("Не удалось загрузить файл в S3"), w)
				return
			}
			defer resp.Body.Close()

			log.Printf("Файл %s успешно загружен для работы %d", fileName, workID)
			dialog.ShowInformation("Успех", fmt.Sprintf("Файл %s успешно загружен", fileName), w)
		}, w)
		fileDialog.Show()
	})
	// Кнопка для просмотра файла
	viewButton := widget.NewButton("Просмотреть работу", func() {
		w := state.window
		if WorkID == 0 {
			dialog.ShowError(fmt.Errorf("Работа не создана"), w)
			return
		}

		urlResp, err := client.GetAssignmentFileURL(ctx, &workassignmentpb.GetAssignmentFileURLRequest{
			WorkId:   WorkID,
			FileName: "work_file", // Замените на динамическое имя файла, если нужно
		})
		if err != nil {
			log.Printf("Ошибка получения URL для work_id %d: %v", WorkID, err)
			dialog.ShowError(fmt.Errorf("Ошибка получения ссылки: %v", err), w)
			return
		}
		if urlResp.Error != "" {
			log.Printf("Ошибка от сервера при получении URL: %s", urlResp.Error)
			dialog.ShowError(fmt.Errorf("Ошибка сервера: %s", urlResp.Error), w)
			return
		}

		// Открытие ссылки в браузере
		parsedURL, err := url.Parse(urlResp.Url)
		if err != nil {
			log.Printf("Ошибка парсинга URL для work_id %d: %v", WorkID, err)
			dialog.ShowError(fmt.Errorf("Неверный формат ссылки: %v", err), w)
			return
		}

		// Открытие ссылки в браузере
		if err := fyne.CurrentApp().OpenURL(parsedURL); err != nil {
			log.Printf("Ошибка открытия URL для work_id %d: %v", WorkID, err)
			dialog.ShowError(fmt.Errorf("Не удалось открыть файл: %v", err), w)
		}
	})

	backButton := widget.NewButton("Назад", func() {
		state.currentPage = "student_works"
		state.window.SetContent(createContent(state))
	})

	nextButton := widget.NewButton("Критерии", func() {
		state.currentPage = "student_block_criteria"
		state.window.SetContent(createContent(state))
	})

	buttonsContainer := container.NewHBox(backButton, layout.NewSpacer(), downloadButton, viewButton, nextButton)

	inputGrid := container.NewVBox(
		titleLabel,
		scrollableDescription,
		deadlineLabel,
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

type blockingCriterionEntry struct {
	CriterionID     int32
	CommentEntry    *widget.Label
	EvaluationEntry *widget.Label
	Container       *fyne.Container
}

func CreateStudentBlockingCriteriaPage(state *AppState) fyne.CanvasObject {
	w := state.window

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rubricConn, err := grpc.Dial("89.169.39.161:50055", grpc.WithInsecure())
	if err != nil {
		log.Printf("Failed to connect to rubricservice: %v", err)
		return container.NewVBox(widget.NewLabel("Ошибка подключения к сервису"))
	}
	defer rubricConn.Close()
	rubricClient := rubricpb.NewRubricServiceClient(rubricConn)

	marksMap := make(map[int32]gradingpb.CriterionMark)

	if WorkID != 0 {
		gradingConn, err := grpc.Dial("89.169.39.161:50057", grpc.WithInsecure())
		if err != nil {
			log.Printf("Failed to connect to gradingservice: %v", err)
			return container.NewVBox(widget.NewLabel("Ошибка подключения к сервису"))
		}
		defer gradingConn.Close()
		gradingClient := gradingpb.NewGradingServiceClient(gradingConn)

		//Загрузка существующих оценок
		marksResp, err := gradingClient.GetCriteriaMarks(ctx, &gradingpb.GetCriteriaMarksRequest{WorkId: WorkID})
		if err != nil {
			log.Printf("Не удалось загрузить оценки для работы %d: %v", WorkID, err)
			return container.NewVBox(widget.NewLabel("Ошибка загрузки оценок: " + err.Error()))
		}
		if marksResp.Error != "" {
			log.Printf("Ошибка загрузки оценок для работы %d: %s", WorkID, marksResp.Error)
			return container.NewVBox(widget.NewLabel("Ошибка загрузки оценок: " + marksResp.Error))
		}

		// Создаем карту для быстрого доступа к оценкам по criterion_id
		for _, mark := range marksResp.Marks {
			marksMap[mark.CriterionId] = *mark
		}
	}

	// Загрузка блокирующих критериев
	resp, err := rubricClient.LoadTaskBlockingCriterias(ctx, &rubricpb.LoadTaskBlockingCriteriasRequest{TaskId: TaskID})
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
		state.currentPage = "student_assignment"
		w.SetContent(createContent(state))
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
	var activeCriteria []*blockingCriterionEntry
	for _, crit := range resp.Criteria {
		// Поля только для чтения
		nameEntry := widget.NewLabel(crit.Name)
		nameEntryContainer := container.NewMax(nameEntry)
		nameEntryContainer.Resize(fyne.NewSize(250, 60))

		descriptionEntry := widget.NewLabel(crit.Description)
		descriptionEntryContainer := container.NewMax(descriptionEntry)
		descriptionEntryContainer.Resize(fyne.NewSize(250, 60))

		textCommentEntry := ""
		textEvaluationEntry := ""
		if WorkID != 0 {
			//Загружаем существующие данные
			if mark, exists := marksMap[crit.Id]; exists {
				textCommentEntry = mark.Comment
				textEvaluationEntry = fmt.Sprintf("%.2f", mark.Mark)
				log.Printf("Загружена оценка для блокирующего критерия ID %d: mark=%.2f, comment=%s", crit.Id, mark.Mark, mark.Comment)
			} else {
				textCommentEntry = crit.Comment
				textEvaluationEntry = strconv.FormatInt(crit.FinalMark, 10)
				log.Printf("Оценка не найдена для блокирующего критерия ID %d, установлена оценка по умолчанию 0.0 и комментарий лектора: %s", crit.Id, crit.Comment)
			}

		}
		commentEntry := widget.NewLabel(textCommentEntry)
		evaluationEntry := widget.NewLabel(textEvaluationEntry)

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

		activeCriteria = append(activeCriteria, &blockingCriterionEntry{
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
		log.Printf("All blocking criteria marks saved successfully")
		state.currentPage = "student_main_criteria"
		w.SetContent(createContent(state))
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

type mainCriterionEntry struct {
	CriterionID  int32
	Select       *widget.Label
	CommentEntry *widget.Label
}

func CreateStudentMainCriteriaPage(state *AppState) fyne.CanvasObject {
	w := state.window

	// Подключение к gRPC сервисам
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	rubricConn, err := grpc.Dial("89.169.39.161:50055", grpc.WithInsecure())
	if err != nil {
		log.Printf("Не удалось подключиться к RubricService: %v", err)
		return container.NewVBox(widget.NewLabel("Ошибка подключения к сервису"))
	}
	defer rubricConn.Close()
	rubricClient := rubricpb.NewRubricServiceClient(rubricConn)

	marksMap := make(map[int32]gradingpb.CriterionMark)
	if WorkID != 0 {
		gradingConn, err := grpc.Dial("89.169.39.161:50057", grpc.WithInsecure())
		if err != nil {
			log.Printf("Не удалось подключиться к GradingService: %v", err)
			return container.NewVBox(widget.NewLabel("Ошибка подключения к сервису"))
		}
		defer gradingConn.Close()
		gradingClient := gradingpb.NewGradingServiceClient(gradingConn)

		// Загрузка существующих оценок
		marksResp, err := gradingClient.GetCriteriaMarks(ctx, &gradingpb.GetCriteriaMarksRequest{WorkId: WorkID})
		if err != nil {
			log.Printf("Не удалось загрузить оценки для работы %d: %v", WorkID, err)
			return container.NewVBox(widget.NewLabel("Ошибка загрузки оценок: " + err.Error()))
		}
		if marksResp.Error != "" {
			log.Printf("Ошибка загрузки оценок для работы %d: %s", WorkID, marksResp.Error)
			return container.NewVBox(widget.NewLabel("Ошибка загрузки оценок: " + marksResp.Error))
		}

		// Создаем карту для быстрого доступа к оценкам по criterion_id
		for _, mark := range marksResp.Marks {
			marksMap[mark.CriterionId] = *mark
		}
	}

	// Загрузка основных критериев
	resp, err := rubricClient.LoadTaskMainCriterias(ctx, &rubricpb.LoadTaskMainCriteriasRequest{TaskId: TaskID})
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
	var entries []mainCriterionEntry
	type criterionInfo struct {
		groupName string
		critName  string
	}
	criteriaInfo := make(map[int32]criterionInfo)
	totalCriteriaCount := 0

	// Инициализация entries для всех критериев
	for _, group := range resp.Groups {
		for _, crit := range group.Criteria {
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

			// Инициализация виджетов с значениями по умолчанию
			selectWidget := widget.NewLabel("0.0") // Значение по умолчанию
			commentEntry := widget.NewLabel("")    // Пустой комментарий по умолчанию
			if hasMark {
				selectWidget.SetText(currentMark)
				commentEntry.SetText(marksMap[crit.Id].Comment)
			}

			entries = append(entries, mainCriterionEntry{
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
			var selectedEntry *mainCriterionEntry
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

			// Проверка на nil и создание контейнера для отображения
			selectLabel := selectedEntry.Select
			if selectLabel == nil {
				selectLabel = widget.NewLabel("0.0") // Значение по умолчанию
			}

			commentLabel := selectedEntry.CommentEntry
			if commentLabel == nil {
				commentLabel = widget.NewLabel("") // Пустой комментарий по умолчанию
			}

			// Создаем контейнер для отображения комментариев
			commentsContainer := container.NewVBox(
				widget.NewLabelWithStyle("Комментарии лектора:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
				widget.NewLabel(fmt.Sprintf("Для 0.00: %s", crit.Comment_000)),
				widget.NewLabel(fmt.Sprintf("Для 0.25: %s", crit.Comment_025)),
				widget.NewLabel(fmt.Sprintf("Для 0.50: %s", crit.Comment_050)),
				widget.NewLabel(fmt.Sprintf("Для 0.75: %s", crit.Comment_075)),
				widget.NewLabel(fmt.Sprintf("Для 1.00: %s", crit.Comment_100)),
			)

			content := container.NewVBox(
				widget.NewLabel("Критерий: "+crit.Name),
				commentsContainer,
				container.NewHBox(widget.NewLabel("Оценка:"), selectLabel),
				container.NewHBox(widget.NewLabel("Комментарий:"), commentLabel),
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
		state.currentPage = "student_block_criteria"
		w.SetContent(createContent(state))
	})

	finalizeButton := widget.NewButton("Вернуться к описанию", func() {
		// Логика отправки работы (временно закомментирована)
		state.currentPage = "student_assignment"
		w.SetContent(createContent(state))
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
