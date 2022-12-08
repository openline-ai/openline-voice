#!/usr/bin/python3

import grpc
from proto import messagestore_pb2
from proto import messagestore_pb2_grpc
import logging
from pydub import AudioSegment
from pydub.silence import detect_nonsilent
import whisper
import itertools
import sys
import os
import dumper
from twisted.protocols import sip
from google.protobuf.timestamp_pb2 import Timestamp
import time

def makeChunks(audio_segment_in, audio_segment_out, output_ranges_in, output_ranges_out):
    # from the itertools documentation
    def pairwise(iterable):
        "s -> (s0,s1), (s1,s2), (s2, s3), ..."
        a, b = itertools.tee(iterable)
        next(b, None)
        return zip(a, b)

    for range_i, range_ii in pairwise(output_ranges_in):
        last_end = range_i[1]
        next_start = range_ii[0]
        if next_start < last_end:
            range_i[1] = (last_end+next_start)//2
            range_ii[0] = range_i[1]

    for range_i, range_ii in pairwise(output_ranges_out):
        last_end = range_i[1]
        next_start = range_ii[0]
        if next_start < last_end:
            range_i[1] = (last_end+next_start)//2
            range_ii[0] = range_i[1]

    in_chunks = [{"start":start, "direction":"in", "chunks":audio_segment_in[max(start, 0): min(end, len(audio_segment_in))]} for start, end in output_ranges_in]
    out_chunks = [{"start":start, "direction":"out", "chunks":audio_segment_out[max(start, 0): min(end, len(audio_segment_out))]} for start, end in output_ranges_out]

    def sortStart(val):
        return val['start']

    chunks = in_chunks + out_chunks
    chunks.sort(key=sortStart)
    return chunks

def sendMessage(user: str, contact: str, isInbound: bool, msgTime: Timestamp, msg: str):
    if isInbound:
        direction=messagestore_pb2.INBOUND
    else:
        direction=messagestore_pb2.OUTBOUND
    with grpc.insecure_channel(os.getenv('MESSAGE_STORE')) as channel:
        stub = messagestore_pb2_grpc.MessageStoreServiceStub(channel)
        response = stub.saveMessage(messagestore_pb2.Message(message=msg,
                                                             username=contact,
                                                             channel=messagestore_pb2.VOICE,
                                                             direction=direction,
                                                             time=msgTime,
                                                             type=messagestore_pb2.MESSAGE
                                                             ))
        print(("Message created with id: " + str(response.id)))

def computeIsInbound(direction: str, fromPstn: bool):
    if direction == "in":
        return fromPstn
    else:
        return not fromPstn

def run():
    n = len(sys.argv)
    if n != 5:
        print("Incorrect number of arguments passed got %d expected 5" % n)
        exit(-1)

    fromUrl = sip.parseURL(sys.argv[2])
    toUrl = sip.parseURL(sys.argv[3])

    if fromUrl.username and fromUrl.username.startswith("+"):
        contact=fromUrl
        user=toUrl
        fromPstn=True
    else:
        contact=toUrl
        user=fromUrl
        fromPstn=False

    callTime = int(sys.argv[4])


    inFile = "/var/spool/asterisk/monitor/"+sys.argv[1] + "_in.wav"
    outFile = "/var/spool/asterisk/monitor/" + sys.argv[1] + "_out.wav"
    soundIn = AudioSegment.from_wav(inFile)
    soundOut = AudioSegment.from_wav(outFile)

    print("Call duration %f" % soundIn.duration_seconds)

    keep_silence = 100
    min_silence_len = 500
    silence_thresh = -40
    seek_step = 1

    output_ranges_in = [
        [ start - keep_silence, end + keep_silence ]
        for (start,end)
            in detect_nonsilent(soundIn, min_silence_len, silence_thresh, seek_step)
    ]
    print ("Got %d messages " % len(output_ranges_in))

    output_ranges_out = [
        [ start - keep_silence, end + keep_silence ]
        for (start,end)
            in detect_nonsilent(soundOut, min_silence_len, silence_thresh, seek_step)
    ]
    print ("Got %d messages " % len(output_ranges_out))
    model = whisper.load_model("base")

    audio_chunks = makeChunks(soundIn, soundOut, output_ranges_in, output_ranges_out)
    #print(dumper.dump(audio_chunks))
    for element in audio_chunks:
        outputFileName = "%s_%d.wav"%(element['direction'],element['start'])
        element['chunks'].export(outputFileName, format="wav")
        result = model.transcribe(outputFileName)
        os.remove(outputFileName)
        if result["text"] != "":
            msgTime = Timestamp(seconds=callTime+int(element['start']/1000), nanos=0)
            sendMessage(user.username, contact.username, computeIsInbound(element['direction'], fromPstn), msgTime, result["text"])
            print("At %d seconds %s said: %s " % (element['start']/1000,element['direction'],result["text"]))
    os.remove(inFile)
    os.remove(outFile)

if __name__ == '__main__':
    logging.basicConfig()
    run()
