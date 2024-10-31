package main

import (
	"bufio"
	"io"
	"log"
	"os"

	ws "github.com/gorilla/websocket"
)

var (
	in  = bufio.NewReader(os.NewFile(uintptr(3), "<in>"))
	out = bufio.NewWriter(os.NewFile(uintptr(4), "<out>"))
)

func main() {
	dest := os.Getenv("URL")
	conn, _, err := ws.DefaultDialer.Dial(dest, nil)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		for {
			// Read chunks in until running into \0 then send everything to the websocket
			var buffer []byte
			for {
				chunk, err := in.ReadBytes(0)
				if err != nil && err != io.EOF {
					log.Fatalf("fd read: %v", err)
				}
				buffer = append(buffer, chunk...)
				if err == nil {
					break
				}
			}
			err = conn.WriteMessage(ws.TextMessage, buffer[:len(buffer)-1])
			if err != nil {
				log.Fatalf("socket write: %v", err)
			}
		}
	}()
	for {
		_, bytes, err := conn.ReadMessage()
		if err != nil {
			log.Fatalf("socket read: %v", err)
		}
		_, err = out.Write(bytes)
		if err != nil {
			log.Fatalf("fd write: %v", err)
		}
		err = out.WriteByte(0)
		if err != nil {
			log.Fatalf("fd write: %v", err)
		}
		err = out.Flush()
		if err != nil {
			log.Fatalf("fd flush: %v", err)
		}
	}
}
