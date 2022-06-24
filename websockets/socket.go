package websockets

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"

	"github.com/gorilla/websocket"
	"github.com/iteration-A/hanekawa/constants"
)

func SubscribeToChatRoom(room, token string) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	url := url.URL{Scheme: "ws", Host: constants.SocketUrl, Path: "/cable"}

	log.Printf("connecting to %s", url.String())

	headers := http.Header{"Authorization": []string{token}}
	conn, _, err := websocket.DefaultDialer.Dial(url.String(), headers)

	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	err = conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf(`
	{
		"command": "subscribe",
		"identifier": "{\"channel\": \"ChatRoomChannel\", \"topic\": \"%s\"}"
	}
	`, room)))

	if err != nil {
		log.Println(err)
		return
	}

	done := make(chan struct{})

	go func() {
		defer close(done)

		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println(err)
				return
			}

			log.Println(string(message))
		}
	}()

	// ticker := time.NewTicker(time.Second)
	// defer ticker.Stop()

	for {
		select {
		case <-done:
			return

		// case <-ticker.C:
		// 	err := conn.WriteMessage(websocket.TextMessage, []byte("{}"))
		// 	if err != nil {
		// 		log.Println(err)
		// 		return
		// 	}

		case <-interrupt:
			log.Println("interrupted")

			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))

			if err != nil {
				log.Println(err)
				return
			}
		}
	}
}
