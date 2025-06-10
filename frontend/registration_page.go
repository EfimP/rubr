package frontend

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"image/color"
)

func RegistrationPage() {
	a := app.New()
	w := a.NewWindow("Greeting page")

	// Создаем фоновое изображение для левой части
	leftBackground := canvas.NewImageFromFile("bin/logo/hse_logo.svg")
	leftBackground.FillMode = canvas.ImageFillStretch // Растягиваем изображение для заполнения области

	// Создаем поля ввода и кнопку
	loginEntry := widget.NewEntry()
	loginEntry.SetPlaceHolder("Введите почту")

	NameEntry := widget.NewEntry()
	NameEntry.SetPlaceHolder("Имя")

	SurnameEntry := widget.NewEntry()
	SurnameEntry.SetPlaceHolder("Фамилия")

	PatronymicEntry := widget.NewEntry()
	PatronymicEntry.SetPlaceHolder("Отчество")

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Введите пароль")

	EnterButton := widget.NewButton("Зарегистрироваться", func() {
		fmt.Printf("Login: %s\nPassword: %s\nName: %s\nSurName: %s\nPatronymic: %s\n",
			loginEntry.Text, passwordEntry.Text, NameEntry.Text, SurnameEntry.Text, PatronymicEntry.Text)
	})
	EnterButton.Importance = widget.HighImportance

	// Центрируем кнопки вертикально в левой части
	leftContent := container.NewVBox(
		layout.NewSpacer(),
		NameEntry,
		SurnameEntry,
		PatronymicEntry,
		loginEntry,
		passwordEntry,
		EnterButton,
		layout.NewSpacer(),
	)
	leftContainer := container.NewStack(leftBackground, container.NewCenter(leftContent))

	// Создаем прямоугольник для фона правой части
	rightBackground := canvas.NewRectangle(color.RGBA{23, 44, 101, 255})

	// Создаем надпись
	rightText := widget.NewLabel("Зарегистрируйтесь")
	rightText.TextStyle = fyne.TextStyle{Bold: true}

	// Центрируем надпись в правой части
	rightContent := container.NewCenter(rightText)
	rightContainer := container.NewStack(rightBackground, rightContent)

	// Строгое разделение пополам
	split := container.New(layout.NewGridLayout(2), leftContainer, rightContainer)
	w.SetContent(split)

	w.Resize(fyne.NewSize(1280, 720))
	w.ShowAndRun()
}
