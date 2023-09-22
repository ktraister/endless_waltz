package main

import (
    "fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("AIM Messenger Layout")

	// Buddy list on the left
	buddyList := container.NewVBox(
		widget.NewLabel("Buddy 1"),
		widget.NewLabel("Buddy 2"),
		widget.NewLabel("Buddy 3"),
		// Add more buddy labels as needed
	)

	// Chat window on the right
	chatWindow := widget.NewMultiLineEntry()
	chatWindow.MultiLine = true
	chatWindow.Disable()

	// Split container to organize buddy list and chat window
	splitContainer := container.NewHSplit(buddyList, chatWindow)

	myWindow.SetContent(splitContainer)
	myWindow.Resize(fyne.NewSize(400, 300))
	myWindow.ShowAndRun()
}

