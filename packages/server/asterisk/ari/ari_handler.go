package main

import (
	"encoding/json"
	"github.com/CyCoreSystems/ari/v6"
	"github.com/CyCoreSystems/ari/v6/ext/bridgemon"
	"github.com/ghettovoice/gosip/sip/parser"
	"github.com/google/uuid"
	"github.com/openline-ai/openline-oasis/packages/server/channels-api/model"
	"log"
	"os"
	"os/exec"
)

type ChannelVar struct {
	Uuid             string
	Dest             string
	KamailioIP       string
	EndpointName     string
	OrigEndpointName string
	OriginCarrier    *string
	DestCarrier      *string
	From             *model.VConParty
	To               *model.VConParty
}

func makeMetaData(direction CallDirection, vars *ChannelVar) *CallMetadata {
	return &CallMetadata{
		Direction: direction,
		Uuid:      vars.Uuid,
		From:      vars.From,
		To:        vars.To,
	}
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
	origEndpointName, err := h.GetVariable("PJSIP_HEADER(read,X-Openline-Endpoint-Type)")
	if err != nil {
		log.Printf("Missing channel var PJSIP_HEADER(read,X-Openline-Endpoint-Type): %v", err)
		return nil, err
	}
	fromId := &model.VConParty{}
	from, err := h.GetVariable("PJSIP_HEADER(read,From)")
	if err != nil {
		log.Printf("Missing channel var PJSIP_HEADER(read,From): %v", err)
		return nil, err
	}
	_, uri, _, err := parser.ParseAddressValue(from)
	if err != nil {
		log.Printf("Error parsing From header: %v", err)
		return nil, err
	}

	if origEndpointName == "webrtc" {
		fromIdStr := uri.User().String() + "@" + uri.Host()
		fromId.Mailto = &fromIdStr

	} else {
		fromIdStr := uri.User().String()
		fromId.Tel = &fromIdStr
	}

	toId := &model.VConParty{}
	toUser, err := h.GetVariable("PJSIP_HEADER(read,X-Openline-Dest-User)")
	if err != nil {
		toUser = ""
	}

	to, err := h.GetVariable("PJSIP_HEADER(read,X-Openline-Dest)")
	if err != nil {
		log.Printf("Missing channel var PJSIP_HEADER(read,X-Openline-Dest): %v", err)
		return nil, err
	}

	if endpointName == "webrtc" {
		uri, err = parser.ParseUri(to)
		if err != nil {
			log.Printf("Error parsing To header: %v", err)
			return nil, err
		}

		toStr := uri.User().String() + "@" + uri.Host()
		toId.Mailto = &toStr
	} else if toUser != "" {
		toId.Mailto = &toUser
	} else {
		toStr := uri.User().String()
		toId.Tel = &toStr
	}

	return &ChannelVar{Uuid: callUuid,
		Dest:             dest,
		KamailioIP:       kamailioIP,
		EndpointName:     endpointName,
		OrigEndpointName: origEndpointName,
		OriginCarrier:    originCarrierPtr,
		DestCarrier:      destCarrierPtr,
		From:             fromId,
		To:               toId}, nil
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
func app(cl ari.Client, h *ari.ChannelHandle, conf *RecordServiceConfig) {
	log.Printf("Running app channel: %s", h.Key().ID)
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
	subHangup := dialedChannel.Subscribe(ari.Events.ChannelHangupRequest)
	aHangup := h.Subscribe(ari.Events.ChannelHangupRequest)
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
				//func NewVConEventPublisher(conf *RecordServiceConfig, uuid string, from *model.VConParty, to *model.VConParty) *VConEventPublisher {

				vconPublisher := NewVConEventPublisher(conf, channelVars.Uuid, channelVars.From, channelVars.To)
				go vconPublisher.Run()
				record(cl, h, makeMetaData(IN, channelVars), &counter, vconPublisher, conf)
				record(cl, h, makeMetaData(OUT, channelVars), &counter, vconPublisher, conf)
			}

		case e := <-subHangup.Events():
			v := e.(*ari.ChannelHangupRequest)
			log.Printf("Got Channel Hangup for channel: %s", v.Channel.ID)
			h.Hangup()
			dialBridge.Delete()
			return
		case e := <-aHangup.Events():
			v := e.(*ari.ChannelHangupRequest)
			log.Printf("Got Channel Hangup for channel: %s", v.Channel.ID)
			dialedChannel.Hangup()
			dialBridge.Delete()
			return
		}
	}

}

