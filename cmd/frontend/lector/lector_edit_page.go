package lector // Объявление пакета `pages`. Этот пакет, вероятно, содержит функции для создания различных страниц или представлений в вашем приложении Fyne.

import (
	"fmt"         // Импорт пакета `fmt` для форматированного ввода/вывода (например, для печати в консоль).
	"image/color" // Импорт пакета `image/color` для определения цветов (например, для фона или текста).
	"strconv"     // Импорт пакета `strconv` для преобразования строк в числа (например, для часов и минут).
	"time"        // Импорт пакета `time` для работы с датами и временем.

	"fyne.io/fyne/v2"           // Импорт основного пакета `fyne/v2`, предоставляющего базовые типы и функции Fyne (например, fyne.NewSize).
	"fyne.io/fyne/v2/app"       // Импорт пакета `app` для создания и управления жизненным циклом приложения Fyne.
	"fyne.io/fyne/v2/canvas"    // Импорт пакета `canvas` для низкоуровневой отрисовки графических элементов (таких как текст, прямоугольники).
	"fyne.io/fyne/v2/container" // Импорт пакета `container` для компоновки и группировки виджетов (например, container.NewVBox, container.NewStack).
	"fyne.io/fyne/v2/dialog"    // Импорт пакета `dialog` для создания стандартных и кастомных диалоговых окон.
	"fyne.io/fyne/v2/layout"    // Импорт пакета `layout` для различных алгоритмов размещения виджетов внутри контейнеров (например, layout.NewHBoxLayout, layout.NewBorderLayout).
	"fyne.io/fyne/v2/widget"    // Импорт пакета `widget` для стандартных UI-компонентов (кнопки, поля ввода, метки, календарь и т.д.).
)

