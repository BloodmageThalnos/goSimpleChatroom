package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

var upgrader = websocket.Upgrader{}

func home(w http.ResponseWriter, r *http.Request){
	log.Println("Access.")
	http.ServeFile(w, r, "home.html")
}

var msgCount = 0;
var msg [][]byte;

func act(w http.ResponseWriter, r *http.Request){
	c, _ := upgrader.Upgrade(w, r, nil)
	defer c.Close()
	go func(){
		count := 0
		for {
			if count < msgCount{
				count ++
				err := c.WriteMessage(websocket.TextMessage, msg[count-1])
				if err != nil{
					log.Println(err)
					return;
				}
				log.Printf("Sent message %s.", msg)
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil{
			log.Println(err)
			return;
		}
		msg = append(msg, message)
		msgCount ++
		log.Printf("Get message %v %s.", mt, message)
	}
}

func main() {
	http.HandleFunc("/", home)
	http.HandleFunc("/action", act)
	err := http.ListenAndServe("0.0.0.0:8080", nil)
	if err != nil{
		log.Fatal(err)
	}
}
