package main

import (
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/app"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/widget"
)

func main() {
    myApp := app.New()
    myWindow := myApp.NewWindow("Fyne Messenger")

    // Create a scrollable container for chat messages
    chatContainer := container.NewVBox()
    scrollContainer := container.NewVScroll(chatContainer)

    // Create an entry field for typing messages
    messageEntry := widget.NewEntry()
    messageEntry.SetPlaceHolder("Type your message...")

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

    // Create a container for the message entry and send button
    messageContainer := container.NewVBox(messageEntry)
    messageContainer.Resize(fyne.NewSize(300, 350))
    sendContainer := container.NewHBox(sendButton)
    inputContainer := container.NewHBox(messageContainer, sendContainer)
    inputContainer.Resize(fyne.NewSize(300,400))

    // Create a vertical split container for chat and input
    splitContainer := container.NewVSplit(scrollContainer, inputContainer)

    myWindow.SetContent(splitContainer)
    myWindow.Resize(fyne.NewSize(600,800))
    myWindow.ShowAndRun()
}

