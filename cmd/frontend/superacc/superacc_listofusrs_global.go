package superacc

import (
	"fmt"
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	//"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// User представляет структуру данных для одного пользователя.
// ФИО и Email хранятся отдельно, но могут отображаться вместе.
type User struct {
	ID     int
	FIO    string // Имя и Фамилия
	Email  string // Электронная почта
	Group  string // Группа, в которой состоит пользователь
	Status string // Статус пользователя (асс, студ, лек, семи)
}

// Global data store - симулированная база данных всех пользователей.
// Используется для поиска и добавления новых пользователей в список.
var allUsers = []*User{
	{ID: 1, FIO: "Иванов Иван", Email: "ivanov@mail.ru", Group: "Математики", Status: "студ"},
	{ID: 2, FIO: "Петров Петр", Email: "petrov@mail.ru", Group: "Физики", Status: "лек"},
	{ID: 3, FIO: "Сидорова Анна", Email: "sidorova@mail.ru", Group: "Математики", Status: "асс"},
	{ID: 4, FIO: "Кузнецов Дмитрий", Email: "kuznetsov@mail.ru", Group: "Химики", Status: "семи"},
	{ID: 5, FIO: "Васильева Елена", Email: "vasilieva@mail.ru", Group: "Физики", Status: "студ"},
	{ID: 6, FIO: "Смирнов Артем", Email: "smirnov@mail.ru", Group: "Биологи", Status: "студ"},
	{ID: 7, FIO: "Волкова Мария", Email: "volkova@mail.ru", Group: "Химики", Status: "асс"},
	{ID: 8, FIO: "Морозов Сергей", Email: "morozov@mail.ru", Group: "Математики", Status: "лек"},
	{ID: 9, FIO: "Новикова Ольга", Email: "novikova@mail.ru", Group: "Биологи", Status: "семи"},
	{ID: 10, FIO: "Федоров Алексей", Email: "fedorov@mail.ru", Group: "Физики", Status: "студ"},
}

// currentDisplayedUsers - это список пользователей, которые отображаются в таблице.
// Он может быть отфильтрован по поиску или изменен добавлением/удалением.
var currentDisplayedUsers []*User

// updateUsersTableUI объявляем здесь, чтобы она была доступна в замыканиях.
// Эта функция будет обновлять UI таблицы пользователей.
var updateUsersTableUI func(searchText string)

// ShowUsersListPage отображает страницу "Список пользователей".
func ShowUsersListPage() { // Изменено на ShowUsersListPage без groupName, т.к. это общий список
	a := app.New()
	a.Settings().SetTheme(theme.LightTheme()) // Устанавливаем постоянную белую тему

	w := a.NewWindow("Супер-акк: Список пользователей")
	w.Resize(fyne.NewSize(1200, 720))

	// Цвета
	headerTextColor := color.White
	darkBlue := color.NRGBA{R: 20, G: 40, B: 80, A: 255}
	lightGrayDivider := color.NRGBA{R: 220, G: 220, B: 220, A: 255}
	mediumGrayDivider := color.NRGBA{R: 180, G: 180, B: 180, A: 255}

	// --- Верхняя панель (Header) ---
	logo := canvas.NewText("ВШЭ", headerTextColor) // Логотип "Р"
	logo.TextStyle.Bold = true
	logo.TextSize = 28
	logoContainer := container.NewMax(logo)

	headerTitleText := canvas.NewText("Список пользователей", headerTextColor)
	headerTitleText.TextStyle.Bold = true
	headerTitleText.TextSize = 24
	headerTitleText.Alignment = fyne.TextAlignCenter // Заголовок по центру

	header := container.New(layout.NewBorderLayout(nil, nil, logoContainer, nil),
		container.NewPadded(logoContainer),
		container.NewCenter(headerTitleText),
	)
	headerBackground := canvas.NewRectangle(darkBlue)
	headerWithBackground := container.NewStack(headerBackground, header)

	// --- Левая боковая панель ---
	sidePanelBackground := canvas.NewRectangle(darkBlue)
	sidePanel := container.NewVBox()
	sidePanel.Size()
	sidePanelWithBackground := container.NewStack(sidePanelBackground, sidePanel)

	// --- Центральный контент ---
	tableRowsContainer := container.NewVBox() // Контейнер для строк таблицы
	scrollableTable := container.NewVScroll(tableRowsContainer)
	scrollableTable.SetMinSize(fyne.NewSize(0, 450)) // Минимальная высота для прокрутки

	// Строка поиска
	searchEntry := widget.NewEntry()
	searchEntry.PlaceHolder = "поиск                                     " // Как на макете
	searchEntry.OnChanged = func(s string) {
		updateUsersTableUI(s) // Обновляем список при изменении поиска
	}
	searchBox := container.New(layout.NewVBoxLayout(),
		widget.NewLabel("поиск"), // Метка "поиск"
		searchEntry,
		layout.NewSpacer(), // Растягивает поле поиска
	)

	// Заголовки таблицы: "ФИО, почта", "группы", "Статус"
	headerFIOEmail := widget.NewLabelWithStyle("ФИО почта", fyne.TextAlignCenter, fyne.TextStyle{Bold: false})
	headerGroup := widget.NewLabelWithStyle("группа", fyne.TextAlignCenter, fyne.TextStyle{Bold: false})
	headerStatus := widget.NewLabelWithStyle("Статус", fyne.TextAlignCenter, fyne.TextStyle{Bold: false})

	headerCellBackground := canvas.NewRectangle(color.White) // Темно-синий фон заголовков

	verticalHeaderDivider := canvas.NewRectangle(mediumGrayDivider)
	verticalHeaderDivider.SetMinSize(fyne.NewSize(1, 0))

	columnHeaders := container.New(layout.NewHBoxLayout(),
		container.NewMax(headerCellBackground, container.NewPadded(container.NewCenter(headerFIOEmail))),
		verticalHeaderDivider,
		container.NewMax(headerCellBackground, container.NewPadded(container.NewCenter(headerGroup))),
		verticalHeaderDivider,
		container.NewMax(headerCellBackground, container.NewPadded(container.NewCenter(headerStatus))),
	)
	columnHeadersContainer := container.New(layout.NewHBoxLayout(), columnHeaders, layout.NewSpacer())

	// --- Функция для создания одной строки таблицы пользователя ---
	// user: данные пользователя, idx: индекс в текущем отображаемом списке
	createUserTableRow := func(user *User, idx int) *fyne.Container {
		// Объединяем ФИО и Email для отображения в одной ячейке
		fioEmailCombinedLabel := widget.NewLabel(fmt.Sprintf("%s, %s", user.FIO, user.Email))
		fioEmailCombinedLabel.Wrapping = fyne.TextWrapWord

		groupLabel := widget.NewLabel(user.Group)
		groupLabel.Wrapping = fyne.TextWrapWord

		statusOptions := []string{"асс", "студ", "лек", "семи"}
		statusSelect := widget.NewSelect(statusOptions, func(selectedStatus string) {
			user.Status = selectedStatus // Обновляем статус пользователя в данных
			fmt.Printf("Статус пользователя %s (%s) изменен на: %s\n", user.FIO, user.Email, selectedStatus)
		})
		statusSelect.SetSelected(user.Status) // Устанавливаем текущий статус

		// Отступы и растягивание для ячеек
		cellFIOEmail := container.NewPadded(container.NewMax(fioEmailCombinedLabel))
		cellGroup := container.NewPadded(container.NewMax(groupLabel))
		cellStatus := container.NewPadded(container.NewMax(statusSelect))

		verticalCellDivider := canvas.NewRectangle(mediumGrayDivider)
		verticalCellDivider.SetMinSize(fyne.NewSize(1, 0))

		// Горизонтальная компоновка строки: ФИО+Email | Группа | Статус
		rowContainer := container.New(layout.NewVBoxLayout(),
			cellFIOEmail,
			verticalCellDivider,
			cellGroup,
			verticalCellDivider,
			cellStatus,
		)
		return rowContainer
	}

	// --- updateUsersTableUI: Функция для обновления всего UI таблицы пользователей ---
	// searchText: текст из поля поиска для фильтрации
	updateUsersTableUI = func(searchText string) {
		tableRowsContainer.RemoveAll() // Очищаем все текущие строки

		// Фильтруем пользователей
		filteredUsers := []*User{}
		if searchText == "" {
			filteredUsers = allUsers // Если поиск пуст, показываем всех
		} else {
			lowerSearchText := strings.ToLower(searchText)
			for _, user := range allUsers {
				if strings.Contains(strings.ToLower(user.FIO), lowerSearchText) ||
					strings.Contains(strings.ToLower(user.Email), lowerSearchText) {
					filteredUsers = append(filteredUsers, user)
				}
			}
		}
		currentDisplayedUsers = filteredUsers // Обновляем список отображаемых пользователей

		if len(currentDisplayedUsers) == 0 {
			tableRowsContainer.Add(container.NewCenter(widget.NewLabel("Нет пользователей для отображения по заданным критериям.")))
		} else {
			for i, user := range currentDisplayedUsers {
				tableRowsContainer.Add(createUserTableRow(user, i))
				tableRowsContainer.Add(canvas.NewRectangle(lightGrayDivider)) // Горизонтальный разделитель
			}
		}
		tableRowsContainer.Refresh()  // Важно обновить контейнер после всех изменений
		scrollableTable.ScrollToTop() // Прокрутка к началу после обновления
	}

	// Инициализируем UI таблицы пользователей при старте (отображаем всех)
	updateUsersTableUI("")

	// Центральный контент страницы
	centralContentPanel := container.NewVBox(
		container.NewPadded(searchBox), // Поле поиска
		columnHeadersContainer,         // Заголовки таблицы
		scrollableTable,                // Прокручиваемая таблица
	)

	contentBackground := canvas.NewRectangle(color.White)
	centralContentWithBackground := container.NewStack(contentBackground, container.NewPadded(centralContentPanel))

	// --- Общая компоновка окна ---
	w.SetContent(container.NewBorder(
		headerWithBackground,         // Верхняя панель
		nil,                          // Нижняя панель (пусто на этом макете)
		sidePanelWithBackground,      // Левая боковая панель
		nil,                          // Правая панель (пусто)
		centralContentWithBackground, // Центральный контент
	))

	w.ShowAndRun()
}

// main функция для запуска приложения.
