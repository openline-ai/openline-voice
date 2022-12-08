#!/usr/bin/python3

import grpc
from proto import messagestore_pb2
from proto import messagestore_pb2_grpc
import logging
import wave
import sys

def run():
    n = len(sys.argv)
    if n != 4:
        print("Incorrect number of arguments passed got %d expected 4" % n)
        exit(-1)

    with wave.open("/var/spool/asterisk/monitoring/" + sys.argv[1]) as mywav:
        duration_seconds = mywav.getnframes() / mywav.getframerate()
        print(f"Length of the WAV file: {duration_seconds:.1f} s")

    print("Call From %s to %s" % (sys.argv[2], sys.argv[3]))
    with grpc.insecure_channel('localhost:9009') as channel:
        stub = messagestore_pb2_grpc.MessageStoreServiceStub(channel)
        response = stub.saveMessage(messagestore_pb2.Message(message='Call of duration 5 minutes',
                                                             username='+32485112970',
                                                             channel=messagestore_pb2.VOICE,
                                                             direction=messagestore_pb2.INBOUND,
                                                             type=messagestore_pb2.MESSAGE
                                                             ))
        print(("Message created with id: " + str(response.id)))

if __name__ == '__main__':
    logging.basicConfig()
    run()
