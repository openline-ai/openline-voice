package main

import (
	"github.com/CyCoreSystems/ari/v6"
	"github.com/CyCoreSystems/ari/v6/ext/bridgemon"
	"github.com/google/uuid"
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
	mediaInChannel, err := cl.Channel().ExternalMedia(nil, ari.ExternalMediaOptions{
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
	inBridge, err := cl.Bridge().Create(ari.NewKey(ari.BridgeKey, uuid.New().String()), "mixing", "managed-inboundBridge-"+h.ID())
	if err != nil {
		log.Printf("Error creating Inbound Bridge: %v", err)
		err = cl.Channel().Hangup(h.Key(), "")
		//err = cl.Channel().Hangup(outChannel.Key(), "")
		err = cl.Channel().Hangup(mediaInChannel.Key(), "")
		//err = cl.Channel().Hangup(mediaOutChannel.Key(), "")
		//streamMap.RemoveStream(outData)
		return
	}
	err = inBridge.AddChannel(h.ID())
	if err != nil {
		log.Printf("Error adding inbound channel to bridge: %v", err)
		err = cl.Channel().Hangup(h.Key(), "")
		//err = cl.Channel().Hangup(outChannel.Key(), "")
		err = cl.Channel().Hangup(mediaInChannel.Key(), "")
		//err = cl.Channel().Hangup(mediaOutChannel.Key(), "")
		//streamMap.RemoveStream(outData)
		return
	}
	err = inBridge.AddChannel(mediaInChannel.ID())
	if err != nil {
		log.Printf("Error adding inbound media channel to bridge: %v", err)
		err = cl.Channel().Hangup(h.Key(), "")
		//err = cl.Channel().Hangup(outChannel.Key(), "")
		err = cl.Channel().Hangup(mediaInChannel.Key(), "")
		//err = cl.Channel().Hangup(mediaOutChannel.Key(), "")
		//streamMap.RemoveStream(outData)
		return
	}
	inMonitor := bridgemon.New(inBridge)
	inEvents := inMonitor.Watch()

	go func() {
		log.Printf("Inbound Bridge Monitor started")
		for {
			m, ok := <-inEvents

			if !ok {
				log.Printf("Inbound Bridge Monitor closed")
				return
			}
			log.Printf("Got event: %v", m)

			if len(m.Channels()) <= 1 {
				err = cl.Channel().Hangup(mediaInChannel.Key(), "")
				err = cl.Bridge().Delete(inBridge.Key())
				inRtpServer.Close()

			}
		}
	}()

}
