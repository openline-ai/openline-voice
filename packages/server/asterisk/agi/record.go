package main

import (
	"context"
	"github.com/CyCoreSystems/agi"
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

	cd := NewCallData()
	if err != nil {
		log.Fatalf("Unable to create ari server %v", err)
	}

	go agi.Listen(":8080", func(a *agi.AGI) { handler(a, cl, cd) })

	Listen(context.Background(), cd)

}
