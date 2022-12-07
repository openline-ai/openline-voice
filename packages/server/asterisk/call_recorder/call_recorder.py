#!/usr/bin/python3

import grpc
from proto import messagestore_pb2
from proto import messagestore_pb2_grpc
import logging

def run():
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