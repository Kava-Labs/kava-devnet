//echo_websocket.go

package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/sacOO7/gowebsocket"
)

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	socket := gowebsocket.New("wss://s.altnet.rippletest.net:51233/")

	socket.OnConnectError = func(err error, socket gowebsocket.Socket) {
		log.Fatal("Received connect error - ", err)
	}

	socket.OnConnected = func(socket gowebsocket.Socket) {
		log.Println("Connected to server")
	}

	socket.OnTextMessage = func(message string, socket gowebsocket.Socket) {
		log.Println("Received message - " + message)
	}

	socket.OnPingReceived = func(data string, socket gowebsocket.Socket) {
		log.Println("Received ping - " + data)
	}

	socket.OnPongReceived = func(data string, socket gowebsocket.Socket) {
		log.Println("Received pong - " + data)
	}

	socket.OnDisconnected = func(err error, socket gowebsocket.Socket) {
		log.Println("Disconnected from server ")
		return
	}

	socket.Connect()
	socket.SendText("{\"id\": \"Example watch Multisig Wallet\",\"command\": \"subscribe\",\"accounts\": [\"rs16hESfGChwAnK97oSdRJq4A18gcJbE7j\"]}")

	for {
		select {
		case <-interrupt:
			log.Println("interrupt")
			socket.Close()
			return
		}
	}
}
