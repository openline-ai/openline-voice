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
	streamMapping map[string]CallMetadata
}

func (cd *CallData) AddStream(streamUuid string, data CallMetadata) {
	cd.streamMapping[streamUuid] = data
}

func (cd *CallData) GetStream(streamUuid string) *CallMetadata {
	val, ok := cd.streamMapping[streamUuid]
	if !ok {
		return nil
	}
	return &val
}

func (cd *CallData) RemoveStream(streamUuid string) {
	delete(cd.streamMapping, streamUuid)
}

func NewCallData() *CallData {
	return &CallData{make(map[string]CallMetadata)}
}
