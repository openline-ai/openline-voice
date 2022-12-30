const twilio = require('twilio');
const {fetchMapping} = require('./fetch')
const {insertCall, getCalls, deleteCall, getCallHuntGroup} = require('./calls')
const {twCreateCall, twHangUpCall} = require('./twilio_wrapper')

const VoiceResponse = twilio.twiml.VoiceResponse;

exports.welcome = function welcome() {
    const twiml = new VoiceResponse();

    const gather = twiml.gather({
        action: '/menu',
        numDigits: '1',
        method: 'POST',
    });

    gather.say({loop: 3},
        'Thanks for calling Openline. Please press 1 to talk to Sales. Press 2 to talk to Support. '
    );

    return twiml.toString();
};

exports.menu = function menu(digit, from, callSid) {
    const optionActions = {
        '1': dialHuntGroup,
        '2': dialHuntGroup,
        '3': dialExtension,
    };

    return (optionActions[digit])
        ? optionActions[digit](digit, from, callSid)
        : redirectWelcome();
};

exports.voicemail = function voicemail(digit) {
    const optionActions = {
        '*': recordVoicemail
    };

    return (optionActions[digit])
        ? optionActions[digit]()
        : redirectWelcome();
};

exports.dial = function dial(digits) {
    const twiml = new VoiceResponse();

    twiml.dial().sip(`sip:+${digits}@kamailio.openline.ninja`);

    return twiml.toString();
};

exports.dial_queue = function dial_queue() {
    console.log('dial_queue')
    const twiml = new VoiceResponse();
    const dial = twiml.dial({timeout: process.env.CALL_TIMEOUT});
    // the name of the queue needs to be the same as the enqueue in the huntgrouphandler function
    dial.queue('queue_name')

    return twiml.toString();
};

const reducer = (map, activeCall, sid) => {
    if (activeCall.calls)
        if (map[activeCall.priority] == null) {
            map[row.priority] = [{'call_type': row.call_type, 'e164': row.e164}];
        } else {
            map[row.priority].push({'call_type': row.call_type, 'e164': row.e164});
        }
    return map;
};

exports.events = function events(body) {
    console.log('events => ' + 'to: ' + body.To + ' call: ' + body.CallSid + " status: " + body.CallStatus)

    if (body.CallStatus === 'in-progress') {
        return getCalls(body.CallSid)
            .then((res) => {
            if (res.rows?.length > 0) {
                for (let row of res.rows) {
                    if (row.callsid === body.CallSid) continue
                    twHangUpCall(body.CallSid)
                        .then(() => {
                            console.log('Call hangup successfully')
                            deleteCall(body.CallSid)
                                .then(() => console.log('Call delete successfully'))
                                .catch((err) => console.log(err))
                        })
                }
            }
        });
    }

    if (body.CallStatus === 'failed' || body.CallStatus === 'no-answer') {
        //delete calls[body.CallSid];
    }

    // if (Object.keys(calls).length === 0) {
    //     console.log('Update call')
    //     client.calls(sid)
    //         .update({twiml: askToRecordVoicemail()})
    //         .then(call => console.log('completed: ' + call.to));
    // }
}

exports.recording_events = function recording_events() {
    console.log('recording_events')
};

/**
 * Returns an xml with the redirect
 * @return {String}
 */
function redirectWelcome() {
    const twiml = new VoiceResponse();

    twiml.say('Returning to the main menu', {
        voice: 'alice',
        language: 'en-GB',
    });

    twiml.redirect('/welcome');

    return twiml.toString();
}

/**
 * Returns Twiml
 * @return {String}
 */
function dialExtension() {
    const twiml = new VoiceResponse();

    const gather = twiml.gather({
        action: '/dial',
        numDigits: '3',
        method: 'POST',
    });

    gather.say(
        'Please introduce the extension you would like to dial',
    );

    return twiml.toString();
}

/**
 * Returns Twiml
 * @return {String}
 */
function recordVoicemail() {
    const twiml = new VoiceResponse();
    twiml.record({
        playBeep: true, timeout: process.env.CALL_TIMEOUT,
        recordingStatusCallback: `${process.env.BASE_URL}/recording_events`,
        recordingStatusCallbackEvent: 'completed'
    })

    return twiml.toString();
}

function askToRecordVoicemail() {
    const twiml = new VoiceResponse();

    const gather = twiml.gather({
        action: '/voicemail',
        numDigits: '1',
        method: 'POST',
    });

    gather.say(
        'Sorry, none of our agents were available. ' +
        'Press 1 to record a message after the beep: '
    );

    return twiml.toString();
}

/**
 * Returns a TwiML to interact with the client
 * @return {String}
 */
async function dialHuntGroup(digit, from, callSid) {
    const res = await fetchMapping(digit)
    return hungGroupHandler(res.rows, from, callSid)
}

function hungGroupHandler(rows, from, parentSid) {
    const twiml = new VoiceResponse();

    if (rows.length > 0) {
        for (let row of rows) {
            let toDial = row.call_type === 'sip' ? `sip:${row.e164}@kamailio.openline.ninja` : row.e164;
            twCreateCall(toDial, from)
                .then((call) => {
                    insertCall(call.sid, parentSid)
                        .then(() => console.log('Call to: ' + call.to + ' saved'))
                        .catch((err) => console.log('Error while saving call to: ' + call.to + ': ' + err));
                })
                .catch((err) => console.log('Error while creating call: ' + err));
        }
    }

    twiml.enqueue('queue_name');
    return twiml.toString();
}
