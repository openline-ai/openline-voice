package main

import (
	"github.com/pion/rtp"
	"log"
	"net"
	"os"
)

const listenAddr = "127.0.0.1:0"

type RtpServer struct {
	Address      string
	Data         *CallMetadata
	PayloadId    int
	socket       net.PacketConn
	file         *os.File
	gladiaClient *GladiaClient
}

func (rtpServer RtpServer) ListenForText() {
	for {
		select {
		case text := <-rtpServer.gladiaClient.channel:
			log.Println("************************Received text:", text)
		case <-rtpServer.gladiaClient.completed:
			return
		}
	}
}

func NewRtpServer(cd *CallMetadata) *RtpServer {
	l, err := net.ListenPacket("udp", listenAddr)
	if err != nil {
		log.Fatal(err)
	}
	f, err := os.Create("/tmp/" + cd.Uuid + "-" + string(cd.Direction) + ".raw")

	return &RtpServer{
		Address:      l.LocalAddr().String(),
		Data:         cd,
		socket:       l,
		file:         f,
		gladiaClient: NewGladiaClient(48000),
	}
}

func (rtpServer RtpServer) Close() {
	rtpServer.socket.Close()
	rtpServer.file.Close()
	rtpServer.gladiaClient.Close()
}

func (rtpServer RtpServer) Listen() error {
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
