package main

import (
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/app"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/widget"
    "fyne.io/fyne/v2/layout"
)

//different layouts avail
//https://developer.fyne.io/explore/layouts.html#border

//its starting to really look like I want a custom layout for my GUI. I'll sketch it out in my notes for TN

func main() {
    myApp := app.New()
    myWindow := myApp.NewWindow("Fyne Messenger")

    // Create a scrollable container for chat messages
    chatContainer := container.NewVBox()
    scrollContainer := container.NewVScroll(chatContainer)

    // Create an entry field for typing messages
    messageEntry := widget.NewMultiLineEntry()
    messageEntry.SetPlaceHolder("Type your message...")

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

    //layout.NewBorderLayout(toolbarLines, nil, middleLines, nil)

    // Create a container for the message entry and send button
    messageContainer := container.New(layout.NewBorderLayout(nil, nil, nil, nil), messageEntry)
    sendContainer := container.New(layout.NewVBoxLayout(), sendButton)

    //Will eventually need an online container in Border lefthand orientation

    // Create a vertical split container for chat and input
    splitContainer := container.NewVSplit(messageContainer, sendContainer)
    //Create another vertical split for chat and input
    splitContainer = container.NewVSplit(scrollContainer, splitContainer)

    myWindow.SetContent(splitContainer)
    myWindow.Resize(fyne.NewSize(600,800))
    myWindow.ShowAndRun()
}

