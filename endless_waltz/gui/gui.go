package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"image/color"
	//"fyne.io/fyne/v2/dialog"

	"fmt"
	"github.com/sirupsen/logrus"
	"time"
)

//this is how you show dialog box
//dialog.ShowConfirm("foo", "foo", nil, myWindow)

//different layouts avail
//https://developer.fyne.io/explore/layouts.html#border

var users = []string{}
var targetUser = ""

func listen(logger *logrus.Logger, configuration Configurations) {
	cm, err := exConnect(logger, configuration, "server")
	if err != nil {
		return
	}
	defer cm.Close()
	for {
		//here's our server function, but it needs to write to gui
		handleConnection(cm, logger, configuration)
	}
}

func send(logger *logrus.Logger, configuration Configurations, container *fyne.Container, sendButton *fyne.Button) {
	cm, err := exConnect(logger, configuration, "client")
	if err != nil {
		return
	}
	defer cm.Close()
	for {
		message := <-outgoingMsgChan
		//set container to sending progressbar widget
		//container.
		targetUser := fmt.Sprintf("%s_%s", string(message.User), "server")
		ok := ew_client(logger, configuration, cm, message.Msg, targetUser)
		//reset container to prior
		incomingMsgChan <- Post{Msg: message.Msg, User: configuration.Server.User, ok: ok}
	}
}

func post(container *fyne.Container) {
	for {
		message := <-incomingMsgChan
		if message.ok {
			messageLabel := widget.NewLabel(fmt.Sprintf("%s: %s", message.User, message.Msg))
			container.Add(messageLabel)
		} else {
			messageLabel := widget.NewLabel(fmt.Sprintf("ERROR SENDING MSG %s", message.Msg))
			messageLabel.Importance = widget.DangerImportance
			container.Add(messageLabel)
		}
	}
}

func refreshUsers(logger *logrus.Logger, configuration Configurations, container *fyne.Container) {
	for {
		users = []string{}
		users, _ = getExUsers(logger, configuration)
		logger.Debug("refreshUsers --> ", users)
		container.Refresh()
		time.Sleep(5 * time.Second)
	}
}

