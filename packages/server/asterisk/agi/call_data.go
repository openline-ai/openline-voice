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
	streamMapping   map[string]CallMetadata
	channelCounting map[string]map[CallDirection]int
}

func (cd *CallData) AddStream(streamUuid string, data CallMetadata) {
	cd.streamMapping[streamUuid] = data
	if cd.channelCounting[data.Uuid] == nil {
		cd.channelCounting[data.Uuid] = make(map[CallDirection]int)
	}
	cd.channelCounting[data.Uuid][data.Direction] = 1
}

func (cd *CallData) GetStream(streamUuid string) *CallMetadata {
	val, ok := cd.streamMapping[streamUuid]
	if !ok {
		return nil
	}
	return &val
}

func (cd *CallData) RemoveStream(streamUuid string) bool {
	stream := cd.GetStream(streamUuid)
	if stream != nil {
		delete(cd.streamMapping, streamUuid)
		delete(cd.channelCounting[stream.Uuid], stream.Direction)
		if len(cd.channelCounting[stream.Uuid]) == 0 {
			delete(cd.channelCounting, stream.Uuid)
			return true // signal last channel has stopped
		}
	}
	return false
}

func NewCallData() *CallData {
	return &CallData{streamMapping: make(map[string]CallMetadata), channelCounting: make(map[string]map[CallDirection]int)}
}
