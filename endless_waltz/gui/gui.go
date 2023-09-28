package main

import (
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/sirupsen/logrus"
	"image/color"
	"net/http"
	"strings"
	"time"
)

//this is how you show dialog box
//dialog.ShowConfirm("foo", "foo", nil, myWindow)

//different layouts avail
//https://developer.fyne.io/explore/layouts.html#border

var users = []string{}
var targetUser = ""

func checkCreds(configuration Configurations) (bool, string) {
	//check and make sure inserted creds
	//Random and Exchange will use same mongo, so the creds will be valid for both

	health_url := fmt.Sprintf("%s%s", strings.Split(configuration.RandomURL, "/otp")[0], "/healthcheck")
	req, err := http.NewRequest("GET", health_url, nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("User", configuration.User)
	req.Header.Set("Passwd", configuration.Passwd)
	client := http.Client{Timeout: 3 * time.Second}
	resp, err := client.Do(req)
	errorText := ""
	if err != nil {
		errorText = "Couldn't Connect to RandomAPI"
		fmt.Println(errorText + " " + configuration.RandomURL)
		fmt.Println("Quietly exiting now. Please reconfigure.")
		return false, errorText
	}
	if resp == nil {
		errorText = "No Response From RandomAPI"
		fmt.Println(errorText + " " + configuration.RandomURL)
		fmt.Println("Quietly exiting now. Please reconfigure.")
		return false, errorText
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		errorText = "Invalid Username/Password Combination"
		fmt.Println(errorText)
		fmt.Printf("Request failed with status: %s\n", resp.Status)
		return false, errorText
	}
	return true, ""
}

//this thread should just read HELO and pass off to another thread
func listen(logger *logrus.Logger, configuration Configurations) {
	localUser := fmt.Sprintf("%s_%s", configuration.User, "server")
	cm, err := exConnect(logger, configuration, localUser)
	if err != nil {
		return
	}
	defer cm.Close()

	//listen for incoming connections
	for {
		_, incoming, err := cm.Read()
		if err != nil {
			logger.Error("Error reading message:", err)
			continue
		}

		err = json.Unmarshal([]byte(incoming), &dat)
		if err != nil {
			logger.Error("Error unmarshalling json:", err)
			continue
		}

		//new connections should always ask
		if dat["msg"] == "HELO" {
			logger.Debug("Received HELO from ", dat["from"])
		} else {
			logger.Warn("New connection didn't HELO, bouncing")
			continue
		}
		//handle connection creates new socket inside goRoutine
		go handleConnection(dat, logger, configuration)
	}
}

//send needs to be a wrapper thread for go functions
func send(logger *logrus.Logger, configuration Configurations, sendButton *widget.Button, progressBar *widget.ProgressBarInfinite) {
	for {
		message := <-outgoingMsgChan
		//set container to sending progressbar widget
		sendButton.Hide()
		progressBar.Show()
		//set container to sending progressbar widget

		//update user and send message to server root socket
		ok := ew_client(logger, configuration, message)
		//reset container to prior
		sendButton.Show()
		progressBar.Hide()

		//post our sent message
		incomingMsgChan <- Post{Msg: message.Msg, User: configuration.User, ok: ok}
	}
}

func post(container *fyne.Container) {
	for {
		message := <-incomingMsgChan
		if message.User == "Sending messages to" {
			messageLabel := widget.NewLabelWithStyle("Sending messages to: "+message.Msg, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
			container.Add(messageLabel)
		} else if message.ok {
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
	messageLabel := widget.NewLabelWithStyle("Select a user to send messages", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	messageLabel.Importance = widget.MediumImportance
	chatContainer.Add(messageLabel)

	// Create an entry field for typing messages
	messageEntry := widget.NewMultiLineEntry()
	messageEntry.SetPlaceHolder("Type your message...")
	//hiding the entry until a user is selected
	//come up with something cute to go here
	messageEntry.Hide()

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
		messageEntry.Show()
		fmt.Println(users[id])
		targetUser = users[id]
		//clear the chat when switching users
		chatContainer.Objects = chatContainer.Objects[:0]
		messageLabel.Hide()
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
			if targetUser == configuration.User {
				incomingMsgChan <- Post{Msg: "Sending messages to yourself is not allowed", User: "foo", ok: false}
				return
			}

			//drop the messsage on the outgoing channel
			outgoingMsgChan <- Post{Msg: message, User: targetUser, ok: true}

			// Clear the message entry field after sending
			messageEntry.SetText("")
		}
	})
	//turn the send button blue
	sendButton.Importance = widget.HighImportance

	//define progress bar to use when sending a message
	infinite := widget.NewProgressBarInfinite()
	buttonContainer := container.New(layout.NewVBoxLayout(), infinite)
	buttonContainer.Add(sendButton)
	infinite.Hide()

	//define the chat clear button
	clearButton := widget.NewButton("Clear", func() {
		//clear chatContainer and messageEntry
		chatContainer.Objects = chatContainer.Objects[:0]
		chatContainer.Refresh()
		messageEntry.SetText("")
	})
	clearButton.Importance = widget.DangerImportance

	//create the widget to display current user
	myText := widget.NewLabelWithStyle("Logged in as: "+configuration.User, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	myText.Importance = widget.WarningImportance

	// Create a container for the message entry container, clear button widget and send button container
	sendContainer := container.NewBorder(clearButton, buttonContainer, nil, nil, messageEntry)

	// Create a vertical split container for chat and input
	splitContainer := container.NewVSplit(scrollContainer, sendContainer)
	splitContainer.Offset = .7
	//Create borders for buttons
	finalContainer := container.NewBorder(topLine, nil, onlineContainer, nil, splitContainer)
	finalContainer = container.NewBorder(myText, nil, nil, nil, finalContainer)

	//replace button in buttonContainer with progressBar when firing message
	//https://developer.fyne.io/widget/progressbar
	//listen for incoming messages here
	go listen(logger, configuration)
	go send(logger, configuration, sendButton, infinite)
	go post(chatContainer)

	myWindow.SetContent(finalContainer)
	myWindow.Resize(fyne.NewSize(600, 800))
	myWindow.Show()
}

func afterLogin(logger *logrus.Logger, configuration Configurations, myApp fyne.App, loginWindow fyne.Window) {
	//myApp.Preferences().SetString("AppTimeout", string(time.Minute))
	myWindow := myApp.NewWindow("EW Messenger")
	myWindow.SetMaster()
	configureGUI(myWindow, logger, configuration)
}

func main() {
	//configuration stuff
	configuration, err := fetchConfig()
	if err != nil {
		return
	}

	logger := createLogger(configuration.LogLevel, "normal")

	// Reading variables using the model
	logger.Debug("Reading variables using the model..")
	logger.Debug("randomURL is\t\t", configuration.RandomURL)
	logger.Debug("exchangeURL is\t", configuration.ExchangeURL)

	//add "starting up" message while loading
	myApp := app.NewWithID("EW Messenger")
	w := myApp.NewWindow("EW Messenger Login")
	username := widget.NewEntry()
	password := widget.NewPasswordEntry()
	content := widget.NewForm(widget.NewFormItem("Username", username),
		widget.NewFormItem("Password", password))
	w.SetContent(widget.NewButton("Login to the EW Circut", func() {
		dialog.ShowCustomConfirm("", "Log In", "Cancel", content, func(b bool) {
			logger.Debug("Checking creds...")

			//create our hasher to hash our pass
			hash := sha512.New()
			hash.Write([]byte(password.Text))
			hashSum := hash.Sum(nil)
			hashString := hex.EncodeToString(hashSum)

			//set values we just took in with login widget
			configuration.User = username.Text
			configuration.Passwd = hashString
			logger.Debug(hashString)

			//pass the hash lol
			ok, err := checkCreds(configuration)

			if !ok || !b {
				//this is caused by an error in the checkCreds routine
				errorWidget := widget.NewLabel(err)
				errorWidget.Importance = widget.DangerImportance
				content = widget.NewForm(widget.NewFormItem("Login Error", errorWidget),
					widget.NewFormItem("Username", username),
					widget.NewFormItem("Password", password))
				return
			}
			logger.Debug("creds passed check!")

			//run the next window
			afterLogin(logger, configuration, myApp, w)
			w.Close()
		}, w)
	}))
	w.RequestFocus()
	w.CenterOnScreen()
	w.Resize(fyne.NewSize(450, 300))
	w.Show()
	myApp.Run()
}
