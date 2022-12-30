const twilio = require('twilio');

const {fetchHuntGroupMappings} = require('./common/fetch')
const {insertCall, getCalls, deleteCall} = require('./common/calls')
const {twCreateCall, twHangUpCall} = require('./common/twilio_wrapper')

const VoiceResponse = twilio.twiml.VoiceResponse;

exports.welcome_openline = function welcome_openline(from, callSid) {
    return huntGroupDial('openline', '1', from, callSid).then(() => {
        const twiml = new VoiceResponse();
        twiml.say('Thanks for calling Openline. Please wait while we find someone to talk to. ');
        twiml.enqueue('queue_name');
        return twiml.toString();
    });
};

/**
 * Returns a TwiML to interact with the client
 * @return {String}
 */
function huntGroupDial(tenant_name, digit, from, parentSid) {
    return fetchHuntGroupMappings(tenant_name, digit).then((res) => {
        if (res.rows.length > 0) {
            for (let row of res.rows) {
                let toDial = row.call_type === 'sip' ? `sip:${row.e164}@kamailio.openline.ninja` : row.e164;
                twCreateCall(toDial, from)
                    .then((call) => {
                        console.log('Call created successfully: ' + call.sid)
                        insertCall(row.hunt_group_id, call.sid, parentSid)
                            .then(() => console.log('Call to: ' + call.to + ' saved'))
                            .catch((err) => console.log('Error while saving call: ' + call.to + ': ' + err));
                    })
                    .catch((err) => console.log('Error while creating call: ' + err));
            }
        } else {
            console.log('No hunt group mappings found: ' + tenant_name + '->'+ digit)
        }
    });
}

exports.events = async function events(body) {
    console.log('events => ' + 'to: ' + body.To + ' call: ' + body.CallSid + " status: " + body.CallStatus)

    if (body.CallStatus === 'in-progress') {
        return getCalls(body.CallSid)
            .then((res) => {
                if (res.rows?.length > 0) {
                    for (let row of res.rows) {
                        if (row.callsid === body.CallSid) continue;
                        twHangUpCall(body.CallSid)
                            .then(() => {
                                console.log('Call hangup successfully: ' + body.CallSid)
                                deleteCall(body.CallSid)
                                    .then(() => console.log('Call delete successfully'))
                                    .catch((err) => console.log(err))
                            }).catch((err) => console.log(err))
                    }
                }
            }).catch((err) => console.log(err));
    }


    if (body.CallStatus === 'failed' || body.CallStatus === 'no-answer') {
        return getCalls(body.CallSid).then((res) => {
            if (res.rows?.length === 1) {
                huntGroupDial('openline', '2', body.From, body.CallSid).then(() => {
                    console.log('Hunt group called')
                });
            }
            deleteCall(body.CallSid)
                .then(() => console.log('Call delete successfully'))
                .catch((err) => console.log('Call delete failed: ' + err))
        }).catch((err) => console.log('Unable to retrieve calls: ' + err));
    }
}

exports.dial_queue = function dial_queue() {
    console.log('dial_queue')
    const twiml = new VoiceResponse();
    const dial = twiml.dial({timeout: process.env.CALL_TIMEOUT});
    // the name of the queue needs to be the same as the enqueue in the huntgrouphandler function
    dial.queue('queue_name')

    return twiml.toString();
};
