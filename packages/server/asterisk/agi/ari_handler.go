package main

import (
	"github.com/CyCoreSystems/ari/v6"
	"github.com/CyCoreSystems/ari/v6/ext/bridgemon"
	"github.com/google/uuid"
	"log"
	"os/exec"
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
	callUuid, err := h.GetVariable("UUID")
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
	return &ChannelVar{Uuid: callUuid, Dest: dest, KamailioIP: kamailioIP, EndpointName: endpointName, OriginCarrier: originCarrierPtr, DestCarrier: destCarrierPtr}, nil
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
	aHangup := h.Subscribe(ari.Events.ChannelDestroyed)
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
				var counter int = 0
				record(cl, h, IN, &counter)
				record(cl, h, OUT, &counter)
			}

		case e := <-subHangup.Events():
			v := e.(*ari.ChannelDestroyed)
			log.Printf("Got Channel Destroyed for channel: %s", v.Channel.ID)
			h.Hangup()
			return
		case e := <-aHangup.Events():
			v := e.(*ari.ChannelDestroyed)
			log.Printf("Got Channel Destroyed for channel: %s", v.Channel.ID)
			dialedChannel.Hangup()
			return
		}
	}

}

func record(cl ari.Client, h *ari.ChannelHandle, direction CallDirection, counter *int) {
	callUuid, _ := h.GetVariable("UUID")
	metadata := CallMetadata{Uuid: callUuid, Direction: direction}

	snoopOptions := &ari.SnoopOptions{
		App: cl.ApplicationName(),
	}
	if direction == IN {
		snoopOptions.Spy = ari.DirectionIn
	} else {
		snoopOptions.Spy = ari.DirectionOut
	}
	snoopChannel, err := cl.Channel().Snoop(h.Key(), "managed-"+string(direction)+"-snoop-"+h.ID(), snoopOptions)
	if err != nil {
		log.Printf("Error making %s Snoop: %v", direction, err)
		return
	}

	rtpServer := NewRtpServer(&metadata)
	log.Printf("%s RTP Server created: %s", direction, rtpServer.Address)
	go rtpServer.Listen()
	mediaChannel, err := cl.Channel().ExternalMedia(nil, ari.ExternalMediaOptions{
		App:          cl.ApplicationName(),
		ExternalHost: rtpServer.Address,
		Format:       "slin16",
		ChannelID:    "managed-" + string(direction) + "-" + h.ID(),
	})
	if err != nil {
		log.Printf("Error making %s AudioSocket: %v", direction, err)
		err = cl.Channel().Hangup(h.Key(), "")
		return
	}
	log.Printf("%s AudioSocket created: %v", direction, mediaChannel.Key())
	bridge, err := cl.Bridge().Create(ari.NewKey(ari.BridgeKey, uuid.New().String()), "mixing", "managed-"+string(direction)+"-"+h.ID())
	if err != nil {
		log.Printf("Error creating %s Bridge: %v", direction, err)
		err = cl.Channel().Hangup(h.Key(), "")
		err = cl.Channel().Hangup(mediaChannel.Key(), "")
		return
	}
	err = bridge.AddChannel(snoopChannel.ID())
	if err != nil {
		log.Printf("Error adding %s channel to bridge: %v", direction, err)
		err = cl.Channel().Hangup(h.Key(), "")
		err = cl.Channel().Hangup(mediaChannel.Key(), "")
		return
	}
	err = bridge.AddChannel(mediaChannel.ID())
	if err != nil {
		log.Printf("Error adding %s media channel to bridge: %v", direction, err)
		err = cl.Channel().Hangup(h.Key(), "")
		err = cl.Channel().Hangup(mediaChannel.Key(), "")
		return
	}
	inMonitor := bridgemon.New(bridge)
	inEvents := inMonitor.Watch()
	*counter++
	go func() {
		log.Printf("%s Bridge Monitor started", direction)
		for {
			m, ok := <-inEvents

			if !ok {
				log.Printf("%s Bridge Monitor closed", direction)
				return
			}
			log.Printf("%s Got event: %v", direction, m)

			if len(m.Channels()) <= 1 {
				err = cl.Channel().Hangup(mediaChannel.Key(), "")
				err = cl.Bridge().Delete(bridge.Key())
				rtpServer.Close()
				*counter--
				if *counter == 0 {
					processAudio(callUuid)
				}
			}
		}
	}()
}
func processAudio(callUuid string) error {
	cmd := exec.Command("sox", "-M", "-r", "16000", "-e", "signed-integer", "-c", "1", "-B", "-b", "16", "/tmp/"+callUuid+"-in.raw", "-r", "16000", "-e", "signed-integer", "-c", "1", "-B", "-b", "16", "/tmp/"+callUuid+"-out.raw", "/tmp/"+callUuid+".wav")
	err := cmd.Run()
	if err != nil {
		log.Printf("Error running sox: %v", err)
		return err
	} else {
		log.Printf("Wrote file: /tmp/%s.wav", callUuid)
		os.Remove("/tmp/" + callUuid + "-in.raw")
		os.Remove("/tmp/" + callUuid + "-out.raw")

	}
	return nil
}