// CreateWorkPage является основной функцией этой страницы.
// Она инициализирует Fyne-приложение, создает главное окно и компонует все UI-элементы для страницы "Создать работу".
func CreateWorkPage() {
	a := app.New()                     // Инициализирует новое приложение Fyne. Это первый шаг при создании Fyne-приложения.
	w := a.NewWindow("Создать работу") // Создает новое окно приложения с указанным заголовком.
	w.Resize(fyne.NewSize(700, 900))   // Устанавливает начальный размер окна в 700 пикселей по ширине и 900 по высоте.

	// --- Компоненты Заголовка и Боковой Панели ---
	headerTextColor := color.White // Определяем цвет текста для заголовков.

	// Создание текстового элемента для логотипа "ВШЭ" в верхнем левом углу.
	logoText := canvas.NewText("ВШЭ", headerTextColor)
	logoText.TextStyle.Bold = true                   // Делает текст жирным.
	logoText.TextSize = 24                           // Устанавливает размер шрифта.
	logoText.Alignment = fyne.TextAlignCenter        // Выравнивает текст по центру в его области.
	leftHeaderObject := container.NewStack(logoText) // Помещает текст в контейнер Stack, который позволяет элементам накладываться (хотя здесь он один).

	// Создание текстового элемента для основного заголовка страницы "Создать работу".
	headerTitle := canvas.NewText("Создать работу", headerTextColor)
	headerTitle.TextStyle.Bold = true            // Делает текст жирным.
	headerTitle.TextSize = 20                    // Устанавливает размер шрифта.
	headerTitle.Alignment = fyne.TextAlignCenter // Выравнивает текст по центру.

	// Компоновка содержимого заголовка окна с использованием `layout.NewBorderLayout`.
	// Этот макет размещает элементы по краям (верх, низ, лево, право) и в центре.
	headerContent := container.New(layout.NewBorderLayout(nil, nil, leftHeaderObject, nil),
		leftHeaderObject,                 // `leftHeaderObject` (логотип "ВШЭ") размещается слева.
		container.NewCenter(headerTitle), // `headerTitle` центрируется в оставшемся пространстве.
	)

	// Создание контейнера для кнопок боковой панели. В данном случае он пуст,
	// но `layout.NewSpacer()` гарантирует, что если кнопки появятся, они будут прижаты к нижней части.
	sideBarButtons := container.NewVBox(
		layout.NewSpacer(),
	)

	// --- Центральный Контент: Поля Ввода ---

	// Поле ввода для названия работы.
	titleEntry := widget.NewEntry()
	titleEntry.SetPlaceHolder("Название") // Текст-заполнитель, видимый, когда поле пустое.

	// Многострочное поле ввода для описания работы.
	descriptionEntry := widget.NewMultiLineEntry()
	descriptionEntry.SetPlaceHolder("Описание работы")
	// Мы не устанавливаем `descriptionEntry.SetMinRowsVisible()`,
	// так как `container.NewVScroll` сам обеспечит прокрутку, если текст не помещается.

	// Оборачиваем многострочное поле описания в контейнер с вертикальной прокруткой.
	scrollableDescription := container.NewVScroll(descriptionEntry)
	// Устанавливаем минимальную высоту для области прокрутки.
	// `descriptionEntry.MinSize().Height * 5` означает, что изначально будет видно как минимум 5 строк текста,
	// но пользователь сможет прокручивать дальше, если текста будет больше.
	// Ширина `0` означает, что контейнер будет занимать всю доступную ширину по горизонтали.
	scrollableDescription.SetMinSize(fyne.NewSize(0, descriptionEntry.MinSize().Height*5))

	// Поле ввода для отображения выбранной даты и времени дедлайна.
	dateAndTimeEntry := widget.NewEntry()
	dateAndTimeEntry.SetPlaceHolder("Выберите дату и время дедлайна")
	dateAndTimeEntry.Disable() // Делает поле нередактируемым, так как оно заполняется из диалога выбора даты/времени.

	var selectedDateTime time.Time // Переменная для хранения выбранной даты и времени.
	var isDateTimeSelected bool    // Флаг, указывающий, была ли дата и время выбраны пользователем.

	// Кнопка для вызова диалога выбора даты и времени.
	selectDateTimeButton := widget.NewButton("Выбрать", func() {
		// При нажатии на кнопку вызывается функция `showDateTimePickerDialog`, которая открывает диалог.
		showDateTimePickerDialog(w, &selectedDateTime, &isDateTimeSelected, dateAndTimeEntry)
	})

	// Горизонтальный контейнер для поля ввода дедлайна и кнопки "Выбрать".
	// В данном случае `dateAndTimeEntry` не оборачивается в `NewExpandLayout`,
	// что означает, что он будет занимать свой минимальный размер, а остальное место достанется кнопке.
	// Если бы нужно было, чтобы `dateAndTimeEntry` растягивался, его следовало бы обернуть в `container.New(layout.NewExpandLayout(), ...)`.
	dateTimeInputContainer := container.New(layout.NewHBoxLayout(),
		dateAndTimeEntry,
		selectDateTimeButton,
	)

	// Основная сетка для полей ввода. Используется `layout.NewVBoxLayout`, что означает вертикальную компоновку.
	// Это значит, что все элементы внутри будут располагаться друг под другом.
	inputGrid := container.New(layout.NewVBoxLayout(),
		// Горизонтальный контейнер для `titleEntry` и `dateTimeInputContainer`.
		// `layout.NewGridLayoutWithColumns(1)` вокруг `titleEntry` здесь не дает эффекта растягивания,
		// так как это просто один элемент в одной колонке.
		// `layout.NewSpacer()` будет отталкивать `titleEntry` от `dateTimeInputContainer`.
		container.New(layout.NewHBoxLayout(),
			container.New(layout.NewGridLayoutWithColumns(1), titleEntry), // titleEntry здесь в колонке 1x1.
			layout.NewSpacer(),                                            // Растягивающийся пробел между полями.
			dateTimeInputContainer,                                        // Контейнер с полем дедлайна и кнопкой.
		),
		scrollableDescription, // Контейнер с прокручиваемым описанием работы.
	)

	// Кнопка "Далее" для обработки введенных данных.
	nextButton := widget.NewButton("Далее", func() {
		fmt.Printf("Название: %s\n", titleEntry.Text)       // Выводим название в консоль.
		fmt.Printf("Описание: %s\n", descriptionEntry.Text) // Выводим описание в консоль.

		if isDateTimeSelected { // Проверяем, была ли выбрана дата/время.
			fmt.Printf("Полный дедлайн: %s\n", selectedDateTime.Format("02.01.2006 15:04")) // Выводим дедлайн.
		} else {
			fmt.Println("Полный дедлайн: Дата и время не выбраны") // Сообщение, если дедлайн не выбран.
		}
		// Здесь обычно добавляется логика для обработки данных (например, сохранение) и переход к следующему шагу.
	})
	// Контейнер для кнопки "Далее", чтобы она была прижата к правому краю.
	nextButtonContainer := container.New(layout.NewHBoxLayout(), layout.NewSpacer(), nextButton)

	// Фон для центрального контента.
	contentBackground := canvas.NewRectangle(color.White)
	// Добавляем отступы вокруг основной сетки ввода.
	contentWithPadding := container.NewPadded(inputGrid)
	// Собираем фон и контент с отступами в один Stack контейнер (позволяет накладывать элементы).
	centralContent := container.NewStack(contentBackground, contentWithPadding)

	// Устанавливаем основной контент окна. Используется `layout.NewBorderLayout` для общей структуры окна.
	// Stack используется для фона окна.
	w.SetContent(container.NewStack(
		canvas.NewRectangle(color.NRGBA{R: 20, G: 40, B: 80, A: 255}), // Основной темно-синий фон всего окна.
		container.NewBorder( // Главный Border-макет окна.
			headerContent,       // Размещает заголовок в верхней части.
			nextButtonContainer, // Размещает кнопку "Далее" в нижней части.
			sideBarButtons,      // Размещает боковую панель слева.
			nil,                 // Правая часть окна пуста.
			centralContent,      // Центральная часть окна.
		),
	))

	w.ShowAndRun() // Отображает окно и запускает главный цикл событий Fyne. Это блокирующая функция.
}

