package everybody

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

func GreetingPage() {
	a := app.New()
	w := a.NewWindow("Greeting page")

	// Создаем фоновое изображение для левой части
	leftBackground := canvas.NewImageFromFile("bin/logo/hse_logo.svg")
	leftBackground.FillMode = canvas.ImageFillStretch // Растягиваем изображение для заполнения области

	// Создаем кнопки
	loginButton := widget.NewButton("Авторизоваться", func() {
		fmt.Println("authorization button clicked")
	})
	loginButton.Importance = widget.HighImportance

	registerButton := widget.NewButton("Зарегистрироваться", func() {
		fmt.Println("registration button clicked")
	})
	registerButton.Importance = widget.MediumImportance

	// Центрируем кнопки вертикально в левой части
	leftContent := container.NewVBox(
		layout.NewSpacer(),
		loginButton,
		registerButton,
		layout.NewSpacer(),
	)
	leftContainer := container.NewStack(leftBackground, container.NewCenter(leftContent))

	// Создаем прямоугольник для фона правой части
	rightBackground := canvas.NewRectangle(color.RGBA{23, 44, 101, 255})

	// Создаем надпись
	rightText := widget.NewLabel("Вход в систему оценивания")
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
