package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/openline-ai/openline-oasis/packages/server/channels-api/model"
	"io"
	"log"
	"net/http"
	"time"
)

type VConMessage struct {
	Sender  *model.VConParty
	Message string
}
type VConEventPublisher struct {
	conf           *RecordServiceConfig
	messageChannel chan VConMessage
	killChannel    chan bool
	uuid           string
	first          bool
	parties        []*model.VConParty
	transcript     []string
}

func NewVConEventPublisher(conf *RecordServiceConfig, uuid string, from *model.VConParty, to *model.VConParty) *VConEventPublisher {
	return &VConEventPublisher{
		conf:           conf,
		messageChannel: make(chan VConMessage),
		killChannel:    make(chan bool),
		uuid:           uuid,
		first:          true,
		parties:        []*model.VConParty{from, to},
		transcript:     make([]string, 0),
	}
}

func (v *VConEventPublisher) SendMessage(sender *model.VConParty, message string) {
	v.messageChannel <- VConMessage{
		Sender:  sender,
		Message: message,
	}
}

func partyToString(p *model.VConParty) string {
	if p.Name != nil {
		return *p.Name
	} else if p.Mailto != nil {
		return *p.Mailto
	} else if p.Tel != nil {
		return *p.Tel
	}
	return ""
}

func (v *VConEventPublisher) publish(message *model.VCon) {
	// send an http request to the channels-api
	// to publish the message
	body, err := json.Marshal(message)
	if err != nil {
		log.Printf("client: could not marshal request: %s\n", err)
		return
	}
	bodyReader := bytes.NewReader(body)

	log.Printf("client: sending request: \n%s\n", string(body))
	requestURL := fmt.Sprintf("%s/api/v1/vcon", v.conf.ChannelsApiService)
	req, err := http.NewRequest(http.MethodPost, requestURL, bodyReader)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Openline-VCon-Api-Key", v.conf.ChannelsApiKey)
	if err != nil {
		log.Printf("client: could not create request: %s\n", err)
		return
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("client: could not send request: %s\n", err)
		return
	}

	log.Printf("client: got response!\n")
	log.Printf("client: status code: %d\n", res.StatusCode)
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("client: could not read response body: %s\n", err)
		return
	}
	log.Printf("client: response body: %s\n", string(resBody))

}

func (v *VConEventPublisher) Kill() {
	v.killChannel <- true
}

func compareParties(p1 *model.VConParty, p2 *model.VConParty) bool {
	if p1.Name != nil && p2.Name != nil {
		return *p1.Name == *p2.Name
	} else if p1.Mailto != nil && p2.Mailto != nil {
		return *p1.Mailto == *p2.Mailto
	} else if p1.Tel != nil && p2.Tel != nil {
		return *p1.Tel == *p2.Tel
	}
	return false
}

func (v *VConEventPublisher) Run() {
	log.Printf("client: starting vcon event publisher: %s\n", v.uuid)
	for {
		select {
		case message := <-v.messageChannel:
			vconf := &model.VCon{}
			if v.first {
				v.uuid = uuid.New().String()
				vconf.UUID = v.uuid
				v.first = false
			} else {
				vconf.UUID = uuid.New().String()
				vconf.Appended = &model.VConAppended{UUID: v.uuid}
			}
			vconf.Parties = make([]model.VConParty, len(v.parties))
			vconf.Parties[0] = *v.parties[0]
			vconf.Parties[1] = *v.parties[1]

			vconf.Dialog = make([]model.VConDialog, 1)
			vconf.Dialog[0] = model.VConDialog{
				Body:     message.Message,
				Type:     "text",
				MimeType: "text/plain",
				Encoding: "None",
				Start:    time.Now(),
			}
			// current convention is to put sender as the first index
			if compareParties(message.Sender, v.parties[0]) {
				vconf.Dialog[0].Parties = []int64{0, 1}
			} else {
				vconf.Dialog[0].Parties = []int64{1, 0}
			}
			v.transcript = append(v.transcript, partyToString(message.Sender)+": "+message.Message)
			v.publish(vconf)
		case _ = <-v.killChannel:
			log.Printf("client: killing vcon event publisher\n")
			return

		}
	}
}
