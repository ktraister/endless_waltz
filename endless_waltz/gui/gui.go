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

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"encoding/json"
	"net/http"
	"net/url"
	"time"
	"fmt"
)

//this is how you show dialog box
//dialog.ShowConfirm("foo", "foo", nil, myWindow)

//different layouts avail
//https://developer.fyne.io/explore/layouts.html#border

var users = []string{"Kayleigh", "KayleighToo"}
var targetUser = ""

func configureGUI(myWindow fyne.Window, logger *logrus.Logger, configuration Configurations, conn *websocket.Conn) {
	// Create a scrollable container for chat messages
	chatContainer := container.NewVBox()
	scrollContainer := container.NewVScroll(chatContainer)
	scrollContainer.Resize(fyne.NewSize(500, 0))

	// Create an entry field for typing messages
	messageEntry := widget.NewMultiLineEntry()
	messageEntry.SetPlaceHolder("Type your message...")

	// TODO: add a box at top/bottom left for currentUser

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
        
	/*
	//actually add the users to the panel
	onlineUsers.Add(widget.NewLabel("TestUser"))
	*/

	//below is code from the example fyne page
	//use it to expand this functionality
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
	    messageEntry.SetText("")
        } 

	//actually add the users to the panel
	onlineUsers.Add(userList)

	//need to add a goroutine here to listen for messages, a goroutine to populate new labels, and a chan to communicate

	sendButton := widget.NewButton("Send", func() {
		// Get the message text from the entry field
		message := messageEntry.Text
		if message != "" {
			//ohh shit we have to configure the user too
			//send the message thru the EW circut
			//add something here to return false if the send fails, true if success
			ok := ew_client(logger, configuration, conn, message, targetUser)

			if ok {
				// Create a label widget for the message and add it to the chat container
				messageLabel := widget.NewLabel("You: " + message)
				chatContainer.Add(messageLabel)
			} else {
				messageLabel := widget.NewLabel("FAILED TO SEND: " + message)
				messageLabel.Importance = widget.DangerImportance
				chatContainer.Add(messageLabel)
			}

			// Clear the message entry field after sending
			messageEntry.SetText("")
		}
	})
	//turn the send button blue
	sendButton.Importance = widget.HighImportance

	clearButton := widget.NewButton("Clear", func() {
	        //clear chatContainer and messageEntry
		chatContainer.Objects = chatContainer.Objects[:0]
		chatContainer.Refresh()
		messageEntry.SetText("")
	})
	clearButton.Importance = widget.DangerImportance

	// Create a container for the message entry and send button
	onlineContainer := container.New(layout.NewHBoxLayout(), onlineUsers)
	sendContainer := container.NewBorder(clearButton, sendButton, nil, nil, messageEntry)

	// Create a vertical split container for chat and input
	splitContainer := container.NewVSplit(scrollContainer, sendContainer)
	splitContainer.Offset = .7
	//Create another vertical split for chat and input
	finalContainer := container.NewBorder(nil, nil, onlineContainer, nil, splitContainer)

	myWindow.SetContent(finalContainer)
	myWindow.Resize(fyne.NewSize(600, 800))

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

	// Parse the WebSocket URL
	u, err := url.Parse(configuration.Server.ExchangeURL)
	if err != nil {
		logger.Fatal(err)
	}

	// Establish a WebSocket connection
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), http.Header{"Passwd": []string{configuration.Server.Passwd}, "User": []string{configuration.Server.User}})
	if err != nil {
		logger.Fatal("Could not establish WebSocket connection with ", u.String())
		return
	}
	logger.Debug("Connected to exchange server!")

	defer conn.Close()

        //connect to exchange with our username for mapping
        message := &Message{Type: "startup", User: configuration.Server.User}
        b, err := json.Marshal(message)
        if err != nil {
                fmt.Println(err)
                return
        }  
        err = conn.WriteMessage(websocket.TextMessage, b)
        if err != nil {
                logger.Fatal(err)
        } 

	myApp := app.NewWithID("Main")
	myApp.Preferences().SetString("AppTimeout", string(time.Minute))
	myWindow := myApp.NewWindow("EW Messenger")
	configureGUI(myWindow, logger, configuration, conn)
	myWindow.ShowAndRun()
}
