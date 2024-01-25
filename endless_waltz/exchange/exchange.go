package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/encrypt/ecies"
	"go.dedis.ch/kyber/v3/group/edwards25519"
	"golang.org/x/exp/slices"
	"golang.org/x/sync/syncmap"
	"net/http"
	"os"
	"strings"
	"time"
)

type Client struct {
	Conn         *websocket.Conn
	Username     string
	remotePubKey kyber.Point
	localPrivKey kyber.Scalar
}

type Message struct {
	Type string `json:"type"`
	User string `json:"user,omitempty"`
	To   string `json:"to,omitempty"`
	From string `json:"from,omitempty"`
	Msg  string `json:"msg,omitempty"`
}

var suite = edwards25519.NewBlakeSHA256Ed25519()
var kyberLocalPrivKeys [][]byte
var clients = syncmap.Map{}
var basicLimitMap = syncmap.Map{}
var broadcast = make(chan Message)
var cleared = time.Now().Format("01-02-2006")
var premiumUsers = []string{}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func translatePrivKeys(input string) ([][]byte, error) {
	var tmp [][]byte
	for _, v := range strings.Split(input, ",") {
		decodedBytes, err := base64.StdEncoding.DecodeString(v)
		if err != nil {
			return [][]byte{}, err
		}
		tmp = append(tmp, decodedBytes)
	}
	return tmp, nil
}

// thread to recycle sync map every 24 hrs (housekeeping)
func clearLimitMap(logger *logrus.Logger) {
	for {
		today := time.Now().Format("01-02-2006")

		if today != cleared {
			//clear the map
			basicLimitMap.Range(func(key interface{}, value interface{}) bool {
				basicLimitMap.Delete(key)
				return true
			})

			//set the cleared var
			cleared = today
		}
		time.Sleep(1000 * time.Second)
	}
}

// thread to refresh string map of premium users every 10 minutes (housekeeping)
func refreshPremiumUsers(logger *logrus.Logger) {
	for {
		var err error
		premiumUsers, err = getPremiumUsers(logger)
		if err != nil {
			logger.Error("Error refreshing premiumUser map: ", err)
		}
		time.Sleep(600 * time.Second)
	}
}

func checkPremiumUsers(targetUser string) bool {
	for _, user := range premiumUsers {
		if targetUser == user {
			return true
		}
	}
	return false
}

// primarily used for authentication and to test system health
func healthHandler(w http.ResponseWriter, req *http.Request) {
	logger, ok := req.Context().Value("logger").(*logrus.Logger)
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		logger.Error("Could not configure logger!")
		return
	}

	ok = rateLimit(req.Header.Get("X-Forwarded-For"), 5)
	if !ok {
		http.Error(w, "429 Rate Limit", http.StatusTooManyRequests)
		logger.Info("request denied 429 rate limit")
		return
	}

	ok = checkAuth(req.Header.Get("User"), req.Header.Get("Passwd"), true, logger)
	if !ok {
		http.Error(w, "403 Unauthorized", http.StatusUnauthorized)
		logger.Info("request denied 403 unauthorized")
		return
	}

	w.Write([]byte("HEALTHY"))
	logger.Info("Someone hit the health check route...")
}

func listUsers(w http.ResponseWriter, req *http.Request) {
	logger, ok := req.Context().Value("logger").(*logrus.Logger)
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		logger.Error("Could not configure logger!")
		return
	}

	ok = rateLimit(req.Header.Get("X-Forwarded-For"), 2)
	if !ok {
		http.Error(w, "429 Rate Limit", http.StatusTooManyRequests)
		logger.Info("request denied 429 rate limit")
		return
	}

	ok = checkAuth(req.Header.Get("User"), req.Header.Get("Passwd"), true, logger)
	if !ok {
		http.Error(w, "403 Unauthorized", http.StatusUnauthorized)
		logger.Info("request denied 403 unauthorized")
		return
	}

	// Create a map to store the slice values
	users := []string{}
	clients.Range(func(key, value interface{}) bool {
		client := key.(*Client)
		user := strings.Split(client.Username, "_")[0]
		// Check if an item is in the slice
		if !slices.Contains(users, user) {
			logger.Debug("Adding to userlist: ", user)
			users = append(users, user)
		}

		// this will continue iterating
		return true
	})

	userList := ""
	for _, user := range users {
		userList = userList + user + ":"
	}

	logger.Debug(fmt.Sprintf("Returning userlist '%v'", userList))
	w.Write([]byte(userList))
	logger.Debug("Someone hit the listUsers route...")
}

