package main

import (
	"fmt"
	"github.com/caarlos0/env/v6"
	_ "github.com/lib/pq"
	c "github.com/openline-ai/openline-voice/packages/apps/voice-plugin/config"
	"github.com/openline-ai/openline-voice/packages/apps/voice-plugin/gen"
	"github.com/openline-ai/openline-voice/packages/apps/voice-plugin/routes"
	"log"
)

//go:generate swag init -g routes/main.go
//go:generate go run -mod=mod entgo.io/ent/cmd/ent generate --target ./gen ./schema

func main() {
	conf := c.Config{}
	if err := env.Parse(&conf); err != nil {
		fmt.Printf("missing required config")
		return
	}

	var connUrl = fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=disable", conf.DB.Host, conf.DB.Port, conf.DB.User, conf.DB.Name, conf.DB.Password)
	log.Printf("Connecting to database %s", connUrl)
	client, err := gen.Open("postgres", connUrl)

	if err != nil {
		log.Fatalf("failed opening connection to postgres: %v", err)
	}
	routes.Run(&conf, client)
}
