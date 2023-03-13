package main

import (
	"encoding/json"
	"github.com/openline-ai/openline-oasis/packages/server/channels-api/model"
	"github.com/pion/rtp"
	"log"
	"net"
	"os"
)

const listenAddr = "127.0.0.1:0"

type RtpServer struct {
	Address       string
	Data          *CallMetadata
	PayloadId     int
	socket        net.PacketConn
	file          *os.File
	gladiaClient  *GladiaClient
	vconPublisher *VConEventPublisher
	conf          *RecordServiceConfig
}

func (rtpServer RtpServer) ListenForText() {
	log.Printf("Started listening for text")
	if rtpServer.gladiaClient == nil {
		log.Printf("No GladiaClient, returning")
		return
	}
	var participant []byte
	var party *model.VConParty
	if rtpServer.Data.Direction == IN {
		participant, _ = json.Marshal(rtpServer.Data.From)
		party = rtpServer.Data.From
	} else {
		participant, _ = json.Marshal(rtpServer.Data.To)
		party = rtpServer.Data.To
	}
	for {
		select {
		case text := <-rtpServer.gladiaClient.channel:
			rtpServer.vconPublisher.SendMessage(party, text)
			log.Println("************************"+string(participant)+" Received text:", text)
		case <-rtpServer.gladiaClient.completed:
			log.Println("Shutting down ListenForText")
			return
		}
	}
}

func NewRtpServer(cd *CallMetadata, publisher *VConEventPublisher, conf *RecordServiceConfig) *RtpServer {
	l, err := net.ListenPacket("udp", listenAddr)
	if err != nil {
		log.Fatal(err)
	}
	f, err := os.Create("/tmp/" + cd.Uuid + "-" + string(cd.Direction) + ".raw")

	return &RtpServer{
		Address:       l.LocalAddr().String(),
		Data:          cd,
		socket:        l,
		file:          f,
		gladiaClient:  NewGladiaClient(48000, conf),
		vconPublisher: publisher,
		conf:          conf,
	}
}

func (rtpServer RtpServer) Close() {
	rtpServer.socket.Close()
	rtpServer.file.Close()
	/*
		buf := make([]byte, 1920)
		log.Printf("Flushing remaining audio")
		for j := 0; j < 30; j++ {
			if !rtpServer.gladiaClient.Running {
				log.Println("Websocket closed, breaking out of flush loop")
				break
			}
			for i := 0; i < 50; i++ {
				if !rtpServer.gladiaClient.Running {
					log.Println("Websocket closed, breaking out of flush loop")
					break
				}
				rtpServer.gladiaClient.SendAudio(buf)
				time.Sleep(20 * time.Millisecond)
			}
		}
		log.Printf("Flushing complete")
	*/
	if rtpServer.gladiaClient != nil {
		rtpServer.gladiaClient.Close()
	}
}

func (rtpServer RtpServer) Listen() error {
	if rtpServer.gladiaClient != nil {
		go rtpServer.gladiaClient.ReadText()
		go rtpServer.gladiaClient.AudioLoop()
	}

	for {
		buf := make([]byte, 2000)
		packetSize, _, err := rtpServer.socket.ReadFrom(buf)
		if err != nil {
			log.Println("Error reading from socket:", err)
			return err
		}

		rtpPacket := &rtp.Packet{}
		err = rtpPacket.Unmarshal(buf[:packetSize])
		if err != nil {
			log.Println("Error unmarshalling rtp packet:", err)
			continue
		}
		rtpServer.PayloadId = int(rtpPacket.PayloadType)
		_, err = rtpServer.file.Write(rtpPacket.Payload)
		if err != nil {
			log.Println("Error writing to file:", err)
		}
		rtpServer.gladiaClient.SendAudio(rtpPacket.Payload)
	}
}
