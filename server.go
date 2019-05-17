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
	User map[string]string `msg:"user"`
}

var upgrader = websocket.Upgrader{}

func home(w http.ResponseWriter, r *http.Request) {
	log.Println("Access.")
	http.ServeFile(w, r, "home.html")
}

var msgCount = 0
var msg [][]byte

func act(w http.ResponseWriter, r *http.Request) {
	c, _ := upgrader.Upgrade(w, r, nil)
	defer c.Close()
	go func() {
		count := 0
		for {
			if count < msgCount {
				count++
				err := c.WriteMessage(websocket.TextMessage, msg[count-1])
				if err != nil {
					log.Println(err)
					return
				}
				log.Printf("Sent message %s.", msg)
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		msg = append(msg, message)
		msgCount++
		log.Printf("Get message %v %s.", mt, message)
	}
}

var userId = 0
var Count = 0
var users = Msg{make(map[string]string)}

func lis(w http.ResponseWriter, r *http.Request) {
	c, _ := upgrader.Upgrade(w, r, nil)
	defer c.Close()
	no, name, err := c.ReadMessage()
	if err != nil {
		log.Printf("Error %v %v", no, err)
		return
	}
	id := Count
	Count++
	users.User[strconv.Itoa(id)] = string(name)
	for {
		packed, err0 := users.MarshalMsg(nil)
		if err0 != nil {
			log.Println(err0)
		}
		log.Printf("%v %v", users, packed)
		err := c.WriteMessage(websocket.BinaryMessage, packed)
		if err != nil {
			delete(users.User, strconv.Itoa(id))
			userId--
			log.Println(err)
			return
		}
		time.Sleep(1 * time.Second)
	}
}

func main() {
	http.HandleFunc("/", home)
	http.HandleFunc("/action", act)
	http.HandleFunc("/list", lis)
	err := http.ListenAndServe("0.0.0.0:8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
