package main

import "github.com/openline-ai/openline-oasis/packages/server/channels-api/model"

type CallMetadata struct {
	From       *model.VConParty
	To         *model.VConParty
	FromWebrtc bool
	Uuid       string
	Direction  CallDirection
}

type CallDirection string

const (
	IN  CallDirection = "in"
	OUT CallDirection = "out"
)