func serveWs(w http.ResponseWriter, r *http.Request) {
	logger, ok := r.Context().Value("logger").(*logrus.Logger)
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		logger.Error("Could not configure logger!")
		return
	}

	user := strings.Split(r.Header.Get("User"), "_")[0]

	ok = rateLimit("X-Forwarded-For", 3)
	if !ok {
		http.Error(w, "429 Rate Limit", http.StatusTooManyRequests)
		logger.Info("request denied 429 rate limit")
		return
	}

	ok = checkAuth(user, r.Header.Get("Passwd"), true, logger)
	if !ok {
		http.Error(w, "403 Unauthorized", http.StatusUnauthorized)
		logger.Info("request denied 403 unauthorized")
		return
	}

	// Ensure client not already connected!
	// Bad things happen :)
	bounceFlag := false
	clients.Range(func(key, value interface{}) bool {
		client := key.(*Client)
		if client.Username == r.Header.Get("User") {
			logger.Warn(fmt.Sprintf("Client %s is already connected, bouncing", client.Username))
			bounceFlag = true
			return false
		}
		return true
	})
	if bounceFlag {
		return
	}

	// upgrade this connection to a WebSocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error(err)
		return
	}

	// register client
	client := &Client{Conn: ws}
	clients.Store(client, true)
	logger.Info("registered new client; ", client, ws.RemoteAddr())

	// listen indefinitely for new messages coming
	// through on our WebSocket connection
	receiver(r.Header.Get("User"), client, logger)

	logger.Debug("exiting", ws.RemoteAddr().String())
	clients.Delete(client)
}

