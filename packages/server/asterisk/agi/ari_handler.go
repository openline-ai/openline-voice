package main

import (
	"github.com/CyCoreSystems/ari/v6"
	"github.com/CyCoreSystems/ari/v6/ext/bridgemon"
	"github.com/google/uuid"
	"log"
)

type ChannelVar struct {
	Uuid          string
	Dest          string
	KamailioIP    string
	EndpointName  string
	OriginCarrier *string
	DestCarrier   *string
}

func getChannelVars(h *ari.ChannelHandle) (*ChannelVar, error) {
	uuid, err := h.GetVariable("UUID")
	if err != nil {
		log.Printf("Missing channel var UUID: %v", err)
		return nil, err
	}
	dest, err := h.GetVariable("DEST")
	if err != nil {
		log.Printf("Missing channel var DEST: %v", err)
		return nil, err
	}
	kamailioIP, err := h.GetVariable("KAMAILIO_IP")
	if err != nil {
		log.Printf("Missing channel var KAMAILIO_IP: %v", err)
		return nil, err
	}
	endpointName, err := h.GetVariable("ENDPOINT_NAME")
	if err != nil {
		log.Printf("Missing channel var ENDPOINT_NAME: %v", err)
		return nil, err
	}
	originCarrier, err := h.GetVariable("ORIGIN_CARRIER")
	var originCarrierPtr *string = nil
	if err == nil {
		originCarrierPtr = &originCarrier
	}

	destCarrier, err := h.GetVariable("DEST_CARRIER")
	var destCarrierPtr *string = nil
	if err == nil {
		destCarrierPtr = &destCarrier
	}
	return &ChannelVar{Uuid: uuid, Dest: dest, KamailioIP: kamailioIP, EndpointName: endpointName, OriginCarrier: originCarrierPtr, DestCarrier: destCarrierPtr}, nil
}

func setDialVariables(h *ari.ChannelHandle, channelVars *ChannelVar) {
	h.SetVariable("PJSIP_HEADER(add,X-Openline-UUID)", channelVars.Uuid)
	h.SetVariable("UUID", channelVars.Uuid)
	h.SetVariable("PJSIP_HEADER(add,X-Openline-DEST)", channelVars.Dest)
	h.SetVariable("DEST", channelVars.Dest)
	h.SetVariable("KAMAILIO_IP", channelVars.KamailioIP)
	if channelVars.OriginCarrier != nil {
		h.SetVariable("PJSIP_HEADER(add,X-Openline-Origin-Carrier)", *channelVars.OriginCarrier)
		h.SetVariable("ORIGIN_CARRIER", *channelVars.OriginCarrier)
	}
	if channelVars.DestCarrier != nil {
		h.SetVariable("PJSIP_HEADER(add,X-Openline-Dest-Carrier)", *channelVars.DestCarrier)
		h.SetVariable("DEST_CARRIER", *channelVars.DestCarrier)
	}
	h.SetVariable("TRANSFER_CONTEXT", "transfer")

}
func app(cl ari.Client, h *ari.ChannelHandle) {
	log.Printf("running app channel: %s", h.Key().ID)
	channelVars, err := getChannelVars(h)
	if err != nil {
		log.Printf("Error getting channel vars: %v", err)
		err = cl.Channel().Hangup(h.Key(), "")
		h.Busy()
		return
	}

	dialString := "PJSIP/" + channelVars.EndpointName + "/sip:" + channelVars.KamailioIP
	dialedChannel, err := h.Create(ari.ChannelCreateRequest{
		Endpoint:  dialString,
		App:       cl.ApplicationName(),
		ChannelID: "managed-dialed-channel-" + h.ID(),
	})

	if err != nil {
		log.Printf("Error creating outbound channel: %v", err)
		h.Busy()
		return
	}
	setDialVariables(dialedChannel, channelVars)
	subAnswer := dialedChannel.Subscribe(ari.Events.ChannelStateChange)
	subHangup := dialedChannel.Subscribe(ari.Events.ChannelDestroyed)
	id, _ := h.GetVariable("CALLERID(num)")

	dialBridge, err := cl.Bridge().Create(ari.NewKey(ari.BridgeKey, uuid.New().String()), "mixing", "managed-dialBridge-"+h.ID())

	if err != nil {
		log.Printf("Error creating bridge: %v", err)
		h.Busy()
		return
	}
	err = dialBridge.AddChannel(h.ID())
	if err != nil {
		log.Printf("Error adding calling channel to bridge: %v", err)
		h.Busy()
		return
	}
	err = dialBridge.AddChannel(dialedChannel.ID())
	if err != nil {
		log.Printf("Error adding dialed channel to bridge: %v", err)
		h.Busy()
		return
	}

	err = cl.Channel().Dial(dialedChannel.Key(), id, 120)
	if err != nil {
		log.Printf("Error dialing: %v", err)
		h.Busy()
		return
	}
	for {
		select {
		case e := <-subAnswer.Events():
			v := e.(*ari.ChannelStateChange)
			log.Printf("Got Channel State Change for channel: %s new state: %s", v.Channel.ID, v.Channel.State)
			if v.Channel.State == "Up" {
				record(cl, h)
				return
			}
		case e := <-subHangup.Events():
			v := e.(*ari.ChannelDestroyed)
			log.Printf("Got Channel Destroyed for channel: %s", v.Channel.ID)
			return
		}
	}

}

func record(cl ari.Client, h *ari.ChannelHandle) {
	callUuid, _ := h.GetVariable("UUID")
	inData := CallMetadata{Uuid: callUuid, Direction: IN}

	inChannel, err := cl.Channel().Snoop(h.Key(), "", &ari.SnoopOptions{
		App: cl.ApplicationName(),
		Spy: ari.DirectionIn,
	})
	if err != nil {
		log.Printf("Error making Inbound Snoop: %v", err)
		return
	}

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
	err = inBridge.AddChannel(inChannel.ID())
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