// showDateTimePickerDialog отображает диалоговое окно для выбора даты и времени.
// parent: родительское окно, к которому привязан диалог.
// selectedTime: указатель на переменную `time.Time`, в которую будет сохранено выбранное время.
// isSelected: указатель на булеву переменную, которая будет истинной, если дата/время выбраны.
// displayEntry: поле ввода на главной странице, в которое будет отображен выбранный дедлайн.
func showDateTimePickerDialog(parent fyne.Window, selectedTime *time.Time, isSelected *bool, displayEntry *widget.Entry) {
	var initialDate time.Time // Переменная для хранения начальной даты, которая будет показана в календаре.
	var currentHour string    // Переменная для хранения выбранного часа в виде строки (например, "09").
	var currentMinute string  // Переменная для хранения выбранных минут в виде строки (например, "05").

	// Если дата и время уже были выбраны ранее, используем их как начальные значения для диалога.
	if *isSelected {
		initialDate = *selectedTime
	} else {
		initialDate = time.Now() // Иначе используем текущее системное время.
	}
	currentHour = fmt.Sprintf("%02d", initialDate.Hour())     // Форматируем час в двухзначный формат.
	currentMinute = fmt.Sprintf("%02d", initialDate.Minute()) // Форматируем минуты в двухзначный формат.

	dateFromCalendar := initialDate // Переменная для хранения даты, выбранной из виджета календаря.

	// Создаем виджет календаря с начальной датой и функцией обратного вызова при изменении даты.
	calendar := widget.NewCalendar(initialDate, func(t time.Time) {
		dateFromCalendar = t // Обновляем выбранную дату при взаимодействии с календарем.
	})

	// Генерируем список часов (от "00" до "23") для выпадающего списка.
	hours := make([]string, 24)
	for i := 0; i < 24; i++ {
		hours[i] = fmt.Sprintf("%02d", i)
	}

	// Генерируем список минут (от "00" до "59") для выпадающего списка.
	minutes := make([]string, 60)
	for i := 0; i < 60; i++ {
		minutes[i] = fmt.Sprintf("%02d", i)
	}

	var hourSelect *widget.Select   // Объявление переменной для выпадающего списка часов.
	var minuteSelect *widget.Select // Объявление переменной для выпадающего списка минут.

	// Создаем выпадающий список для выбора часа.
	hourSelect = widget.NewSelect(hours, func(s string) {
		currentHour = s // Обновляем выбранный час.
	})
	hourSelect.SetSelected(currentHour) // Устанавливаем начальное выбранное значение.
	// SetMinSize удален. Размер будет по умолчанию или определен контейнером.

	// Создаем выпадающий список для выбора минут.
	minuteSelect = widget.NewSelect(minutes, func(s string) {
		currentMinute = s // Обновляем выбранные минуты.
	})
	minuteSelect.SetSelected(currentMinute) // Устанавливаем начальное выбранное значение.
	// SetMinSize удален. Размер будет по умолчанию или определен контейнером.

	// Кнопка "Now" для установки текущего времени.
	nowButton := widget.NewButton("Now", func() {
		currentTime := time.Now()
		currentHour = fmt.Sprintf("%02d", currentTime.Hour())
		currentMinute = fmt.Sprintf("%02d", currentTime.Minute())
		if hourSelect != nil {
			hourSelect.SetSelected(currentHour) // Обновляем выбранный час в UI.
		}
		if minuteSelect != nil {
			minuteSelect.SetSelected(currentMinute) // Обновляем выбранные минуты в UI.
		}
	})

	// Горизонтальный контейнер для элементов выбора времени.
	// NewExpandLayout удален. Элементы не будут растягиваться.
	timeLayout := container.New(layout.NewHBoxLayout(),
		widget.NewLabel("Time"), // Метка "Time".
		hourSelect,              // Выпадающий список часов.
		widget.NewLabel(":"),    // Разделитель.
		minuteSelect,            // Выпадающий список минут.
		layout.NewSpacer(),      // Растягивающийся пробел, прижимает `nowButton` вправо.
		nowButton,               // Кнопка "Now".
	)
	// Добавляем отступы вокруг контейнера выбора времени.
	timeLayoutWithPadding := container.NewPadded(timeLayout)

	// Собираем все компоненты диалогового окна выбора даты и времени вертикально.
	dialogContent := container.NewVBox(
		widget.NewLabelWithStyle("Choose date and time", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}), // Заголовок диалога.
		widget.NewSeparator(),                                                                              // Горизонтальный разделитель.
		calendar,                                                                                           // Календарь.
		widget.NewSeparator(),                                                                              // Еще один разделитель.
		timeLayoutWithPadding,                                                                              // Контейнер с выбором времени.
	)

	// Создаем кастомное диалоговое окно подтверждения.
	d := dialog.NewCustomConfirm(
		"",            // Пустой заголовок (так как есть свой заголовок внутри dialogContent).
		"Ok",          // Текст для кнопки подтверждения.
		"Cancel",      // Текст для кнопки отмены.
		dialogContent, // Содержимое диалогового окна.
		func(ok bool) { // Функция обратного вызова при закрытии диалога.
			if ok {     // Если пользователь нажал "Ok".
				h, _ := strconv.Atoi(currentHour)   // Преобразуем час из строки в число.
				m, _ := strconv.Atoi(currentMinute) // Преобразуем минуты из строки в число.

				// Создаем окончательный объект `time.Time` из выбранной даты и времени.
				finalTime := time.Date(
					dateFromCalendar.Year(),
					dateFromCalendar.Month(),
					dateFromCalendar.Day(),
					h,
					m,
					0, 0, // Секунды и наносекунды устанавливаются в 0.
					dateFromCalendar.Location(), // Используем временную зону выбранной даты.
				)

				*selectedTime = finalTime                                  // Сохраняем окончательное время.
				*isSelected = true                                         // Устанавливаем флаг, что дата/время выбраны.
				displayEntry.SetText(finalTime.Format("02.01.2006 15:04")) // Обновляем текст в поле на главной странице.
			} else { // Если пользователь нажал "Cancel" или закрыл диалог.
				fmt.Println("Выбор даты и времени отменен") // Сообщение об отмене.
			}
		},
		parent, // Родительское окно.
	)

	d.Resize(fyne.NewSize(400, 500)) // Устанавливаем размер диалогового окна.
	d.Show()                         // Отображаем диалоговое окно.
}