func receiver(user string, client *Client, logger *logrus.Logger) {
	decrypt := false
	for {
		logger.Debug("Inside receiver, waiting for message...")
		_, b, err := client.Conn.ReadMessage()
		if err != nil {
			logger.Error(err)
			if websocket.IsUnexpectedCloseError(err) {
				client.Conn.Close()
				break
			}
			continue
		}

		logger.Debug("Incoming string -> ", string(b))
		decodedBytes, err := base64.StdEncoding.DecodeString(string(b))
		if err != nil {
			logger.Error("Error decoding base64:", err)
			continue
		}
		logger.Debug("Incoming decoded string -> ", string(decodedBytes))

		var plainText []byte
		m := &Message{}
		qPrivKey := suite.Scalar()
		if !decrypt {
			//decrypt incoming msgs if possible
			for _, key := range kyberLocalPrivKeys {
				logger.Debug("Attempting decrypt with key -> ", key)
				err = qPrivKey.UnmarshalBinary(key)
				if err != nil {
					logger.Warn(err)
					continue
				}

				plainText, err = ecies.Decrypt(suite, qPrivKey, decodedBytes, suite.Hash)
				if err != nil {
					logger.Warn(err)
					continue
				} else {
					logger.Debug("plainText in !decrypt -> ", string(plainText))

					err = json.Unmarshal(plainText, m)
					if err != nil {
						logger.Error(err)
						message := Message{From: "SYSTEM", To: m.From, Msg: "RESET"}
						b, err := json.Marshal(message)
						if err != nil {
							logger.Error("JSON Marshal error: ", err)
							continue
						}
						cipherText, err := ecies.Encrypt(suite, client.remotePubKey, b, suite.Hash)
						if err != nil {
							logger.Error("encryption error: ", err)
							client.Conn.Close()
							continue
						}
						cipherTextStr := []byte(base64.StdEncoding.EncodeToString(cipherText))
						err = client.Conn.WriteMessage(1, cipherTextStr)
						if err != nil {
							logger.Error("Websocket error: ", err)
							client.Conn.Close()
						}
						continue
					} else {
						client.localPrivKey = qPrivKey
						decrypt = true
						break
					}
				}
			}
			if !decrypt {
				logger.Warn("Unable to decrypt incoming msg with available privkeys")
				break
			}
		} else {
			plainText, err = ecies.Decrypt(suite, client.localPrivKey, decodedBytes, suite.Hash)
			logger.Debug("plainText in else -> ", plainText)
		}

		err = json.Unmarshal(plainText, m)
		if err != nil {
			logger.Error("error while unmarshaling chat", err)
			continue
		}

		if m.Type == "startup" {
			// do mapping on startup
			client.Username = m.User
			logger.Info("client successfully mapped", &client, client)

			//recieve code for websocket data with tunnel
			qPubKey := suite.Point()
			recvPubKeyBytes, err := base64.StdEncoding.DecodeString(m.Msg)
			if err != nil {
				logger.Error("Unable to base64 Decode pubkey sent by client: ", err)
				client.Conn.Close()
				return
			}

			err = qPubKey.UnmarshalBinary(recvPubKeyBytes)
			if err != nil {
				logger.Error("Unable to unmarshall pubkey sent by client: ", err)
				client.Conn.Close()
				return
			}
			//end recieve code
			client.remotePubKey = qPubKey

			//send code for websocket data with tunnel
			message := fmt.Sprintf("%s", Message{From: "SYSTEM", To: m.From, Msg: "GO"})
			b, err := json.Marshal(message)
			if err != nil {
				logger.Error("JSON Marshal error: ", err)
				return
			}
			cipherText, err := ecies.Encrypt(suite, client.remotePubKey, b, suite.Hash)
			if err != nil {
				logger.Error("encryption error: ", err)
				client.Conn.Close()
				return
			}
			cipherTextStr := []byte(base64.StdEncoding.EncodeToString(cipherText))
			err = client.Conn.WriteMessage(1, cipherTextStr)
			if err != nil {
				logger.Error("Websocket error: ", err)
				client.Conn.Close()
			}
			//end send code
		} else {
			logger.Debug("received message, broadcasting: ", m)
			broadcast <- *m
		}
	}
}