func configureGUI(myWindow fyne.Window, logger *logrus.Logger, configuration Configurations) {
	// Create a scrollable container for chat messages
	chatContainer := container.NewVBox()
	scrollContainer := container.NewVScroll(chatContainer)
	scrollContainer.Resize(fyne.NewSize(500, 0))

	//set greeting warning lable
	messageLabel := widget.NewLabel("Select user to start sending messages")
	messageLabel.Importance = widget.MediumImportance
	chatContainer.Add(messageLabel)

	// Create an entry field for typing messages
	messageEntry := widget.NewMultiLineEntry()
	messageEntry.SetPlaceHolder("Type your message...")

	// add lines to use with onlinePanel
	text := widget.NewLabelWithStyle("    Online Users    ", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	topLine := canvas.NewLine(color.RGBA{0, 0, 0, 255})
	topLine.StrokeWidth = 5
	bLine := canvas.NewLine(color.RGBA{0, 0, 0, 255})
	bLine.StrokeWidth = 2
	sideLine := canvas.NewLine(color.RGBA{0, 0, 0, 255})
	sideLine.StrokeWidth = 5
	sideLine2 := canvas.NewLine(color.RGBA{0, 0, 0, 255})
	sideLine2.StrokeWidth = 5

	// add onlineUsers panel to show and select users
	onlineUsers := container.NewHBox(text)
	onlineUsers = container.NewBorder(topLine, bLine, nil, sideLine2, onlineUsers)
	onlineUsers = container.NewBorder(onlineUsers, nil, nil, sideLine)

	//add a goroutine here to read ExchangeAPI for live users and populate with labels
	go refreshUsers(logger, configuration, onlineUsers)

	//build our user list
	userList := widget.NewList(
		//length
		func() int {
			return len(users)
		},
		//create Item
		func() fyne.CanvasObject {
			label := widget.NewLabel("Text")
			return container.NewBorder(nil, nil, nil, nil, label)
		},
		//updateItem
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			text := obj.(*fyne.Container).Objects[0].(*widget.Label)
			text.SetText(users[id])
		})
	userList.OnSelected = func(id widget.ListItemID) {
		fmt.Println(users[id])
		targetUser = users[id]
		//clear the chat when switching users
		chatContainer.Objects = chatContainer.Objects[:0]
		chatContainer.Refresh()
		incomingMsgChan <- Post{Msg: users[id], User: "Sending messages to", ok: true}
		messageEntry.SetText("")
	}

	//actually add the users to the panel
	onlineUsers.Add(userList)
	//add container to hold the users list
	onlineContainer := container.New(layout.NewHBoxLayout(), onlineUsers)

	//define the sendbutton and OnClickFunc
	sendButton := widget.NewButton("Send", func() {
		// Get the message text from the entry field
		message := messageEntry.Text
		if message != "" {
		        //check, spelled like it sounds
			if targetUser == configuration.Server.User {
				incomingMsgChan <- Post{Msg: "Sending messages to yourself is not allowed", User: "foo", ok: false}
				return
			}

			//drop msg on correct channel
			outgoingMsgChan <- Post{Msg: message, User: targetUser, ok: true}

			// Clear the message entry field after sending
			messageEntry.SetText("")
		}
	})
	//turn the send button blue
	sendButton.Importance = widget.HighImportance
	buttonContainer := container.New(layout.NewVBoxLayout(), sendButton)

	//define the chat clear button
	clearButton := widget.NewButton("Clear", func() {
		//clear chatContainer and messageEntry
		chatContainer.Objects = chatContainer.Objects[:0]
		chatContainer.Refresh()
		messageEntry.SetText("")
	})
	clearButton.Importance = widget.DangerImportance

	//create the widget to display current user
	myText := widget.NewLabelWithStyle("Logged in as: "+configuration.Server.User, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	myText.Importance = widget.WarningImportance

	// Create a container for the message entry container, clear button widget and send button container
	sendContainer := container.NewBorder(clearButton, buttonContainer, nil, nil, messageEntry)

	// Create a vertical split container for chat and input
	splitContainer := container.NewVSplit(scrollContainer, sendContainer)
	splitContainer.Offset = .7
	//Create borders for buttons
	finalContainer := container.NewBorder(topLine, nil, onlineContainer, nil, splitContainer)
	finalContainer = container.NewBorder(myText, nil, nil, nil, finalContainer)

	myWindow.SetContent(finalContainer)
	myWindow.Resize(fyne.NewSize(600, 800))

	//replace button in buttonContainer with progressBar when firing message
	//https://developer.fyne.io/widget/progressbar
	//listen for incoming messages here
	go listen(logger, configuration)
	go send(logger, configuration, buttonContainer, sendButton)
	go post(chatContainer)

}

func main() {
	//add "starting up" message while loading

	//configuration stuff
	configuration, err := fetchConfig()
	if err != nil {
		return
	}
	logger := createLogger(configuration.Server.LogLevel, "normal")

	// Reading variables using the model
	logger.Debug("Reading variables using the model..")
	logger.Debug("keypath is\t\t", configuration.Server.Key)
	logger.Debug("crtpath is\t\t", configuration.Server.Cert)
	logger.Debug("randomURL is\t\t", configuration.Server.RandomURL)
	logger.Debug("exchangeURL is\t", configuration.Server.ExchangeURL)
	logger.Debug("user is\t\t", configuration.Server.User)
	logger.Debug("Passwd is\t\t", configuration.Server.Passwd)

	//have the user login every time -- it's no longer APIKeyAuth
	logger.Debug("Checking creds...")
	ok := checkCreds(configuration)
	if !ok {
		return
	}
	logger.Debug("creds passed check!")

	myApp := app.NewWithID("Main")
	myApp.Preferences().SetString("AppTimeout", fmt.Sprint(time.Minute))
	myWindow := myApp.NewWindow("EW Messenger")
	configureGUI(myWindow, logger, configuration)
	myWindow.ShowAndRun()
}
