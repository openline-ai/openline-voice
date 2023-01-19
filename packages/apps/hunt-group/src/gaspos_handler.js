const twilio = require('twilio');
const VoiceResponse = twilio.twiml.VoiceResponse;

const {redirectWelcomeTwiml, dialExtensionTwiml} = require('./common/cpaas')
const {huntGroupStart} = require('./common/hung_group')

exports.welcome_gaspos = function welcome_gaspos() {
    const twiml = new VoiceResponse();

    const gather = twiml.gather({
        action: '/menu',
        numDigits: '1',
        method: 'POST',
    });

    gather.say({loop: 3},
        'Thanks for calling GasPos. Please press 1 to talk to Sales. Press 2 to talk to Support. '
    );

    return twiml.toString();
};

exports.menu = function menu(extension, from, callSid) {
    const optionActions = {
        '1': huntGroupStart,
        '2': huntGroupStart,
        '3': dialExtensionTwiml(),
    };

    return (optionActions[extension])
        ? optionActions[extension]('gaspos', extension, from, callSid)
        : redirectWelcomeTwiml('gaspos');
};

exports.gaspos_events = async (body) => {
    console.log('events => ' + 'from: ' + body.From + 'to: ' + body.To + ' call: ' + body.CallSid + " status: " + body.CallStatus)
}

exports.status = function status(body) {
    console.log('events => ' + 'to: ' + body.To + ' call: ' + body.CallSid + " status: " + body.CallStatus)
}

exports.gaspos_dial_queue = () => {
    const twiml = new VoiceResponse();
    const dial = twiml.dial({timeout: process.env.CALL_TIMEOUT});
    dial.queue('gaspos_queue_name')

    return twiml.toString();
};

exports.gaspos_recording_events = function gaspos_recording_events(body) {
    console.log('body => ' + JSON.stringify(body))
}
