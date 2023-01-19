const twilio = require("twilio");
const VoiceResponse = require('twilio').twiml.VoiceResponse;

const client = twilio(process.env.TWILIO_ACCOUNT_ID, process.env.TWILIO_AUTH_TOKEN);

exports.createCall = (tenant_name, to, from) => {
    return client.calls.create({
        url: `${process.env.BASE_URL}/${tenant_name}_dial_queue`,
        to: to,
        from: from,
        statusCallbackEvent: ['initiated', 'ringing', 'answered', 'completed'],
        statusCallback: `${process.env.BASE_URL}/${tenant_name}_events`,
        timeout: process.env.CALL_TIMEOUT,
    })
}

exports.cancelCall = (sid) => {
    return client.calls(sid).update({status: 'canceled'})
}

exports.recordVoicemail = (sid, say) => {
    return client.calls(sid).update({twiml: recordVoicemailTwiml(say)})
}

function recordVoicemailTwiml(say) {
    const twiml = new VoiceResponse();

    twiml.say(say);

    twiml.record({
        playBeep: true, timeout: process.env.CALL_TIMEOUT,
        recordingStatusCallback: `${process.env.BASE_URL}/recording_events`,
        recordingStatusCallbackEvent: 'completed'
    })

    return twiml.toString();
}

/**
 * Returns an xml with the redirect
 * @return {String}
 */
exports.redirectWelcomeTwiml = (tenant_name) => {
    const twiml = new VoiceResponse();

    twiml.say('Returning to the main menu', {
        voice: 'alice',
        language: 'en-GB',
    });

    twiml.redirect(`/${tenant_name}/welcome`);

    return twiml.toString();
}

exports.dialTwiml = (digits) => {
    const twiml = new VoiceResponse();
    twiml.dial().sip(`sip:+${digits}@${process.env.KAMAILIO_DOMAIN}}`);
    return twiml.toString();
};

/**
 * Returns Twiml
 * @return {String}
 */
exports.dialExtensionTwiml = () => {
    const twiml = new VoiceResponse();

    const gather = twiml.gather({
        action: '/dial',
        numDigits: '3',
        method: 'POST',
    });

    gather.say(
        'Please introduce the extension you would like to dial. ',
    );

    return twiml.toString();
}

