const twilio = require('twilio');
const VoiceResponse = twilio.twiml.VoiceResponse;

const {getNextPriorityInGroup} = require('./common/fetch')
const {
    getActiveHuntGroupCalls,
    removeCall,
    isTrackedCall,
    getChildCalls,
    setCallStatus,
    getTrackedCall
} = require('./common/callTracker')

const {cancelCall, recordVoicemail} = require('./common/cpaas')
const {huntGroupStart} = require('./common/hung_group')

exports.welcome_openline = function welcome_openline(from, callSid) {
    return huntGroupStart('openline', `${process.env.OPENLINE_HUNT_GROUP_EXT}`, '1', from, callSid).then(() => {
        const twiml = new VoiceResponse();
        twiml.say('Thanks for calling Openline. Please wait while we find someone for you to talk to. ');
        twiml.enqueue(`${process.env.OPENLINE_QUEUE_NAME}`);
        return twiml.toString();
    })
};

exports.openline_events = function openline_events(body) {
    console.log('events => ' + 'from: ' + body.From + 'to: ' + body.To + ' call: ' + body.CallSid + " status: " + body.CallStatus)
    setCallStatus(body.CallSid, body.CallStatus);

    if (body.CallStatus === 'completed' || body.CallStatus === 'canceled') {
        return;
    }

    if (body.CallStatus === 'in-progress') {
        const calls = getActiveHuntGroupCalls(body.CallSid);
        const promises = [];
        for (let callsid of Object.keys(calls)) {
            if (callsid === body.CallSid) continue;
            console.log('Complete call => ' + callsid)
            promises.push(cancelCall(callsid)
                .then((call) => setCallStatus(call.sid, call.status))
                .catch((err) => console.log('Error while trying to complete calls: ' + err)));
        }
        return Promise.all(promises);
    }

    if (body.CallStatus === 'failed' || body.CallStatus === 'no-answer' || body.CallStatus === 'busy') {
        let activeCalls = getActiveHuntGroupCalls(body.CallSid);
        if (Object.keys(activeCalls).length === 0) {
            let trackedCall = getTrackedCall(body.CallSid);
            return getNextPriorityInGroup(trackedCall.hunt_group_id, trackedCall.priority)
                .then((res) => {
                    if (res.rows.length === 0) {
                        recordVoicemail(trackedCall.parent_sid, `${process.env.OPENLINE_VOICEMAIL}`)
                            .then(call => console.log('Voicemail recording started: ' + call.sid))
                            .catch((err) => console.log('Error while recording voicemail: ' + err));
                    } else {
                        huntGroupStart('openline', `${process.env.OPENLINE_HUNT_GROUP_EXT}`, res.rows[0].priority, body.From, trackedCall.parent_sid)
                            .then(() => console.log('hung group started: ' + `${process.env.OPENLINE_HUNT_GROUP_EXT}` + ' priority: ' + res.rows[0].priority))
                            .catch((err) => console.log('Error while dialing hunt group: ' + err));
                    }
                }).catch((err) => console.log('Error while fetching next extension in group: ' + err));
        }
    }
}

exports.status = function status(body) {
    console.log('events => ' + 'from: ' + body.From + 'to: ' + body.To + ' call: ' + body.CallSid + " status: " + body.CallStatus)
    if (body.CallStatus === 'completed' || body.CallStatus === 'canceled') {
        let childCalls = getChildCalls(body.CallSid);

        for (let callsid of Object.keys(childCalls)) {
            console.log('Cancel call => ' + callsid)
            cancelCall(callsid)
                .then((call) => removeCall(call.sid))
                .catch((err) => console.log('Error while canceling call: ' + err));
        }
    }
}

exports.openline_dial_queue = (body) => {
    console.log('Call Queued' + body.CallSid)
    const twiml = new VoiceResponse();
    const dial = twiml.dial({timeout: process.env.CALL_TIMEOUT});
    dial.queue(`${process.env.OPENLINE_QUEUE_NAME}`);
    return twiml.toString();
};

exports.openline_recording_events = function openline_recording_events(body) {
    console.log('body => ' + JSON.stringify(body))
}