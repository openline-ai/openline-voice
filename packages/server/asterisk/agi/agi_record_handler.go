package main

import (
	"fmt"
	"github.com/CyCoreSystems/agi"
	"github.com/CyCoreSystems/ari/v6"
	"github.com/CyCoreSystems/ari/v6/ext/bridgemon"
	"github.com/google/uuid"
	"log"
	"os"
	"os/exec"
)

func handler(a *agi.AGI, cl ari.Client, streamMap *CallData) {
	channel := a.Variables["agi_uniqueid"]
	callUuid, err := a.Get("UUID")
	defer a.Close()

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
	inData := CallMetadata{Uuid: callUuid, Direction: IN}
	streamMap.AddStream(inData)
	inRtpServer := NewRtpServer(&inData)
	go inRtpServer.Listen()
	mediaInChannel, err := cl.Channel().ExternalMedia(inChannel.Key(), ari.ExternalMediaOptions{
		App:           cl.ApplicationName(),
		ExternalHost:  "127.0.0.1:8090",
		Encapsulation: "rtp",
		Transport:     "tcp",
		Format:        "slin16",
	})
	if err != nil {
		a.Verbose(fmt.Sprintf("Error making Inbound AudioSocket: %v", err), 1)
		err = cl.Channel().Hangup(inChannel.Key(), "")
		streamMap.RemoveStream(inData)
		return
	}
	a.Verbose(fmt.Sprintf("Inbound AudioSocket created: %v", mediaInChannel.Key()), 1)

	outData := CallMetadata{Uuid: callUuid, Direction: OUT}
	streamMap.AddStream(outData)
	outRtpServer := NewRtpServer(&inData)
	go outRtpServer.Listen()
	mediaOutChannel, err := cl.Channel().ExternalMedia(outChannel.Key(), ari.ExternalMediaOptions{
		App:           cl.ApplicationName(),
		ExternalHost:  "127.0.0.1:8090",
		Encapsulation: "rtp",
		Transport:     "tcp",
		Format:        "slin16",
	})
	if err != nil {
		a.Verbose(fmt.Sprintf("Error making Outbound AudioSocket: %v", err), 1)
		err = cl.Channel().Hangup(inChannel.Key(), "")
		err = cl.Channel().Hangup(outChannel.Key(), "")
		streamMap.RemoveStream(inData)
		streamMap.RemoveStream(outData)
		return
	}

	a.Verbose(fmt.Sprintf("Outbound AudioSocket created: %v", mediaOutChannel.Key()), 1)
	inBridge, err := cl.Bridge().Create(ari.NewKey(ari.BridgeKey, uuid.New().String()), "holding", "inboundBridge")
	if err != nil {
		a.Verbose(fmt.Sprintf("Error creating Inbound Bridge: %v", err), 1)
		err = cl.Channel().Hangup(inChannel.Key(), "")
		err = cl.Channel().Hangup(outChannel.Key(), "")
		err = cl.Channel().Hangup(mediaInChannel.Key(), "")
		err = cl.Channel().Hangup(mediaOutChannel.Key(), "")
		streamMap.RemoveStream(inData)
		streamMap.RemoveStream(outData)
		return
	}
	outBridge, err := cl.Bridge().Create(ari.NewKey(ari.BridgeKey, uuid.New().String()), "holding", "outboundBridge")
	if err != nil {
		a.Verbose(fmt.Sprintf("Error creating Outbound Bridge: %v", err), 1)
		err = cl.Channel().Hangup(inChannel.Key(), "")
		err = cl.Channel().Hangup(outChannel.Key(), "")
		err = cl.Channel().Hangup(mediaInChannel.Key(), "")
		err = cl.Channel().Hangup(mediaOutChannel.Key(), "")
		err = cl.Bridge().Delete(inBridge.Key())
		streamMap.RemoveStream(inData)
		streamMap.RemoveStream(outData)
		return
	}
	err = inBridge.AddChannelWithOptions(inChannel.ID(), &ari.BridgeAddChannelOptions{Role: "announcer"})
	if err != nil {
		a.Verbose(fmt.Sprintf("Error adding Inbound Channel to Inbound Bridge: %v", err), 1)
		err = cl.Channel().Hangup(inChannel.Key(), "")
		err = cl.Channel().Hangup(outChannel.Key(), "")
		err = cl.Channel().Hangup(mediaInChannel.Key(), "")
		err = cl.Channel().Hangup(mediaOutChannel.Key(), "")
		err = cl.Bridge().Delete(inBridge.Key())
		err = cl.Bridge().Delete(outBridge.Key())
		streamMap.RemoveStream(inData)
		streamMap.RemoveStream(outData)
		return
	}
	err = inBridge.AddChannelWithOptions(mediaInChannel.ID(), &ari.BridgeAddChannelOptions{Role: "participant"})
	if err != nil {
		a.Verbose(fmt.Sprintf("Error adding Inbound Media Channel to Inbound Bridge: %v", err), 1)
		err = cl.Channel().Hangup(inChannel.Key(), "")
		err = cl.Channel().Hangup(outChannel.Key(), "")
		err = cl.Channel().Hangup(mediaInChannel.Key(), "")
		err = cl.Channel().Hangup(mediaOutChannel.Key(), "")
		err = cl.Bridge().Delete(inBridge.Key())
		err = cl.Bridge().Delete(outBridge.Key())
		streamMap.RemoveStream(inData)
		streamMap.RemoveStream(outData)
		return
	}
	err = outBridge.AddChannelWithOptions(outChannel.ID(), &ari.BridgeAddChannelOptions{Role: "announcer"})
	if err != nil {
		a.Verbose(fmt.Sprintf("Error adding Outbound Channel to Outbound Bridge: %v", err), 1)
		err = cl.Channel().Hangup(inChannel.Key(), "")
		err = cl.Channel().Hangup(outChannel.Key(), "")
		err = cl.Channel().Hangup(mediaInChannel.Key(), "")
		err = cl.Channel().Hangup(mediaOutChannel.Key(), "")
		err = cl.Bridge().Delete(inBridge.Key())
		err = cl.Bridge().Delete(outBridge.Key())
		streamMap.RemoveStream(inData)
		streamMap.RemoveStream(outData)
		return
	}
	err = outBridge.AddChannelWithOptions(mediaOutChannel.ID(), &ari.BridgeAddChannelOptions{Role: "participant"})
	if err != nil {
		a.Verbose(fmt.Sprintf("Error adding Outbound Media Channel to Outbound Bridge: %v", err), 1)
		err = cl.Channel().Hangup(inChannel.Key(), "")
		err = cl.Channel().Hangup(outChannel.Key(), "")
		err = cl.Channel().Hangup(mediaInChannel.Key(), "")
		err = cl.Channel().Hangup(mediaOutChannel.Key(), "")
		err = cl.Bridge().Delete(inBridge.Key())
		err = cl.Bridge().Delete(outBridge.Key())
		streamMap.RemoveStream(inData)
		streamMap.RemoveStream(outData)
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

				if streamMap.RemoveStream(inData) {
					cmd := exec.Command("sox", "-M", "-r", "8000", "-e", "a-law", "-c", "1", "/tmp/"+callUuid+"-in.raw", "-r", "8000", "-e", "a-law", "-c", "1", "/tmp/"+callUuid+"-out.raw", "/tmp/"+callUuid+".wav")
					err = cmd.Run()
					if err != nil {
						log.Printf("Error running sox: %v", err)
					} else {
						log.Printf("Wrote file: /tmp/%s.wav", callUuid)
						os.Remove("/tmp/" + callUuid + "-in.raw")
						os.Remove("/tmp/" + callUuid + "-out.raw")
					}

				}
			}
		}
	}()

	outMonitor := bridgemon.New(outBridge)
	outEvents := outMonitor.Watch()

	go func() {
		log.Printf("Outbound Bridge Monitor started")
		for {
			m, ok := <-outEvents

			if !ok {
				log.Printf("Outbound Bridge Monitor closed")
				return
			}
			log.Printf("Got event: %v", m)
			if len(m.Channels()) <= 1 {
				err = cl.Channel().Hangup(mediaOutChannel.Key(), "")
				err = cl.Bridge().Delete(outBridge.Key())
				outRtpServer.Close()
			}
			if streamMap.RemoveStream(outData) {
				cmd := exec.Command("sox", "-M", "-r", "1600", "-e", "signed-integer", "-c", "1", "/tmp/"+callUuid+"-in.raw", "-r", "16000", "-e", "signed-integer", "-c", "1", "/tmp/"+callUuid+"-out.raw", "-M", "-c", "2", "/tmp/"+callUuid+".wav")
				err = cmd.Run()
				if err != nil {
					log.Printf("Error running sox: %v", err)
				} else {
					log.Printf("Wrote file: /tmp/%s.wav", callUuid)
					//os.Remove("/tmp/" + callUuid + "-in.raw")
					//os.Remove("/tmp/" + callUuid + "-out.raw")
				}

			}
		}
	}()
}
