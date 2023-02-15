package main

import (
	"context"
	"fmt"
	"github.com/CyCoreSystems/audiosocket"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"io"
	"log"
	"net"
	"os"
)

const listenAddr = ":8090"

func Listen(ctx context.Context, streamMap *CallData) error {
	l, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return errors.Wrapf(err, "failed to bind listener to socket %s", listenAddr)
	}

	log.Printf("listening on %s for recordings", listenAddr)
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println("failed to accept new connection:", err)
			continue
		}

		go Handle(ctx, conn, streamMap)
	}
}

func getCallID(c net.Conn) (uuid.UUID, error) {
	m, err := audiosocket.NextMessage(c)
	if err != nil {
		return uuid.Nil, err
	}

	if m.Kind() != audiosocket.KindID {
		return uuid.Nil, errors.Errorf("invalid message type %d getting CallID", m.Kind())
	}

	return uuid.FromBytes(m.Payload())
}

func Handle(ctx context.Context, c net.Conn, streamMap *CallData) {
	var err error
	var id uuid.UUID

	defer func() {
		// Tell AudioSocket to shut down, if it is still up
		c.Write(audiosocket.HangupMessage()) // nolint: errcheck
	}()

	id, err = getCallID(c)
	if err != nil {
		log.Println("failed to get call ID:", err)
		return
	}
	log.Printf("processing call %s\n", id.String())
	cd := streamMap.GetStream(id.String())

	if cd == nil {
		log.Println("Unknown Stream ID")
		return
	}

	log.Printf("Stream for %s direction %s\n", cd.Uuid, cd.Direction)

	f, err := os.Create("/tmp/" + cd.Uuid + "-" + string(cd.Direction) + ".s16")
	if err != nil {
		fmt.Println("Unable to open file for writing: " + err.Error())
	}

	defer f.Close()

	for ctx.Err() == nil {
		m, err := audiosocket.NextMessage(c)
		if errors.Cause(err) == io.EOF {
			log.Println("audiosocket closed")
			return
		}
		if m.Kind() == audiosocket.KindHangup {
			log.Println("audiosocket received hangup command")
			return
		}
		if m.Kind() == audiosocket.KindError {
			log.Println("error from audiosocket")
			continue
		}
		if m.Kind() != audiosocket.KindSlin {
			log.Println("ignoring non-slin message", m.Kind())
			continue
		}
		if m.ContentLength() < 1 {
			log.Println("no content")
			continue
		}
		f.Write(m.Payload())
	}
	return
}