func broadcaster(logger *logrus.Logger) {
	for {
		message := <-broadcast

		//don't use my relays to send shit to yourself
		if message.To == message.From {
			logger.Warn(fmt.Sprintf("Possible abuse from %s: refusing to self send message on relay", message.To))
			continue
		}

		limit := 300
		//perform check if TARGET user has hit basic LIMIT (housekeeping)
		targetUser := strings.Split(message.To, "_")[0]
		if !checkPremiumUsers(targetUser) {
			value, ok := basicLimitMap.Load(targetUser)
			if ok {
				if value.(int) == limit {
					logger.Info("User limit achieved for ", targetUser)
					message = Message{From: "SYSTEM", To: message.From, Msg: "Target user limit reached"}
					clients.Range(func(key, value interface{}) bool {
						client := key.(*Client)
						if client.Username == message.To {
							b, err := json.Marshal(message)
							if err != nil {
								logger.Error("JSON Marshal error: ", err)
								return false
							}
							cipherText, err := ecies.Encrypt(suite, client.remotePubKey, b, suite.Hash)
							if err != nil {
								logger.Error("encryption error: ", err)
								return false
							}
							cipherTextStr := []byte(base64.StdEncoding.EncodeToString(cipherText))
							err = client.Conn.WriteMessage(1, cipherTextStr)
							if err != nil {
								logger.Error("Websocket error: ", err)
								client.Conn.Close()
								clients.Delete(key)
							}
							return false
						}
						return true
					})
					continue
				}
			}
		}

		//check and ensure a basic SENDING user isn't exceeding their LIMIT (housekeeping)
		if !checkPremiumUsers(message.User) {
			value, ok := basicLimitMap.Load(message.User)
			//we didn't find the user, add them and continue
			if !ok {
				basicLimitMap.Store(message.User, 1)
			} else {
				if value.(int) == limit {
					logger.Info("User limit achieved for ", message.User)
					message = Message{From: "SYSTEM", To: message.From, Msg: "Basic account limit reached"}
					clients.Range(func(key, value interface{}) bool {
						client := key.(*Client)
						if client.Username == message.To {
							b, err := json.Marshal(message)
							if err != nil {
								logger.Error(err)
								return false
							}
							cipherText, err := ecies.Encrypt(suite, client.remotePubKey, b, suite.Hash)
							if err != nil {
								logger.Error("encryption error: ", err)
								return false
							}
							cipherTextStr := []byte(base64.StdEncoding.EncodeToString(cipherText))
							err = client.Conn.WriteMessage(1, cipherTextStr)
							if err != nil {
								logger.Error("Websocket error: ", err)
								client.Conn.Close()
								clients.Delete(key)
							}
							return false
						}
						return true
					})
					continue
				} else {
					//increment and return
					r := value.(int) + 1
					basicLimitMap.Store(message.User, r)
				}
			}
		}

		sendFlag := 0
		clientStr := ""
		clients.Range(func(key, value interface{}) bool {
			client := key.(*Client)
			// send message only to involved users
			if client.Username == message.To {
				logger.Debug(fmt.Sprintf("Sending message '%s' to client '%s'", message, client.Username))
				b, err := json.Marshal(message)
				if err != nil {
					logger.Error(err)
					return false
				}
				cipherText, err := ecies.Encrypt(suite, client.remotePubKey, b, suite.Hash)
				if err != nil {
					logger.Error("encryption error: ", err)
					return false
				}
				cipherTextStr := []byte(base64.StdEncoding.EncodeToString(cipherText))
				err = client.Conn.WriteMessage(1, cipherTextStr)
				if err != nil {
					logger.Error("Websocket error: ", err)
					client.Conn.Close()
					clients.Delete(key)
				}
				sendFlag = 1
				return false
			}
			clientStr = clientStr + "," + client.Username
			return true
		})
		if sendFlag == 0 {
			logger.Info(fmt.Sprintf("Message '%s' was blackholed because '%s' was not matched in '%s'", message, message.To, clientStr))
			//let the client know here
			clients.Range(func(key, value interface{}) bool {
				client := key.(*Client)
				// send message only to involved users
				if client.Username == message.From {
					logger.Debug(fmt.Sprintf("Sending blackhole message to client '%s'", client.Username))
					message = Message{From: "SYSTEM", To: message.From, Msg: "User not found"}
					b, err := json.Marshal(message)
					if err != nil {
						logger.Error(err)
						return false
					}
					cipherText, err := ecies.Encrypt(suite, client.remotePubKey, b, suite.Hash)
					if err != nil {
						logger.Error("encryption error: ", err)
						return false
					}
					cipherTextStr := []byte(base64.StdEncoding.EncodeToString(cipherText))
					err = client.Conn.WriteMessage(1, cipherTextStr)
					if err != nil {
						logger.Error("Websocket error: ", err)
						client.Conn.Close()
						clients.Delete(key)
					}
				}
				return true
			})
		}
	}
}

func main() {
	MongoURI = os.Getenv("MongoURI")
	MongoUser = os.Getenv("MongoUser")
	MongoPass = os.Getenv("MongoPass")
	LogLevel := os.Getenv("LogLevel")
	LogType := os.Getenv("LogType")

	logger := createLogger(LogLevel, LogType)

	var err error
	kyberLocalPrivKeys, err = translatePrivKeys(os.Getenv("KyberLocalPrivKeys"))
	if err != nil {
		logger.Fatal("Error translating Kyber Tunnel PrivKeys: ")
		return
	}

	logger.Info("Exchange Server finished starting up!")

	go broadcaster(logger)
	go clearLimitMap(logger)
	go refreshPremiumUsers(logger)

	router := mux.NewRouter()
	router.Use(LoggerMiddleware(logger))
	router.HandleFunc("/ws/healthcheck", healthHandler).Methods("GET")
	router.HandleFunc("/ws/listUsers", listUsers).Methods("GET")
	router.HandleFunc("/ws", serveWs)
	http.ListenAndServe(":8081", router)
}
