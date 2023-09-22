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

	"net/http"
	"net/url"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

//this is how you show dialog box
//dialog.ShowConfirm("foo", "foo", nil, myWindow)

//different layouts avail
//https://developer.fyne.io/explore/layouts.html#border

func configureGUI(myWindow fyne.Window, logger *logrus.Logger, configuration Configurations, conn *websocket.Conn) {
	// Create a scrollable container for chat messages
	chatContainer := container.NewVBox()
	scrollContainer := container.NewVScroll(chatContainer)
	scrollContainer.Resize(fyne.NewSize(500, 0))

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
		        //ohh shit we have to configure the user too
		        //send the message thru the EW circut
                        go ew_client(logger, configuration, conn, message, "Kayleigh")

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

	myApp := app.New()
	myWindow := myApp.NewWindow("EW Messenger")
	configureGUI(myWindow, logger, configuration, conn)
	myWindow.ShowAndRun()
}
