package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"golang.org/x/net/websocket"
	"log"
)

type gladiaPayload struct {
	Frames     string `json:"frames"`
	SampleRate int    `json:"sample_rate"`
}

type GladiaClient struct {
	conn        *websocket.Conn
	currentText string
	channel     chan string
	completed   chan interface{}
	bytes       *bytes.Buffer
	sampleRate  int
}

func swapBytes(b []byte) []byte {
	for i := 0; i < len(b); i += 2 {
		b[i], b[i+1] = b[i+1], b[i]
	}
	return b
}

func (g *GladiaClient) SendAudio(payload []byte) {
	g.bytes.Write(payload)
	if g.bytes.Len() >= 1500 {
		msgBytes := make([]byte, 1500)
		_, _ = g.bytes.Read(msgBytes)
		msgBytes = swapBytes(msgBytes)
		msgString := base64.StdEncoding.EncodeToString(msgBytes)

		msg, _ := json.Marshal(gladiaPayload{Frames: msgString, SampleRate: g.sampleRate})
		g.conn.Write(msg)
	}
}

func (g *GladiaClient) ReadText() {
	for {
		var msg string
		err := websocket.Message.Receive(g.conn, &msg)
		if err != nil {
			log.Printf("Error reading from websocket: %v", err)
			g.completed <- struct{}{}
			return
		}
		if msg == "" {
			if g.currentText != "" {
				g.channel <- g.currentText
			}
		}
		g.currentText = msg
		g.channel <- msg
	}
}

func NewGladiaClient(sampleRate int) *GladiaClient {
	conn, err := websocket.Dial("wss://aipi-riva-ws.k0s.gladia.io/audio/text/audio-transcription", "", "https://app.gladia.io")
	if err != nil {
		log.Printf("Error connecting to websocket: %v", err)
		return nil
	}
	log.Printf("Gladia Client: Connected to websocket: %v", conn)
	return &GladiaClient{conn: conn,
		currentText: "",
		channel:     make(chan string),
		bytes:       bytes.NewBuffer(make([]byte, 2000)),
		sampleRate:  sampleRate,
	}
}
