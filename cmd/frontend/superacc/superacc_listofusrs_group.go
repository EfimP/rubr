package superacc

import (
	"fmt"
	"fyne.io/fyne/v2/dialog"
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// UserEntry represents the structure for one user in the list.
type UserEntry struct {
	FIOEmail string // ФИО, почта
	Status   string // Статус пользователя
}

// Simulated database of all users
var allUsers = []UserEntry{
	{"Иванов И.И., ivanov@mail.ru", "студ"},
	{"Петров П.П., petrov@mail.ru", "лек"},
	{"Сидоров С.С., sidorov@mail.ru", "асс"},
	{"Козлов К.К., kozlov@mail.ru", "студ"},
	{"Михайлов М.М., mikhailov@mail.ru", "семи"},
}

// ShowUsersListPage displays the "Users List" page for the selected group.
func ShowUsersListPage(groupName string) {
	a := app.New()
	a.Settings().SetTheme(theme.LightTheme()) // Set constant light theme
	logoText := canvas.NewText("ВШЭ", color.White)
	logoText.TextStyle.Bold = true
	logoText.TextSize = 24
	logoText.Alignment = fyne.TextAlignCenter
	leftHeaderObject := container.NewStack(logoText)

	headerTitle := canvas.NewText(fmt.Sprintf("Список пользователей: %s", groupName), color.White)
	headerTitle.TextStyle.Bold = true
	headerTitle.TextSize = 20
	headerTitle.Alignment = fyne.TextAlignCenter
	w := a.NewWindow(fmt.Sprintf("Список пользователей: %s", groupName)) // Window title
	w.Resize(fyne.NewSize(1200, 720))                                    // Increase window size for horizontal layout

	// Header content - logo on the left, title
	headerContent := container.New(layout.NewBorderLayout(nil, nil, leftHeaderObject, nil),
		leftHeaderObject,
		headerTitle,
	)

	// Back button for returning to the previous page
	backButton := widget.NewButton("Назад", func() {
		fmt.Println("Кнопка 'Назад' нажата. Возврат на предыдущую страницу.")
		w.Close() // Example: close the current window to return to the previous one
	})
	// Place the "Back" button in the top left corner, below the header
	backButtonRow := container.NewHBox(backButton, layout.NewSpacer())

	// --- Example data for the users table ---
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

	// Container for user table rows
	usersListContainer := container.NewVBox()

	// Table column headers
	headerFIOEmail := widget.NewLabelWithStyle("ФИО, почта", fyne.TextAlignCenter, fyne.TextStyle{Bold: false})
	headerStatus := widget.NewLabelWithStyle("Статус", fyne.TextAlignCenter, fyne.TextStyle{Bold: false})

	// Header background
	headerCellBackground := canvas.NewRectangle(color.NRGBA{R: 255, G: 255, B: 255, A: 255}) // White background

	// Vertical divider for headers
	verticalHeaderDivider := canvas.NewRectangle(color.NRGBA{R: 180, G: 180, B: 180, A: 255})
	verticalHeaderDivider.SetMinSize(fyne.NewSize(1, 0))

	// Header row (FIO, email | Status)
	columnHeaders := container.New(layout.NewHBoxLayout(),
		container.NewMax(headerCellBackground, container.NewPadded(container.NewCenter(headerFIOEmail))),
		verticalHeaderDivider,
		container.NewMax(headerCellBackground, container.NewPadded(container.NewCenter(headerStatus))),
	)
	columnHeadersContainer := container.New(layout.NewHBoxLayout(), columnHeaders, layout.NewSpacer())

	// Declare updateUsersListUI here for use in closures
	var updateUsersListUI func()

	// Function to create a single user row
	createUserRow := func(user UserEntry, idx int) *fyne.Container {
		fioEmailLabel := widget.NewLabel(user.FIOEmail)
		fioEmailLabel.Wrapping = fyne.TextWrapWord

		// Use widget.Select for status as per the mockup ("opens as a list")
		statusOptions := []string{"асс", "студ", "лек", "семи"}
		statusSelect := widget.NewSelect(statusOptions, func(selected string) {
			usersData[idx].Status = selected // Update data
			fmt.Printf("Статус пользователя %s изменен на: %s\n", usersData[idx].FIOEmail, selected)
		})
		statusSelect.SetSelected(user.Status) // Set current status

		// Delete button
		deleteButton := widget.NewButton("Удалить", func() {
			dialog.ShowConfirm(
				"Подтверждение удаления",
				fmt.Sprintf("Удалить пользователя '%s'?", user.FIOEmail),
				func(confirmed bool) {
					if confirmed {
						usersData = append(usersData[:idx], usersData[idx+1:]...)
						updateUsersListUI()
						fmt.Printf("Удален пользователь: %s\n", user.FIOEmail)
					}
				},
				w,
			)
		})

		// Wrap elements in Padded and MaxLayout for padding and stretching
		cellFIOEmail := container.NewPadded(container.NewMax(fioEmailLabel))
		cellStatus := container.NewPadded(container.NewMax(statusSelect))
		cellDelete := container.NewPadded(container.NewMax(deleteButton))

		// Vertical dividers for cells
		verticalCellDivider1 := canvas.NewRectangle(color.NRGBA{R: 180, G: 180, B: 180, A: 255})
		verticalCellDivider1.SetMinSize(fyne.NewSize(1, 0))
		verticalCellDivider2 := canvas.NewRectangle(color.NRGBA{R: 180, G: 180, B: 180, A: 255})
		verticalCellDivider2.SetMinSize(fyne.NewSize(1, 0))

		// Table row (FIO, email | Status | Delete)
		rowContainer := container.New(layout.NewVBoxLayout(),
			cellFIOEmail,
			verticalCellDivider1,
			cellStatus,
			verticalCellDivider2,
			cellDelete,
		)
		return rowContainer
	}

	// Function to update the entire users list UI
	updateUsersListUI = func() {
		usersListContainer.RemoveAll() // Clear all current rows

		if len(usersData) == 0 {
			usersListContainer.Add(container.NewCenter(widget.NewLabel("Нет пользователей для отображения")))
			usersListContainer.Refresh()
			return
		}

		for i, user := range usersData {
			usersListContainer.Add(createUserRow(user, i))
			// Horizontal divider between rows
			usersListContainer.Add(canvas.NewRectangle(color.NRGBA{R: 220, G: 220, B: 220, A: 255}))
		}
		usersListContainer.Refresh()
	}

	// Initialize users list UI at start
	updateUsersListUI()

	// Scrollable area for table rows
	scrollableUsersList := container.NewVScroll(usersListContainer)
	scrollableUsersList.SetMinSize(fyne.NewSize(215, 450)) // Set minimum height for scrolling

	// "Добавить" button with search dialog
	addButton := widget.NewButton("Добавить", func() {
		// Search dialog
		searchEntry := widget.NewEntry()
		searchEntry.SetPlaceHolder("Введите имя для поиска...")
		var filteredUsers []UserEntry

		updateFilteredUsers := func() {
			query := strings.ToLower(searchEntry.Text)
			filteredUsers = nil
			for _, user := range allUsers {
				if strings.Contains(strings.ToLower(user.FIOEmail), query) {
					filteredUsers = append(filteredUsers, user)
				}
			}
		}

		// List to display search results
		userList := widget.NewList(
			func() int { return len(filteredUsers) },
			func() fyne.CanvasObject {
				return widget.NewLabel("ФИО, почта")
			},
			func(id widget.ListItemID, item fyne.CanvasObject) {
				item.(*widget.Label).SetText(filteredUsers[id].FIOEmail)
			},
		)
		userList.OnSelected = func(id widget.ListItemID) {
			selectedUser := filteredUsers[id]
			usersData = append(usersData, selectedUser)
			updateUsersListUI()
			fmt.Printf("Добавлен пользователь: %s\n", selectedUser.FIOEmail)
		}

		// Update filtered users on text change
		searchEntry.OnChanged = func(s string) {
			updateFilteredUsers()
			userList.Refresh()
		}
		updateFilteredUsers() // Initial update

		// Dialog content
		dialogContent := container.NewVBox(
			searchEntry,
			container.NewVScroll(userList),
		)

		// Show dialog
		dialog.ShowCustom("Поиск пользователя", "Закрыть", dialogContent, w)
	})

	// Central content with background (white rectangle)
	contentBackground := canvas.NewRectangle(color.White)
	usersPanel := container.NewVBox(
		backButtonRow, // Back button
		columnHeadersContainer,
		scrollableUsersList,
		addButton, // Add button
	)
	centralContent := container.NewStack(
		contentBackground,
		container.NewPadded(usersPanel),
	)
	//люблю соню
	// Page layout: header, sidebar, central content
	w.SetContent(container.NewStack(
		canvas.NewRectangle(color.NRGBA{R: 20, G: 40, B: 80, A: 255}), // Window background
		container.NewBorder(
			headerContent,  // Top (header)
			nil,            // Bottom (empty)
			nil,            // Left (empty, but sidebar exists in mockup)
			nil,            // Right (empty)
			centralContent, // Central content (users table)
		),
	))
	w.ShowAndRun()
}
