package main

import (
	"encoding/json"
	"github.com/openline-ai/openline-oasis/packages/server/channels-api/model"
	"github.com/pion/rtp"
	"log"
	"net"
	"os"
	"time"
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
}

func (rtpServer RtpServer) ListenForText() {
	log.Printf("Started listening for text")
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

func NewRtpServer(cd *CallMetadata, publisher *VConEventPublisher) *RtpServer {
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
		gladiaClient:  NewGladiaClient(48000),
		vconPublisher: publisher,
	}
}

func (rtpServer RtpServer) Close() {
	rtpServer.socket.Close()
	rtpServer.file.Close()
	buf := make([]byte, 1920)
	log.Printf("Flushing remaining audio")
	rtpServer.gladiaClient.SendAudio(buf)
	rtpServer.gladiaClient.SendAudio(buf)
	rtpServer.gladiaClient.SendAudio(buf)
	rtpServer.gladiaClient.SendAudio(buf)
	rtpServer.gladiaClient.SendAudio(buf)
	rtpServer.gladiaClient.SendAudio(buf)
	rtpServer.gladiaClient.SendAudio(buf)
	rtpServer.gladiaClient.SendAudio(buf)
	rtpServer.gladiaClient.SendAudio(buf)
	rtpServer.gladiaClient.SendAudio(buf)
	time.Sleep(5 * time.Second)
	rtpServer.gladiaClient.Close()
}

func (rtpServer RtpServer) Listen() error {
	go rtpServer.gladiaClient.ReadText()
	go rtpServer.gladiaClient.AudioLoop()
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
