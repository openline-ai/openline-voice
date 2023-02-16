package main

import (
	"github.com/CyCoreSystems/ari/v6"
	"github.com/CyCoreSystems/ari/v6/client/native"
	"gopkg.in/ini.v1"
	"log"
)

func main() {
	cfg, err := ini.Load("/etc/asterisk/ari.conf")
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

	//cd := NewCallData()
	if err != nil {
		log.Fatalf("Unable to create ari server %v", err)
	}
	log.Printf("Asterisk ARI client created")
	log.Printf("Listening for new calls")
	sub := cl.Bus().Subscribe(nil, "StasisStart")

	for {
		select {
		case e := <-sub.Events():
			v := e.(*ari.StasisStart)
			log.Printf("Got stasis start channel: %s", v.Channel.ID)
			go app(cl, cl.Channel().Get(v.Key(ari.ChannelKey, v.Channel.ID)))
		}
	}

}
