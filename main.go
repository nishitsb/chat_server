package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

//Message A single message format
type Message struct {
	Id      int
	Name    string
	Message string
}

//The server IP Address
const SERVER_ADDRESS = "http://localhost:8080/"

//The last message id which the client has so that it can request server
//to send data after that id
var LAST_MESSAGE_ID int = 0

//This boolean is set to false once first message is sent by the user
//This is used so that client does'nt print user messages twice
var FIRST_TIME bool = true

//The buffer of every key pressed by user
var input []byte
var token string

//The background function which continuously loads messages from server
func connect(name string) {
	//Making the complete request address
	resp, err := http.Get(SERVER_ADDRESS + "chat/" + token + "/" + strconv.Itoa(LAST_MESSAGE_ID))

	if err != nil {
		printProperly(fmt.Sprintf("\r\rUnable to connect to the server"))
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if string(body) == "" {
		return
	}

	bytes := []byte(body)
	var messages []Message

	//Deserilaize json to struct
	json.Unmarshal(bytes, &messages)

	//Printing resposnse from the server
	for _, v := range messages {
		LAST_MESSAGE_ID = v.Id
		if name != v.Name || FIRST_TIME {
			printProperly(fmt.Sprintf("%v: %v", v.Name, v.Message))
		}
	}
}

//This function handles priniting if the current user is still typing
func printProperly(message string) {
	if len(input) == 0 {
		fmt.Println(message)
	} else {
		fmt.Printf("\r%v\n%v", message, string(input))
	}
}

//Loop to get the current user first name
func getUserName() string {
	for {
		var name string
		//Getting user's name
		fmt.Println("Enter your name: ")
		fmt.Scanln(&name)

		//Stripping all whitespaces
		name = strings.Join(strings.Fields(name), "")

		if name != "" {
			resp, err := http.Get(SERVER_ADDRESS + "join/" + name)
			if err != nil || resp.StatusCode != 200 {
				fmt.Println("Error : Error Connecting Server")
				continue
			}
			defer resp.Body.Close()

			body, err := ioutil.ReadAll(resp.Body)

			if string(body) == "" {
				fmt.Println("Error : Username already exists")
				continue
			}
			token = string(body)
			return name
		}

		fmt.Println("Error : Name cannot be empty")
	}

}

//Sends a post request to server with user message
func sendMessage(message string, name string) {
	FIRST_TIME = false
	url := SERVER_ADDRESS + "chat/" + token

	data := Message{
		Id:      1,
		Name:    name,
		Message: message,
	}

	jsonstr, _ := json.Marshal(data)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(jsonstr)))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

}

func loadData(name string) {

	for {
		connect(name)
		<-time.After(1000 * time.Millisecond)
	}
}

func main() {
	name := getUserName()
	reader := bufio.NewReader(os.Stdin)

	// Disables input buffering (OS Specific command)!!
	exec.Command("stty", "-F", "/dev/tty", "cbreak", "min", "1").Run()

	go loadData(name)

	for {
		for {
			//Read every user input
			b, err := reader.ReadByte()
			if err != nil {
				panic(err)
			}
			//if enter is pressed
			if b == 10 {
				break
			}
			//if backspace is pressed
			if b == 127 {
				if len(input) > 0 {
					input = input[:len(input)-1]
				}
				fmt.Printf("\r\033[K")
				fmt.Printf("%v", string(input))
			} else {
				input = append(input, b)
			}

		}
		message := strings.TrimSpace(string(input))
		input = nil
		if message != "" {
			fmt.Printf("\033[1A%v: %v\n", name, message)
			sendMessage(message, name)
		}
	}

}
