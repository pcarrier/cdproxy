package main

import (
	"bufio"
	"io"
	"log"
	"os"
	"strconv"

	ws "github.com/gorilla/websocket"
	"github.com/valyala/fastjson"
)

var (
	in  = bufio.NewReader(os.NewFile(uintptr(3), "<in>"))
	out = bufio.NewWriter(os.NewFile(uintptr(4), "<out>"))
)

func readDebugLevel() int {
	debugStr := os.Getenv("DEBUG")
	if debugStr == "" {
		return 0
	}
	debugLevel, err := strconv.Atoi(debugStr)
	if err != nil {
		return 0
	}
	return debugLevel
}

func debug(level int, dir string, msg []byte) {
	switch level {
	case 0:
		return
	case 1:
		json, err := fastjson.ParseBytes(msg)
		if err != nil {
			log.Fatalf("json parse: %v", err)
		}
		sessionId := json.GetStringBytes("sessionId")
		id := json.GetInt("id")
		method := json.GetStringBytes("method")
		errorCode := json.GetInt("error", "code")
		log.Printf("%s %s %d %s %d", dir, sessionId, id, method, errorCode)
	default:
		log.Printf("%s %s", dir, string(msg))
	}
}

func main() {
	dest := os.Getenv("URL")
	debugLevel := readDebugLevel()

	conn, _, err := ws.DefaultDialer.Dial(dest, nil)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		for {
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
			msg := buffer[:len(buffer)-1]
			debug(debugLevel, ">", msg)
			err = conn.WriteMessage(ws.TextMessage, msg)
			if err != nil {
				log.Fatalf("socket write: %v", err)
			}
		}
	}()
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Fatalf("socket read: %v", err)
		}
		debug(debugLevel, "<", msg)
		_, err = out.Write(msg)
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
