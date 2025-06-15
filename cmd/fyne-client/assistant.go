package main

import (
	"context"
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
	"time"

	"google.golang.org/grpc"
	pb "rubr/proto/workassignment"
)

type AssistantWorkItem struct {
	WorkID          int32
	TaskTitle       string
	StudentEmail    string
	StudentFullName string
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

	client := pb.NewWorkAssignmentServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// получение списка работ
	resp, err := client.GetWorksForAssistant(ctx, &pb.GetWorksForAssistantRequest{AssistantId: userID})
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
		state.currentPage = "assistant_work_details"
		state.window.SetContent(CreateAssistantWorkDetailsPage(state, workID, leftBackground))
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

func CreateAssistantWorkDetailsPage(state *AppState, workID int32, leftBackground *canvas.Image) fyne.CanvasObject {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.Dial("localhost:50054", grpc.WithInsecure())
	if err != nil {
		log.Printf("Не удалось подключиться к сервису: %v", err)
		return container.NewVBox(widget.NewLabel("Ошибка подключения к сервису"))
	}
	defer conn.Close()

	client := pb.NewWorkAssignmentServiceClient(conn)
	resp, err := client.GetWorkDetails(ctx, &pb.GetWorkDetailsRequest{WorkId: workID})
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

	// Кнопки
	downloadButton := widget.NewButton("Загрузить работу", func() {
		if resp.ContentUrl == "" {
			dialog.ShowInformation("Ошибка", "Ссылка на работу отсутствует", state.window)
			return
		}

		// Проверяем корректность URL
		parsedURL, err := url.Parse(resp.ContentUrl)
		if err != nil {
			log.Printf("Некорректная ссылка: %v", err)
			dialog.ShowError(err, state.window)
			return
		}

		// Создаём поле для отображения ссылки
		linkEntry := widget.NewEntry()
		linkEntry.SetText(parsedURL.String())
		linkEntry.Disable() // Только для чтения

		// Кнопка "Копировать"
		copyButton := widget.NewButton("Копировать", func() {
			state.window.Clipboard().SetContent(linkEntry.Text)
			dialog.ShowInformation("Успех", "Ссылка скопирована в буфер обмена", state.window)
		})

		// Сборка содержимого диалога
		dialogContent := container.NewVBox(
			widget.NewLabel("Ссылка на работу: "),
			linkEntry,
			container.NewHBox(copyButton),
		)

		dialog.ShowCustom("Ссылка на работу", "Закрыть", dialogContent, state.window)
	})

	gradeButton := widget.NewButton("Оценить", func() {
		dialog.ShowInformation("Оценка", "Окно оценки будет реализовано позже", state.window)
	})

	backButton := widget.NewButton("Назад", func() {
		state.currentPage = "assistant_works"
		state.window.SetContent(CreateAssistantWorksPage(state, leftBackground))
	})

	buttonsContainer := container.NewHBox(backButton, layout.NewSpacer(), downloadButton, gradeButton)

	// Сборка контента
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
