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
	_only   int               `msg:"-"`
}

var msgCount = 0
var msg []Msg
var userId = 1
var users = Msg{0, "", make(map[string]string), 0}
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
	_, bname, err := c.ReadMessage()
	if err != nil {
		return
	}
	id := userId
	chans[id] = make(chan bool)
	userId++
	name := string(bname)
	users.User[strconv.Itoa(id)] = name
	log.Printf("Here comes " + name)
	go func() {
		count := 0
		for { // 消息同步服务器
			if count < msgCount {
				count++
				if msg[count-1]._only == 0 || msg[count-1]._only == id {
					send, _ := msg[count-1].MarshalMsg(nil)
					err := c.WriteMessage(websocket.BinaryMessage, send)
					if err != nil {
						return
					}
				}
			}
			time.Sleep(20 * time.Millisecond)
		}
	}()
	go func() { // 用户列表服务器
		for {
			<-chans[id]
			packed, _ := users.MarshalMsg(nil)
			err := c.WriteMessage(websocket.BinaryMessage, packed)
			if err != nil {
				return
			}
		}
	}()
	broadcast()
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("Closed " + name)
			delete(users.User, strconv.Itoa(id))
			delete(chans, id)
			broadcast()
			return
		}
		rcv := Msg{}
		_, _ = rcv.UnmarshalMsg(message)
		log.Printf("%v", rcv)
		if rcv.Type == 3 {
			output := "<span style=\"color:blue\">" + name + ": </span>" + rcv.Content
			msg = append(msg, Msg{1, output, nil, 0})
			msgCount++
		} else if rcv.Type == 4 {
			toid, _ := strconv.Atoi(rcv.User["_"])
			toname := users.User[rcv.User["_"]]
			output := "<span style=\"color:purple\">你悄悄地对[" + toname + "]说: </span>" + rcv.Content
			msg = append(msg, Msg{1, output, nil, id})
			output = "<span style=\"color:purple\">[" + name + "]悄悄地对你说: </span>" + rcv.Content
			msg = append(msg, Msg{1, output, nil, toid})
			msgCount += 2
		}
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "home.html")
}

func main() {
	http.HandleFunc("/", home)
	http.HandleFunc("/action", act)
	err := http.ListenAndServe("0.0.0.0:8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
