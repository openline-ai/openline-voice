package main

import (
	"github.com/CyCoreSystems/agi"
	"github.com/CyCoreSystems/ari/v6"
	"github.com/CyCoreSystems/ari/v6/client/native"
	"gopkg.in/ini.v1"
	"log"
)

func handler(a *agi.AGI, cl ari.Client) {
	channel := a.Variables["agi_channel"]
	uuid, err := a.Get("UUID")
	if err != nil {
		log.Printf("Mandatory channel var UUID is missing")
		return
	}
	inChannel, err := cl.Channel().Snoop(ari.NewKey(ari.ChannelKey, channel), "", &ari.SnoopOptions{
		App: cl.ApplicationName(),
		Spy: ari.DirectionIn,
	})
	if err != nil {
		log.Printf("Error making Inbound Snoop")
		return
	}
	outChannel, err := cl.Channel().Snoop(ari.NewKey(ari.ChannelKey, channel), "", &ari.SnoopOptions{
		App: cl.ApplicationName(),
		Spy: ari.DirectionOut,
	})
	if err != nil {
		log.Printf("Error making Outbound Snoop")
		err = cl.Channel().Hangup(inChannel.Key(), "")
		return
	}
	err = cl.Channel().Dial(inChannel.Key(), "AudioSocket://127.0.0.1:8090/"+uuid+"-in", 5)
	if err != nil {
		log.Printf("Error making Outbound Snoop")
		err = cl.Channel().Hangup(inChannel.Key(), "")
		return
	}
	err = cl.Channel().Dial(outChannel.Key(), "AudioSocket://127.0.0.1:8090/"+uuid+"-out", 5)
	if err != nil {
		log.Printf("Error making Outbound Snoop")
		err = cl.Channel().Hangup(inChannel.Key(), "")
		err = cl.Channel().Hangup(outChannel.Key(), "")

		return
	}

}
func main() {
	cfg, err := ini.Load("/etc/asterisk/ari.ini")
	if err != nil {
		log.Fatal("Unable to read config file")
	}

	cl, err := native.Connect(&native.Options{
		Application:  "recording",
		Username:     "asterisk",
		Password:     cfg.Section("asterisk").Key("password").String(),
		URL:          "http://localhost:8088/ari",
		WebsocketURL: "ws://localhost:8088/ari/events",
	})

	if err != nil {
		log.Fatal("Unable to create ari server")
	}

	err = agi.Listen(":8080", func(a *agi.AGI) { handler(a, cl) })
	if err != nil {
		log.Fatal("Unable to create agi server")
	}

}
