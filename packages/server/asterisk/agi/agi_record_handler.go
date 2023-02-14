package main

import (
	"fmt"
	"github.com/CyCoreSystems/agi"
	"github.com/CyCoreSystems/ari/v5"
	"github.com/google/uuid"
)

func handler(a *agi.AGI, cl ari.Client, streamMap *CallData) {
	channel := a.Variables["agi_uniqueid"]
	callUuid, err := a.Get("UUID")
	if err != nil {
		a.Verbose(fmt.Sprintf("Mandatory channel var UUID is missing: %v", err), 1)
		return
	}
	inChannel, err := cl.Channel().Snoop(ari.NewKey(ari.ChannelKey, channel), "", &ari.SnoopOptions{
		App: cl.ApplicationName(),
		Spy: ari.DirectionIn,
	})
	if err != nil {
		a.Verbose(fmt.Sprintf("Error making Inbound Snoop: %v", err), 1)
		return
	}
	outChannel, err := cl.Channel().Snoop(ari.NewKey(ari.ChannelKey, channel), "", &ari.SnoopOptions{
		App: cl.ApplicationName(),
		Spy: ari.DirectionOut,
	})
	if err != nil {
		a.Verbose(fmt.Sprintf("Error making Outbound Snoop: %v", err), 1)
		err = cl.Channel().Hangup(inChannel.Key(), "")
		return
	}
	inUuid := uuid.New().String()
	streamMap.AddStream(inUuid, CallMetadata{Uuid: callUuid, Direction: IN})
	mediaInChannel, err := cl.Channel().ExternalMedia(inChannel.Key(), ari.ExternalMediaOptions{
		App:           cl.ApplicationName(),
		Data:          inUuid,
		ExternalHost:  "127.0.0.1:8090",
		Encapsulation: "audiosocket",
		Transport:     "tcp",
		Format:        "slin16",
	}, map[string]string{})
	if err != nil {
		a.Verbose(fmt.Sprintf("Error making Inbound AudioSocket: %v", err), 1)
		err = cl.Channel().Hangup(inChannel.Key(), "")
		streamMap.RemoveStream(inUuid)
		return
	}
	a.Verbose(fmt.Sprintf("Inbound AudioSocket created: %v", mediaInChannel.Key()), 1)

	outUuid := uuid.New().String()
	streamMap.AddStream(outUuid, CallMetadata{Uuid: callUuid, Direction: OUT})
	mediaOutChannel, err := cl.Channel().ExternalMedia(outChannel.Key(), ari.ExternalMediaOptions{
		App:           cl.ApplicationName(),
		Data:          outUuid,
		ExternalHost:  "127.0.0.1:8090",
		Encapsulation: "audiosocket",
		Transport:     "tcp",
		Format:        "slin16",
	}, map[string]string{})
	if err != nil {
		a.Verbose(fmt.Sprintf("Error making Outbound AudioSocket: %v", err), 1)
		err = cl.Channel().Hangup(inChannel.Key(), "")
		err = cl.Channel().Hangup(outChannel.Key(), "")
		streamMap.RemoveStream(inUuid)
		streamMap.RemoveStream(outUuid)
		return
	}

	a.Verbose(fmt.Sprintf("Outbound AudioSocket created: %v", mediaOutChannel.Key()), 1)
	inBridge, err := cl.Bridge().Create(ari.NewKey(ari.BridgeKey, uuid.New().String()), "mixing", "inboundBridge")
	if err != nil {
		a.Verbose(fmt.Sprintf("Error creating Inbound Bridge: %v", err), 1)
		err = cl.Channel().Hangup(inChannel.Key(), "")
		err = cl.Channel().Hangup(outChannel.Key(), "")
		err = cl.Channel().Hangup(mediaInChannel.Key(), "")
		err = cl.Channel().Hangup(mediaOutChannel.Key(), "")
		streamMap.RemoveStream(inUuid)
		streamMap.RemoveStream(outUuid)
		return
	}
	outBridge, err := cl.Bridge().Create(ari.NewKey(ari.BridgeKey, uuid.New().String()), "mixing", "outboundBridge")
	if err != nil {
		a.Verbose(fmt.Sprintf("Error creating Outbound Bridge: %v", err), 1)
		err = cl.Channel().Hangup(inChannel.Key(), "")
		err = cl.Channel().Hangup(outChannel.Key(), "")
		err = cl.Channel().Hangup(mediaInChannel.Key(), "")
		err = cl.Channel().Hangup(mediaOutChannel.Key(), "")
		err = cl.Bridge().Delete(inBridge.Key())
		streamMap.RemoveStream(inUuid)
		streamMap.RemoveStream(outUuid)
		return
	}
	err = inBridge.AddChannel(inChannel.ID())
	if err != nil {
		a.Verbose(fmt.Sprintf("Error adding Inbound Channel to Inbound Bridge: %v", err), 1)
		err = cl.Channel().Hangup(inChannel.Key(), "")
		err = cl.Channel().Hangup(outChannel.Key(), "")
		err = cl.Channel().Hangup(mediaInChannel.Key(), "")
		err = cl.Channel().Hangup(mediaOutChannel.Key(), "")
		err = cl.Bridge().Delete(inBridge.Key())
		err = cl.Bridge().Delete(outBridge.Key())
		streamMap.RemoveStream(inUuid)
		streamMap.RemoveStream(outUuid)
		return
	}
	err = inBridge.AddChannel(mediaInChannel.ID())
	if err != nil {
		a.Verbose(fmt.Sprintf("Error adding Inbound Media Channel to Inbound Bridge: %v", err), 1)
		err = cl.Channel().Hangup(inChannel.Key(), "")
		err = cl.Channel().Hangup(outChannel.Key(), "")
		err = cl.Channel().Hangup(mediaInChannel.Key(), "")
		err = cl.Channel().Hangup(mediaOutChannel.Key(), "")
		err = cl.Bridge().Delete(inBridge.Key())
		err = cl.Bridge().Delete(outBridge.Key())
		streamMap.RemoveStream(inUuid)
		streamMap.RemoveStream(outUuid)
		return
	}
	err = outBridge.AddChannel(outChannel.ID())
	if err != nil {
		a.Verbose(fmt.Sprintf("Error adding Outbound Channel to Outbound Bridge: %v", err), 1)
		err = cl.Channel().Hangup(inChannel.Key(), "")
		err = cl.Channel().Hangup(outChannel.Key(), "")
		err = cl.Channel().Hangup(mediaInChannel.Key(), "")
		err = cl.Channel().Hangup(mediaOutChannel.Key(), "")
		err = cl.Bridge().Delete(inBridge.Key())
		err = cl.Bridge().Delete(outBridge.Key())
		streamMap.RemoveStream(inUuid)
		streamMap.RemoveStream(outUuid)
		return
	}
	err = outBridge.AddChannel(mediaOutChannel.ID())
	if err != nil {
		a.Verbose(fmt.Sprintf("Error adding Outbound Media Channel to Outbound Bridge: %v", err), 1)
		err = cl.Channel().Hangup(inChannel.Key(), "")
		err = cl.Channel().Hangup(outChannel.Key(), "")
		err = cl.Channel().Hangup(mediaInChannel.Key(), "")
		err = cl.Channel().Hangup(mediaOutChannel.Key(), "")
		err = cl.Bridge().Delete(inBridge.Key())
		err = cl.Bridge().Delete(outBridge.Key())
		streamMap.RemoveStream(inUuid)
		streamMap.RemoveStream(outUuid)
		return
	}

}
