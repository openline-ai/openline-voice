package main

type CallMetadata struct {
	From      string
	To        string
	Uuid      string
	Direction CallDirection
}

type CallDirection string

const (
	IN  CallDirection = "in"
	OUT CallDirection = "out"
)
