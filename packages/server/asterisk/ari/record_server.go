package main

import (
	"github.com/pion/rtp"
	"log"
	"net"
	"os"
)

const listenAddr = "127.0.0.1:0"

type RtpServer struct {
	Address   string
	Data      *CallMetadata
	PayloadId int
	socket    net.PacketConn
	file      *os.File
}

func NewRtpServer(cd *CallMetadata) *RtpServer {
	l, err := net.ListenPacket("udp", listenAddr)
	if err != nil {
		log.Fatal(err)
	}
	f, err := os.Create("/tmp/" + cd.Uuid + "-" + string(cd.Direction) + ".raw")

	return &RtpServer{
		Address: l.LocalAddr().String(),
		Data:    cd,
		socket:  l,
		file:    f,
	}
}

func (rtpServer RtpServer) Close() {
	rtpServer.socket.Close()
	rtpServer.file.Close()
}

func (rtpServer RtpServer) Listen() error {
	for {
		buf := make([]byte, 1024)
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
	}
}
