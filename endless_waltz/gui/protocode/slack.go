package main

import (
    	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Slack-Inspired Layout")

	// Create a sidebar with channels (you can customize this)
	channelList := widget.NewList(
		func() int {
			return 10 // Replace with the number of channels you have
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Channel Name") // Replace with your channel names
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			// Implement click event to switch to the selected channel's chat
			//item.(*widget.Label).SetStyle(widget.Bold)
		},
	)

	// Create a chat area (main content)
	chatArea := container.NewVBox(
		widget.NewLabel("Welcome to Slack"),
	)

	// Create a horizontal split container for sidebar and chat area
	splitContainer := container.NewHSplit(
		container.NewVScroll(channelList),
		chatArea,
	)
	splitContainer.Offset = 0.25 // Adjust the initial split position

	// Create a top bar with the Slack logo and user profile (you can customize this)
	topBar := container.NewHBox(
		widget.NewIcon(theme.HomeIcon()),
		widget.NewLabel("Slack"),
		layout.NewSpacer(),
		//widget.NewIcon(theme.PersonIcon()),
		widget.NewLabel("Username"),
	)

	// Create a content container with top bar and split container
	contentContainer := container.NewVBox(
		topBar,
		splitContainer,
	)

	myWindow.SetContent(contentContainer)
	myWindow.Resize(fyne.NewSize(800, 600))
	myWindow.ShowAndRun()
}
