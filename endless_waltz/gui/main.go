package main

import (
    "image/color"
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/app"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/widget"
    "fyne.io/fyne/v2/layout"
    "fyne.io/fyne/v2/canvas"
)

//different layouts avail
//https://developer.fyne.io/explore/layouts.html#border

//its starting to really look like I want a custom layout for my GUI. I'll sketch it out in my notes for TN

func main() {
    myApp := app.New()
    myWindow := myApp.NewWindow("EW Messenger")

    // Create a scrollable container for chat messages
    chatContainer := container.NewVBox()
    scrollContainer := container.NewVScroll(chatContainer)

    // Create an entry field for typing messages
    messageEntry := widget.NewMultiLineEntry()
    messageEntry.SetPlaceHolder("Type your message...")

    text := widget.NewLabel("    Online Users    ")
    topLine := canvas.NewLine(color.RGBA{0, 0, 0, 255})
    topLine.StrokeWidth = 5
    bLine := canvas.NewLine(color.RGBA{0, 0, 0, 255})
    bLine.StrokeWidth = 2
    sideLine := canvas.NewLine(color.RGBA{0, 0, 0, 255})
    sideLine.StrokeWidth = 5
    sideLine2 := canvas.NewLine(color.RGBA{0, 0, 0, 255})
    sideLine2.StrokeWidth = 5
    onlineUsers := container.NewHBox(text)
    onlineUsers = container.NewBorder(topLine, bLine, nil, sideLine2, onlineUsers)
    onlineUsers = container.NewBorder(onlineUsers, nil, nil, sideLine)

    //add to above entry field this enter feature
    //https://developer.fyne.io/explore/layouts.html#border

    sendButton := widget.NewButton("Send", func() {
        // Get the message text from the entry field
        message := messageEntry.Text
        if message != "" {
            // Create a label widget for the message and add it to the chat container
            messageLabel := widget.NewLabel("You: " + message)
            chatContainer.Add(messageLabel)

            // Clear the message entry field after sending
            messageEntry.SetText("")
        }
    })
    //turn the send button blue
    sendButton.Importance = widget.HighImportance

    clearButton := widget.NewButton("Clear", func() {
	// Create a label widget for the message and add it to the chat container
	chatContainer.Objects = chatContainer.Objects[:0]
	//ensure UI change is written
	chatContainer.Refresh()

	// Clear the message entry field after sending
	messageEntry.SetText("")
    })
    clearButton.Importance = widget.DangerImportance

    //layout.NewBorderLayout(toolbarLines, nil, middleLines, nil)

    // Create a container for the message entry and send button
    //messageContainer := container.New(layout.NewMaxLayout(), messageEntry)
    //sendContainer := container.New(layout.NewVBoxLayout(), sendButton)
    //clearContainer := container.New(layout.NewVBoxLayout(), clearButton)
    onlineContainer := container.New(layout.NewHBoxLayout(), onlineUsers)
    sendContainer := container.NewBorder(clearButton, sendButton, nil, nil, messageEntry)

    // Create a vertical split container for chat and input
    //splitContainer := container.NewVSplit(messageContainer, sendContainer)
    //splitContainer = container.NewVSplit(clearContainer, splitContainer)
    splitContainer := container.NewVSplit(scrollContainer, sendContainer)
    //Create another vertical split for chat and input
    finalContainer := container.NewBorder(nil, nil, onlineContainer, nil, splitContainer)

    myWindow.SetContent(finalContainer)
    myWindow.Resize(fyne.NewSize(600,800))
    myWindow.ShowAndRun()
}

