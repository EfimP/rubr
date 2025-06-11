// Пример вызова для тестирования (можно добавить в main.go)
/*
func main() {
	superacc.ShowUsersListPage("Название вашей группы")
}
*/
package superacc

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme" // Импортируем тему для LightTheme()
	"fyne.io/fyne/v2/widget"
)

// UserEntry представляет структуру данных для одного пользователя в списке.
type UserEntry struct {
	FIOEmail string // ФИО, почта
	Status   string // Статус пользователя
}

// ShowUsersListPage отображает страницу "Список пользователей" для выбранной группы.
// groupName - название группы, для которой отображаются пользователи.
func ShowUsersListPage(groupName string) {
	a := app.New()
	a.Settings().SetTheme(theme.LightTheme()) // Устанавливаем постоянную белую тему
	logoText := canvas.NewText("ВШЭ", color.White)
	logoText.TextStyle.Bold = true
	logoText.TextSize = 24
	logoText.Alignment = fyne.TextAlignCenter
	leftHeaderObject := container.NewStack(logoText)

	headerTitle := canvas.NewText(fmt.Sprintf("Список пользователей: %s", groupName), color.White)
	headerTitle.TextStyle.Bold = true
	headerTitle.TextSize = 20
	headerTitle.Alignment = fyne.TextAlignCenter
	w := a.NewWindow(fmt.Sprintf("Список пользователей: %s", groupName)) // Название окна
	w.Resize(fyne.NewSize(1200, 720))                                    // Увеличим размер окна еще больше для горизонтального расположения

	// Заголовок "Список пользователей {Название группы}"

	// Шапка окна - заголовок слева, предмет справа
	headerContent := container.New(layout.NewBorderLayout(nil, nil, leftHeaderObject, nil),
		leftHeaderObject,
		headerTitle,
	)

	// Кнопка "Назад" для возврата на предыдущую страницу
	backButton := widget.NewButton("Назад", func() {
		fmt.Println("Кнопка 'Назад' нажата. Возврат на предыдущую страницу.")
		w.Close() // Пример: закрыть текущее окно, чтобы вернуться к предыдущему
	})
	// Разместим кнопку "Назад" в верхнем левом углу, ниже шапки
	backButtonRow := container.NewHBox(backButton, layout.NewSpacer())

	// --- Пример данных для таблицы пользователей ---
	var usersData []UserEntry
	usersData = []UserEntry{
		{"Иванов И.И., ivanov@mail.ru", "студ"},
		{"Иванов И.И., ivanov@mail.ru", "студ"},
		{"Иванов И.И., ivanov@mail.ru", "студ"},
		{"Иванов И.И., ivanov@mail.ru", "студ"},
		{"Иванов И.И., ivanov@mail.ru", "студ"},
		{"Иванов И.И., ivanov@mail.ru", "студ"},
		{"Иванов И.И., ivanov@mail.ru", "студ"},
		{"Иванов И.И., ivanov@mail.ru", "студ"},
		{"Иванов И.И., ivanov@mail.ru", "студ"},
		{"Иванов И.И., ivanov@mail.ru", "студ"},
		{"Иванов И.И., ivanov@mail.ru", "студ"},
		{"Иванов И.И., ivanov@mail.ru", "студ"},
		{"Иванов И.И., ivanov@mail.ru", "студ"},
		{"Иванов И.И., ivanov@mail.ru", "студ"},
		{"Иванов И.И., ivanov@mail.ru", "студ"},
		{"Иванов И.И., ivanov@mail.ru", "студ"},
	}

	// Контейнер для строк таблицы пользователей
	usersListContainer := container.NewVBox()

	// Заголовки столбцов таблицы
	headerFIOEmail := widget.NewLabelWithStyle("ФИО, почта", fyne.TextAlignCenter, fyne.TextStyle{Bold: false})
	headerStatus := widget.NewLabelWithStyle("Статус", fyne.TextAlignCenter, fyne.TextStyle{Bold: false})

	// Фон заголовков
	headerCellBackground := canvas.NewRectangle(color.NRGBA{R: 255, G: 255, B: 255, A: 255}) // Темно-синий

	// Вертикальный разделитель для заголовков
	verticalHeaderDivider := canvas.NewRectangle(color.NRGBA{R: 180, G: 180, B: 180, A: 255})
	verticalHeaderDivider.SetMinSize(fyne.NewSize(1, 0))

	// Заголовочная строка таблицы (ФИО, почта | Статус)
	// Используем NewHBoxLayout для растягивания по горизонтали
	columnHeaders := container.New(layout.NewHBoxLayout(),
		container.NewMax(headerCellBackground, container.NewPadded(container.NewCenter(headerFIOEmail))),
		verticalHeaderDivider,
		container.NewMax(headerCellBackground, container.NewPadded(container.NewCenter(headerStatus))),
	)
	// Добавим Spacer, чтобы заголовки растягивались на всю ширину
	columnHeadersContainer := container.New(layout.NewHBoxLayout(), columnHeaders, layout.NewSpacer())

	// Объявляем updateUsersListUI здесь для использования в замыканиях
	var updateUsersListUI func()

	// Функция для создания одной строки пользователя
	createUserRow := func(user UserEntry, idx int) *fyne.Container {
		fioEmailLabel := widget.NewLabel(user.FIOEmail)
		fioEmailLabel.Wrapping = fyne.TextWrapWord

		// Для статуса используем widget.Select, как указано в макете ("открывается как список")
		statusOptions := []string{"асс", "студ", "лек", "семи"}
		statusSelect := widget.NewSelect(statusOptions, func(selected string) {
			usersData[idx].Status = selected // Обновляем данные
			fmt.Printf("Статус пользователя %s изменен на: %s\n", usersData[idx].FIOEmail, selected)
			// Нет необходимости вызывать updateUsersListUI, так как widget.Select сам обновляет свой текст
		})
		statusSelect.SetSelected(user.Status) // Устанавливаем текущий статус

		// Оборачиваем элементы в Padded и MaxLayout для отступов и растягивания
		// Используем NewMax, чтобы эти ячейки максимально использовали доступное пространство
		cellFIOEmail := container.NewPadded(container.NewMax(fioEmailLabel))
		cellStatus := container.NewPadded(container.NewMax(statusSelect))

		// Вертикальный разделитель для ячеек
		verticalCellDivider := canvas.NewRectangle(color.NRGBA{R: 180, G: 180, B: 180, A: 255})
		verticalCellDivider.SetMinSize(fyne.NewSize(1, 0))

		// Строка таблицы (ФИО, почта | Статус)
		// Используем NewHBoxLayout для равномерного распределения ячеек по горизонтали
		rowContainer := container.New(layout.NewCustomPaddedVBoxLayout(1),
			cellFIOEmail,
			verticalCellDivider,
			cellStatus,
		)
		return rowContainer
	}

	// Функция для обновления всего UI списка пользователей
	updateUsersListUI = func() {
		usersListContainer.RemoveAll() // Очищаем все текущие строки

		if len(usersData) == 0 {
			usersListContainer.Add(container.NewCenter(widget.NewLabel("Нет пользователей для отображения")))
			usersListContainer.Refresh()
			return
		}

		for i, user := range usersData {
			usersListContainer.Add(createUserRow(user, i))
			// Горизонтальный разделитель между строками
			usersListContainer.Add(canvas.NewRectangle(color.NRGBA{R: 220, G: 220, B: 220, A: 255}))
		}
		usersListContainer.Refresh()
	}

	// Инициализируем UI списка пользователей при старте
	updateUsersListUI()

	// Скроллируемая область для строк таблицы
	scrollableUsersList := container.NewVScroll(usersListContainer)
	scrollableUsersList.SetMinSize(fyne.NewSize(0, 450)) // Устанавливаем минимальную высоту для прокрутки

	// Центральный контент с фоном (белый прямоугольник)
	contentBackground := canvas.NewRectangle(color.White)
	usersPanel := container.NewVBox(
		backButtonRow, // Кнопка "Назад"
		columnHeadersContainer,
		scrollableUsersList,
	)
	centralContent := container.NewStack(
		contentBackground,
		container.NewPadded(usersPanel),
	)

	// Компоновка страницы: заголовок, боковая панель, центральный контент
	w.SetContent(container.NewStack(
		canvas.NewRectangle(color.NRGBA{R: 20, G: 40, B: 80, A: 255}), // Фон окна
		container.NewBorder(

			headerContent,  // Верхняя часть (заголовок)
			nil,            // Нижняя часть (пусто)
			nil,            // Левая часть (пока пусто, но на макете есть боковая панель)
			nil,            // Правая часть (пусто)
			centralContent, // Центральный контент (таблица пользователей)
		),
	))
	w.ShowAndRun()
}
