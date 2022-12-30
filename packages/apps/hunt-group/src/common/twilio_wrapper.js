const twilio = require("twilio");

const client = twilio(process.env.TWILIO_ACCOUNT_ID, process.env.TWILIO_AUTH_TOKEN);

exports.twCreateCall = (to, from) => {
    return client.calls.create({
        url: `${process.env.BASE_URL}/dial_queue`,
        to: to,
        from: from,
        statusCallbackEvent: ['initiated', 'ringing', 'answered', 'completed'],
        statusCallback: `${process.env.BASE_URL}/events`,
        timeout: process.env.CALL_TIMEOUT,
    })
}

exports.twHangUpCall = (sid) => {
    return client.calls(sid)
        .update({twiml: hangUpTwiml()})
}

exports.twVoiceMailCall = (sid) => {
    client.calls(sid)
        .update({twiml: voiceMailTwiml()})
}

function hangUpTwiml() {
    const twiml = new VoiceResponse();
    twiml.hangup()
    return twiml.toString()
}

function voiceMailTwiml() {
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

