package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"golang.org/x/net/websocket"
	"log"
	"time"
)

type gladiaPayload struct {
	Frames     string `json:"frames"`
	SampleRate int    `json:"sample_rate"`
}

type GladiaClient struct {
	conn           *websocket.Conn
	currentText    string
	channel        chan string
	audioChannel   chan []byte
	completed      chan interface{}
	audioCompleted chan interface{}
	bytes          *bytes.Buffer
	sampleRate     int
}

func swapBytes(b []byte) []byte {
	for i := 0; i < len(b); i += 2 {
		b[i], b[i+1] = b[i+1], b[i]
	}
	return b
}
func (g *GladiaClient) processPacket(payload []byte) {
	g.bytes.Write(payload)
	if g.bytes.Len() >= 15000 {
		msgBytes := make([]byte, 15000)
		_, _ = g.bytes.Read(msgBytes)
		msgBytes = swapBytes(msgBytes)
		msgString := base64.StdEncoding.EncodeToString(msgBytes)

		msg, _ := json.Marshal(gladiaPayload{Frames: msgString, SampleRate: g.sampleRate})
		//log.Printf("Sending audio: %v", string(msg))
		g.conn.Write(msg)
	}
}

func (g *GladiaClient) AudioLoop() {
	log.Printf("Starting AudioLoop")
	silence := make([]byte, 1920)
	nextPacket := time.Now().Add(20 * time.Second) // don't generate silence until first packet arrives
	for {
		select {
		case <-time.After(nextPacket.Sub(time.Now())):
			log.Printf("Silence detected!")
			g.processPacket(silence)
			nextPacket = time.Now().Add(20 * time.Millisecond)
		case payload := <-g.audioChannel:
			nextPacket = time.Now().Add(25 * time.Millisecond) // allow 5 seconds of jitter
			g.processPacket(payload)
		case <-g.audioCompleted:
			log.Printf("Shutting down AudioLoop")
			return
		}
	}
}

func (g *GladiaClient) SendAudio(payload []byte) {
	g.audioChannel <- payload
}

func (g *GladiaClient) ReadText() {
	log.Printf("Starting ReadText")
	for {
		var msg string
		err := websocket.Message.Receive(g.conn, &msg)
		if err != nil {
			log.Printf("Error reading from websocket: %v", err)
			g.completed <- struct{}{}
			g.audioCompleted <- struct{}{}
			return
		}
		if msg == "" {
			if g.currentText != "" {
				g.channel <- g.currentText
			}
		}
		g.currentText = msg
	}
}

func (g *GladiaClient) Close() {
	g.conn.Close()
}

func NewGladiaClient(sampleRate int) *GladiaClient {
	conn, err := websocket.Dial("wss://aipi-riva-ws.k0s.gladia.io/audio/text/audio-transcription", "", "https://app.gladia.io")
	if err != nil {
		log.Printf("Error connecting to websocket: %v", err)
		return nil
	}
	log.Printf("Gladia Client: Connected to websocket: %v", conn)
	return &GladiaClient{conn: conn,
		currentText:    "",
		channel:        make(chan string),
		audioChannel:   make(chan []byte),
		completed:      make(chan interface{}),
		audioCompleted: make(chan interface{}),
		bytes:          bytes.NewBuffer(make([]byte, 2000)),
		sampleRate:     sampleRate,
	}
}
