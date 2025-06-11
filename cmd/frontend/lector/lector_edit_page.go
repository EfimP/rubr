package lector

import (
	"fmt"
	"fyne.io/fyne/v2/theme"
	"image/color"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// CreateWorkPage создает и отображает страницу для создания новой работы.
// Функция инициализирует окно с полями ввода для названия, описания и дедлайна,
// а также кнопками для управления процессом.
func CreateWorkPage() {
	// Инициализация приложения и окна с заданными размерами
	a := app.New()
	w := a.NewWindow("Создать работу")
	w.Resize(fyne.NewSize(700, 900))

	// Установка светлой темы для приложения
	a.Settings().SetTheme(theme.LightTheme())

	// Определяем цвет текста заголовка (белый для контраста с фоном)
	headerTextColor := color.White

	// Создаем логотип "ВШЭ" с жирным шрифтом и центрированным выравниванием
	logoText := canvas.NewText("ВШЭ", headerTextColor)
	logoText.TextStyle.Bold = true
	logoText.TextSize = 24
	logoText.Alignment = fyne.TextAlignCenter
	leftHeaderObject := container.NewStack(logoText)

	// Создаем заголовок страницы с жирным шрифтом и центрированным выравниванием
	headerTitle := canvas.NewText("Создать работу", headerTextColor)
	headerTitle.TextStyle.Bold = true
	headerTitle.TextSize = 20
	headerTitle.Alignment = fyne.TextAlignCenter

	// Компонуем заголовок: логотип слева, текст заголовка по центру
	headerContent := container.New(layout.NewBorderLayout(nil, nil, leftHeaderObject, nil),
		leftHeaderObject,
		container.NewCenter(headerTitle),
	)

	// Создаем боковую панель с кнопками (пока пустая, с спейсером)
	sideBarButtons := container.NewVBox(
		layout.NewSpacer(),
	)

	// Создаем поле ввода для названия работы с плейсхолдером
	titleEntry := widget.NewEntry()
	titleEntry.SetPlaceHolder("Название")
	titleEntry.TextStyle.TabWidth = 23 // Устанавливаем ширину табуляции для удобства

	// Создаем многострочное поле ввода для описания работы
	descriptionEntry := widget.NewMultiLineEntry()
	descriptionEntry.SetPlaceHolder("Описание работы")

	// Оборачиваем описание в прокручиваемый контейнер с минимальной высотой
	scrollableDescription := container.NewVScroll(descriptionEntry)
	scrollableDescription.SetMinSize(fyne.NewSize(0, descriptionEntry.MinSize().Height*5))

	// Создаем поле для ввода даты и времени (только для отображения)
	dateAndTimeEntry := widget.NewEntry()
	dateAndTimeEntry.SetPlaceHolder("Выберите дату и время дедлайна")
	dateAndTimeEntry.Disable()

	// Переменные для хранения выбранной даты и времени
	var selectedDateTime time.Time
	var isDateTimeSelected bool

	// Кнопка для открытия диалога выбора даты и времени
	selectDateTimeButton := widget.NewButton("Выбрать", func() {
		showDateTimePickerDialog(w, &selectedDateTime, &isDateTimeSelected, dateAndTimeEntry)
	})

	// Компонуем поле даты и кнопку в горизонтальный контейнер
	dateTimeInputContainer := container.New(layout.NewHBoxLayout(),
		dateAndTimeEntry,
		selectDateTimeButton,
	)

	// Создаем сетку ввода с полями названия, даты и описания
	inputGrid := container.New(layout.NewVBoxLayout(),
		container.New(layout.NewHBoxLayout(),
			container.New(layout.NewGridLayoutWithColumns(1), titleEntry),
			layout.NewSpacer(),
			dateTimeInputContainer,
		),
		scrollableDescription,
	)

	// Кнопка "Далее" для обработки введенных данных
	nextButton := widget.NewButton("Далее", func() {
		fmt.Printf("Название: %s\n", titleEntry.Text)
		fmt.Printf("Описание: %s\n", descriptionEntry.Text)

		if isDateTimeSelected {
			fmt.Printf("Полный дедлайн: %s\n", selectedDateTime.Format("02.01.2006 15:04"))
		} else {
			fmt.Println("Полный дедлайн: Дата и время не выбраны")
		}
	})
	// Помещаем кнопку "Далее" справа с помощью спейсера
	nextButtonContainer := container.New(layout.NewHBoxLayout(), layout.NewSpacer(), nextButton)

	// Создаем белый фон для центральной области
	contentBackground := canvas.NewRectangle(color.White)
	contentWithPadding := container.NewPadded(inputGrid)
	centralContent := container.NewStack(contentBackground, contentWithPadding)

	// Компонуем всю страницу: фон, заголовок, боковая панель, центральный контент и кнопка
	w.SetContent(container.NewStack(
		canvas.NewRectangle(color.NRGBA{R: 20, G: 40, B: 80, A: 255}),
		container.NewBorder(
			headerContent,
			nextButtonContainer,
			sideBarButtons,
			nil,
			centralContent,
		),
	))

	// Запускаем окно приложения
	w.ShowAndRun()
}

// showDateTimePickerDialog отображает диалоговое окно для выбора даты и времени.
// Принимает родительское окно, указатели на выбранное время и флаг выбора,
// а также поле для отображения результата.
func showDateTimePickerDialog(parent fyne.Window, selectedTime *time.Time, isSelected *bool, displayEntry *widget.Entry) {
	// Определяем начальную дату (текущая или ранее выбранная)
	var initialDate time.Time
	var currentHour string
	var currentMinute string

	if *isSelected {
		initialDate = *selectedTime
	} else {
		initialDate = time.Now()
	}
	currentHour = fmt.Sprintf("%02d", initialDate.Hour())
	currentMinute = fmt.Sprintf("%02d", initialDate.Minute())

	// Переменная для хранения выбранной даты из календаря
	dateFromCalendar := initialDate

	// Создаем календарь с обработчиком выбора даты
	calendar := widget.NewCalendar(initialDate, func(t time.Time) {
		dateFromCalendar = t
	})

	// Генерируем списки часов и минут
	hours := make([]string, 24)
	for i := 0; i < 24; i++ {
		hours[i] = fmt.Sprintf("%02d", i)
	}

	minutes := make([]string, 60)
	for i := 0; i < 60; i++ {
		minutes[i] = fmt.Sprintf("%02d", i)
	}

	// Создаем выпадающие списки для выбора часов и минут
	var hourSelect *widget.Select
	var minuteSelect *widget.Select

	hourSelect = widget.NewSelect(hours, func(s string) {
		currentHour = s
	})
	hourSelect.SetSelected(currentHour)

	minuteSelect = widget.NewSelect(minutes, func(s string) {
		currentMinute = s
	})
	minuteSelect.SetSelected(currentMinute)

	// Кнопка "Now" для установки текущего времени
	nowButton := widget.NewButton("Now", func() {
		currentTime := time.Now()
		currentHour = fmt.Sprintf("%02d", currentTime.Hour())
		currentMinute = fmt.Sprintf("%02d", currentTime.Minute())
		if hourSelect != nil {
			hourSelect.SetSelected(currentHour)
		}
		if minuteSelect != nil {
			minuteSelect.SetSelected(currentMinute)
		}
	})

	// Компонуем элементы выбора времени в горизонтальный контейнер
	timeLayout := container.New(layout.NewHBoxLayout(),
		widget.NewLabel("Time"),
		hourSelect,
		widget.NewLabel(":"),
		minuteSelect,
		layout.NewSpacer(),
		nowButton,
	)
	timeLayoutWithPadding := container.NewPadded(timeLayout)

	// Компонуем содержимое диалога
	dialogContent := container.NewVBox(
		widget.NewLabelWithStyle("Choose date and time", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		calendar,
		widget.NewSeparator(),
		timeLayoutWithPadding,
	)

	// Создаем и настраиваем диалоговое окно
	d := dialog.NewCustomConfirm(
		"",
		"Ok",
		"Cancel",
		dialogContent,
		func(ok bool) {
			if ok {
				h, _ := strconv.Atoi(currentHour)
				m, _ := strconv.Atoi(currentMinute)

				// Формируем финальную дату и время
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
			} else {
				fmt.Println("Выбор даты и времени отменен")
			}
		},
		parent,
	)

	// Устанавливаем размер диалога и отображаем его
	d.Resize(fyne.NewSize(400, 500))
	d.Show()
}
