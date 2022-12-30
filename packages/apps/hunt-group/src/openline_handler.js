const twilio = require('twilio');
const VoiceResponse = twilio.twiml.VoiceResponse;

const {getNextPriorityInGroup} = require('./common/fetch')
const {getCalls, deleteCall} = require('./common/calls')
const {hangUpCall, recordVoicemail} = require('./common/cpaas')
const {huntGroupDial} = require('./common/hung_group')

exports.welcome_openline = function welcome_openline(from, callSid) {
    return huntGroupDial('openline', `${process.env.OPENLINE_HUNT_GROUP_EXT}`, '1', from, callSid).then(() => {
        const twiml = new VoiceResponse();
        twiml.say('Thanks for calling Openline. Please wait while we find someone for you to talk to. ');
        twiml.enqueue('queue_name');
        return twiml.toString();
    })
};

exports.openline_events = async function openline_events(body) {
    console.log('events => ' + 'from: ' + body.From + 'to: ' + body.To + ' call: ' + body.CallSid + " status: " + body.CallStatus)

    if (body.CallStatus === 'in-progress') {
        const calls = getCalls(body.CallSid);
        const hangupPromises = [];
        for (let callsid of Object.keys(calls)) {
            if (callsid === body.CallSid) continue;
            hangupPromises.push(hangUpCall(body.CallSid));
        }
        return Promise.all(hangupPromises)
            .then(() => deleteCall(body.CallSid))
            .catch((err) => console.log('Error while hanging up calls: ' + err));
    }

    if (body.CallStatus === 'failed' || body.CallStatus === 'no-answer') {
        let calls = getCalls(body.CallSid);
        let keys = Object.keys(calls);
        if (keys.length === 1) {
            let call = calls[keys[0]];
            return getNextPriorityInGroup(call.hunt_group_id, call.priority).then((res) => {
                if (res.rows.length === 0) {
                    // end of hunt group -> go to voicemail
                    recordVoicemail(call.parent_sid, `${process.env.OPENLINE_VOICEMAIL}`)
                        .then(() => deleteCall(body.CallSid))
                        .catch((err) => console.log('Error while recording voicemail: ' + err));
                } else {
                    // dial next extension
                    let priority = res.rows[0].priority;
                    huntGroupDial('openline', `${process.env.OPENLINE_HUNT_GROUP_EXT}`, priority, body.From, call.parent_sid)
                        .then(() => deleteCall(body.CallSid))
                        .catch((err) => console.log('Error while dialing hunt group: ' + err));
                }
            }).catch((err) => console.log('Error while fetching next extension in group: ' + err));
        } else {
            deleteCall(body.CallSid)
        }
    }
}

exports.status = function status(body) {
    console.log('events => ' + 'to: ' + body.To + ' call: ' + body.CallSid + " status: " + body.CallStatus)
}

exports.openline_dial_queue = function dial_queue() {
    const twiml = new VoiceResponse();
    const dial = twiml.dial({timeout: process.env.CALL_TIMEOUT});
    dial.queue('openline_queue_name')

    return twiml.toString();
};

exports.openline_recording_events = function openline_recording_events(body) {
    console.log('body => ' + JSON.stringify(body))
}