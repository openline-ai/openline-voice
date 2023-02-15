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

type CallData struct {
	channelCounting map[string]map[CallDirection]int
}

func (cd *CallData) AddStream(data CallMetadata) {
	if cd.channelCounting[data.Uuid] == nil {
		cd.channelCounting[data.Uuid] = make(map[CallDirection]int)
	}
	cd.channelCounting[data.Uuid][data.Direction] = 1
}

func (cd *CallData) RemoveStream(data CallMetadata) bool {

	delete(cd.channelCounting[data.Uuid], data.Direction)
	if len(cd.channelCounting[data.Uuid]) == 0 {
		delete(cd.channelCounting, data.Uuid)
		return true // signal last channel has stopped
	}
	return false
}

func NewCallData() *CallData {
	return &CallData{channelCounting: make(map[string]map[CallDirection]int)}
}
