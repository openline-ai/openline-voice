package main

import (
	"github.com/CyCoreSystems/ari/v6"
	"log"
)

func app(cl ari.Client, h *ari.ChannelHandle) {
	log.Printf("running app channel: %s", h.Key().ID)

	if err := h.Answer(); err != nil {
		log.Printf("failed to answer call: %v", err)
		// return
	}
	callUuid, _ := h.GetVariable("UUID")
	inData := CallMetadata{Uuid: callUuid, Direction: IN}

	inRtpServer := NewRtpServer(&inData)
	log.Printf("Inbound RTP Server created: %s", inRtpServer.Address)
	go inRtpServer.Listen()
	mediaInChannel, err := cl.Channel().ExternalMedia(h.Key(), ari.ExternalMediaOptions{
		App:          cl.ApplicationName(),
		ExternalHost: inRtpServer.Address,
		Format:       "slin16",
		ChannelID:    "managed-inbound-" + h.ID(),
	})
	if err != nil {
		log.Printf("Error making Inbound AudioSocket: %v", err)
		err = cl.Channel().Hangup(h.Key(), "")
		return
	}
	log.Printf("Inbound AudioSocket created: %v", mediaInChannel.Key())

}
