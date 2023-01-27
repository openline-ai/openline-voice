package main

import (
	"github.com/CyCoreSystems/agi"
	"github.com/CyCoreSystems/ari/v6"
	"github.com/google/uuid"
)

func handler(a *agi.AGI, cl ari.Client, streamMap *CallData) {
	channel := a.Variables["agi_uniqueid"]
	callUuid, err := a.Get("UUID")
	if err != nil {
		a.Verbose("Mandatory channel var UUID is missing", 1)
		return
	}
	inChannel, err := cl.Channel().Snoop(ari.NewKey(ari.ChannelKey, channel), "", &ari.SnoopOptions{
		App: cl.ApplicationName(),
		Spy: ari.DirectionIn,
	})
	if err != nil {
		a.Verbose("Error making Inbound Snoop", 1)
		return
	}
	outChannel, err := cl.Channel().Snoop(ari.NewKey(ari.ChannelKey, channel), "", &ari.SnoopOptions{
		App: cl.ApplicationName(),
		Spy: ari.DirectionOut,
	})
	if err != nil {
		a.Verbose("Error making Outbound Snoop", 1)
		err = cl.Channel().Hangup(inChannel.Key(), "")
		return
	}
	inUuid := uuid.New().String()
	streamMap.AddStream(inUuid, CallMetadata{Uuid: callUuid, Direction: IN})
	_, err = cl.Channel().ExternalMedia(inChannel.Key(), ari.ExternalMediaOptions{
		App:           cl.ApplicationName(),
		ChannelID:     inUuid,
		ExternalHost:  "127.0.0.1:8090",
		Encapsulation: "audiosocket",
		Transport:     "tcp",
		Format:        "slin",
	})
	if err != nil {
		a.Verbose("Error making Inbound AudioSocket", 1)
		err = cl.Channel().Hangup(inChannel.Key(), "")
		streamMap.RemoveStream(inUuid)
		return
	}
	outUuid := uuid.New().String()
	streamMap.AddStream(outUuid, CallMetadata{Uuid: callUuid, Direction: OUT})
	_, err = cl.Channel().ExternalMedia(outChannel.Key(), ari.ExternalMediaOptions{
		App:           cl.ApplicationName(),
		ChannelID:     outUuid,
		ExternalHost:  "127.0.0.1:8090",
		Encapsulation: "audiosocket",
		Transport:     "tcp",
		Format:        "slin",
	})
	if err != nil {
		a.Verbose("Error making Outbound AudioSocket", 1)
		err = cl.Channel().Hangup(inChannel.Key(), "")
		err = cl.Channel().Hangup(outChannel.Key(), "")
		streamMap.RemoveStream(inUuid)
		streamMap.RemoveStream(outUuid)
		return
	}

}