func record(cl ari.Client, h *ari.ChannelHandle, metadata *CallMetadata, counter *int, publisher *VConEventPublisher, conf *RecordServiceConfig) {

	snoopOptions := &ari.SnoopOptions{
		App: cl.ApplicationName(),
	}
	if metadata.Direction == IN {
		snoopOptions.Spy = ari.DirectionIn
	} else {
		snoopOptions.Spy = ari.DirectionOut
	}
	snoopChannel, err := cl.Channel().Snoop(h.Key(), "managed-"+string(metadata.Direction)+"-snoop-"+h.ID(), snoopOptions)
	if err != nil {
		log.Printf("Error making %s Snoop: %v", metadata.Direction, err)
		return
	}

	rtpServer := NewRtpServer(metadata, publisher, conf)

	log.Printf("%s RTP Server created: %s", metadata.Direction, rtpServer.Address)
	go rtpServer.Listen()
	go rtpServer.ListenForText()
	mediaChannel, err := cl.Channel().ExternalMedia(nil, ari.ExternalMediaOptions{
		App:          cl.ApplicationName(),
		ExternalHost: rtpServer.Address,
		Format:       "slin48",
		ChannelID:    "managed-" + string(metadata.Direction) + "-" + h.ID(),
	})
	if err != nil {
		log.Printf("Error making %s AudioSocket: %v", metadata.Direction, err)
		err = cl.Channel().Hangup(h.Key(), "")
		return
	}
	log.Printf("%s AudioSocket created: %v", metadata.Direction, mediaChannel.Key())
	bridge, err := cl.Bridge().Create(ari.NewKey(ari.BridgeKey, uuid.New().String()), "mixing", "managed-"+string(metadata.Direction)+"-"+h.ID())
	if err != nil {
		log.Printf("Error creating %s Bridge: %v", metadata.Direction, err)
		err = cl.Channel().Hangup(h.Key(), "")
		err = cl.Channel().Hangup(mediaChannel.Key(), "")
		return
	}
	err = bridge.AddChannel(snoopChannel.ID())
	if err != nil {
		log.Printf("Error adding %s channel to bridge: %v", metadata.Direction, err)
		err = cl.Channel().Hangup(h.Key(), "")
		err = cl.Channel().Hangup(mediaChannel.Key(), "")
		return
	}
	err = bridge.AddChannel(mediaChannel.ID())
	if err != nil {
		log.Printf("Error adding %s media channel to bridge: %v", metadata.Direction, err)
		err = cl.Channel().Hangup(h.Key(), "")
		err = cl.Channel().Hangup(mediaChannel.Key(), "")
		return
	}
	inMonitor := bridgemon.New(bridge)
	inEvents := inMonitor.Watch()
	*counter++
	go func() {
		log.Printf("%s Bridge Monitor started", metadata.Direction)
		for {
			m, ok := <-inEvents

			if !ok {
				log.Printf("%s Bridge Monitor closed", metadata.Direction)
				return
			}
			log.Printf("%s Got event: %v", metadata.Direction, m)

			if len(m.Channels()) <= 1 {
				err = cl.Channel().Hangup(mediaChannel.Key(), "")
				err = cl.Bridge().Delete(bridge.Key())
				rtpServer.Close()
				*counter--
				if *counter == 0 {
					audioFile, err := processAudio(metadata.Uuid)
					if err == nil {
						script, err := TranscribeAudio(conf, audioFile, metadata.From, metadata.To)
						if err != nil {
							log.Printf("Error transcribing audio: %v", err)
						} else {
							log.Printf("Transcript: %s", script)
							transcriptBytes, err := json.Marshal(script)
							if err == nil {
								publisher.SendAnalysis(model.TRANSCRIPT, "application/x-openline-transcript", string(transcriptBytes))
							}
							summary, err := ConversationSummary(conf, script)
							if err != nil {
								log.Printf("Error summarizing conversation: %v", err)
							} else {
								log.Printf("Summary: %s", summary)
								publisher.SendAnalysis(model.SUMMARY, "text/plain", summary)
							}
						}

					} else {
						log.Printf("Error processing audio: %v", err)
					}
					publisher.Kill()
				}
			}
		}
	}()
}
func processAudio(callUuid string) (string, error) {
	outputFile := "/tmp/" + callUuid + ".ogg"
	cmd := exec.Command("sox", "-M", "-r", "48000", "-e", "signed-integer", "-c", "1", "-B", "-b", "16", "/tmp/"+callUuid+"-in.raw", "-r", "48000", "-e", "signed-integer", "-c", "1", "-B", "-b", "16", "/tmp/"+callUuid+"-out.raw", outputFile)
	err := cmd.Run()
	if err != nil {
		log.Printf("Error Running sox: %v", err)
		return "", err
	} else {
		log.Printf("Wrote file: %s", callUuid)
		os.Remove("/tmp/" + callUuid + "-in.raw")
		os.Remove("/tmp/" + callUuid + "-out.raw")

	}
	return outputFile, nil
}
