package student

import (
	"image/color"
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

// Структура для хранения информации о задании
type Assignment struct {
	Title       string
	Description string
	Deadline    time.Time
	Submission  time.Time
	FilePath    string
}

// Симулированные данные задания
var assignment = Assignment{
	Title:       "НАЗВАНИЕ РАБОТЫ",
	Description: "ОПИСАНИЕ\nОПИСАНИЕ\nОПИСАНИЕ\nОПИСАНИЕ\nОПИСАНИЕ\nОПИСАНИЕ\nОПИСАНИЕ\nОПИСАНИЕ\n",
	Deadline:    time.Date(2025, 6, 20, 0, 0, 0, 0, time.UTC),
	Submission:  time.Time{},
	FilePath:    "",
}

func AssignmentScreen() {
	// Инициализирует приложение и окно
	myApp := app.New()
	myWindow := myApp.NewWindow("Студент: Задание")
	myWindow.Resize(fyne.NewSize(1920, 1080))

	// Определяет цвета для интерфейса
	headerTextColor := color.White
	darkBlue := color.NRGBA{R: 20, G: 40, B: 80, A: 255}
	separatorColor := color.NRGBA{R: 200, G: 200, B: 200, A: 255}

	// Создаёт логотип и заголовок
	logo := canvas.NewText("ВШЭ", headerTextColor)
	logo.TextStyle.Bold = true
	logoContainer := container.New(layout.NewMaxLayout(), logo)

	headerTitleText := canvas.NewText("Задание", headerTextColor)
	headerTitleText.TextStyle.Bold = true
	headerTitleText.Alignment = fyne.TextAlignCenter

	header := container.New(layout.NewBorderLayout(nil, nil, logoContainer, nil),
		logoContainer,
		container.NewCenter(headerTitleText),
	)
	headerBackground := canvas.NewRectangle(darkBlue)
	headerWithBackground := container.NewStack(headerBackground, header)

	// Создаёт кнопку "назад"
	backButton := widget.NewButton("назад", func() {
		log.Println("Кнопка 'назад' нажата. Возврат на предыдущий экран.")
		myWindow.Close()
	})
	backButtonRow := container.New(layout.NewMaxLayout(), backButton)

	// Формирует содержимое задания
	titleLabel := widget.NewLabel(assignment.Title)
	descriptionLabel := widget.NewLabel(assignment.Description)
	deadlineLabel := widget.NewLabel(assignment.Deadline.Format("02.01.2006 15:04"))
	var submissionLabel *widget.Label
	var filePathLabel *widget.Label
	if !assignment.Submission.IsZero() {
		submissionLabel = widget.NewLabel(assignment.Submission.Format("02.01.2006 15:04"))
		filePathLabel = widget.NewLabel(assignment.FilePath)
	} else {
		submissionLabel = widget.NewLabel("Не сдано")
		filePathLabel = widget.NewLabel("Файл не прикреплён")
	}

	// Кнопка для выбора файла
	attachButton := widget.NewButton("Прикрепить файл", func() {
		fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil || reader == nil {
				log.Println("Ошибка при выборе файла:", err)
				return
			}
			assignment.FilePath = reader.URI().Path()
			filePathLabel.SetText(assignment.FilePath)
			log.Println("Файл прикреплён:", assignment.FilePath)
		}, myWindow)
		fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".txt", ".pdf", ".docx", ".zip", ".rar"}))
		fileDialog.Show()
	})

	// Кнопка для сдачи работы
	submitButton := widget.NewButton("Сдать работу", func() {
		if assignment.FilePath == "" {
			log.Println("Ошибка: прикрепите файл перед сдачей")
			return
		}
		if assignment.Submission.IsZero() {
			assignment.Submission = time.Now()
			submissionLabel.SetText(assignment.Submission.Format("02.01.2006 15:04"))
			log.Println("Работа сдана в", assignment.Submission.Format("02.01.2006 15:04"), "с файлом:", assignment.FilePath)
		} else {
			log.Println("Работа пересдана в", time.Now().Format("02.01.2006 15:04"), "с новым файлом:", assignment.FilePath)
		}
	})

	// Кнопка для просмотра критериев переход на новую страницу
	criteriaButton := widget.NewButton("Посмотреть критерии", func() {
	})

	// Организует содержимое в таблицу
	var content []fyne.CanvasObject
	content = append(content, container.NewHBox(widget.NewLabel("Название"), titleLabel))
	separator1 := canvas.NewLine(separatorColor)
	separator1.StrokeWidth = 2
	separator1.Position1 = fyne.NewPos(0, 0)
	separator1.Position2 = fyne.NewPos(1920, 0)
	content = append(content, container.New(layout.NewMaxLayout(), separator1))
	content = append(content, container.NewHBox(widget.NewLabel("Описание"), descriptionLabel))
	separator2 := canvas.NewLine(separatorColor)
	separator2.StrokeWidth = 2
	separator2.Position1 = fyne.NewPos(0, 0)
	separator2.Position2 = fyne.NewPos(1920, 0)
	content = append(content, container.New(layout.NewMaxLayout(), separator2))
	content = append(content, container.NewHBox(widget.NewLabel("Дедлайн"), deadlineLabel))
	separator3 := canvas.NewLine(separatorColor)
	separator3.StrokeWidth = 2
	separator3.Position1 = fyne.NewPos(0, 0)
	separator3.Position2 = fyne.NewPos(1920, 0)
	content = append(content, container.New(layout.NewMaxLayout(), separator3))
	content = append(content, container.NewHBox(widget.NewLabel("Время сдачи"), submissionLabel))
	separator4 := canvas.NewLine(separatorColor)
	separator4.StrokeWidth = 2
	separator4.Position1 = fyne.NewPos(0, 0)
	separator4.Position2 = fyne.NewPos(1920, 0)
	content = append(content, container.New(layout.NewMaxLayout(), separator4))
	content = append(content, container.NewHBox(widget.NewLabel("Прикреплённый файл"), filePathLabel))
	separator5 := canvas.NewLine(separatorColor)
	separator5.StrokeWidth = 2
	separator5.Position1 = fyne.NewPos(0, 0)
	separator5.Position2 = fyne.NewPos(1920, 0)
	content = append(content, container.New(layout.NewMaxLayout(), separator5))
	content = append(content, container.NewHBox(attachButton, submitButton, criteriaButton))

	// Собирает таблицу в вертикальный контейнер
	mainContent := container.NewVBox(content...)

	// Организует центральную панель
	centralContentPanel := container.NewVBox(
		backButtonRow,
		mainContent,
	)

	// Настраивает фон
	contentBackground := canvas.NewRectangle(color.White)
	centralContentWithBackground := container.NewStack(contentBackground, centralContentPanel)

	// Устанавливает общую компоновку окна
	myWindow.SetContent(container.NewBorder(
		headerWithBackground,
		nil,
		nil,
		nil,
		centralContentWithBackground,
	))

	// Запускает окно
	log.Println("Экран 'Задание' запущен.")
	myWindow.ShowAndRun()
}
