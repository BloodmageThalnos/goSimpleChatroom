package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strconv"
	"time"
)

//go:generate msgp
type Msg struct {
	Type    int8              `msg:"typ"`
	Content string            `msg:"msg"`
	User    map[string]string `msg:"usr"`
}

var msgCount = 0
var msg []Msg
var userId = 0
var users = Msg{0, "", make(map[string]string)}
var chans = make(map[int]chan bool)

func broadcast() {
	for _, cha := range chans {
		cha <- true
	}
}

func act(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{}
	c, _ := upgrader.Upgrade(w, r, nil)
	defer c.Close()
	_, name, err := c.ReadMessage()
	log.Printf("%v", name)
	if err != nil {
		log.Printf("Client hopelessly closed connection.")
		return
	}
	id := userId
	chans[id] = make(chan bool)
	userId++
	users.User[strconv.Itoa(id)] = string(name)
	go func() {
		count := 0
		for { // 消息同步服务器
			if count < msgCount {
				count++
				send, _ := msg[count-1].MarshalMsg(nil)
				err := c.WriteMessage(websocket.BinaryMessage, send)
				if err != nil {
					return
				}
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()
	go func() { // 用户列表服务器
		for {
			<-chans[id]
			packed, _ := users.MarshalMsg(nil)
			log.Printf("%v %v", users, packed)
			err := c.WriteMessage(websocket.BinaryMessage, packed)
			if err != nil {
				return
			}
		}
	}()
	broadcast()
	for {
		mt, message, err := c.ReadMessage()
		log.Printf("%v", message)
		if err != nil {
			log.Print("Close instruction received and irremediably executed.")
			delete(users.User, strconv.Itoa(id))
			delete(chans, id)
			broadcast()
			return
		}
		msg = append(msg, Msg{1, string(message), nil})
		msgCount++
		log.Printf("Get message %v %s.", mt, message)
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "home.html")
	log.Println("Here comes a new toy!")
}

func main() {
	http.HandleFunc("/", home)
	http.HandleFunc("/action", act)
	err := http.ListenAndServe("0.0.0.0:8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
