const VoiceResponse = require('twilio').twiml.VoiceResponse;
const {fetch} = require('./fetch')

exports.welcome = function welcome() {
    const voiceResponse = new VoiceResponse();

    const gather = voiceResponse.gather({
        action: '/menu',
        numDigits: '1',
        method: 'POST',
    });

    gather.say({loop: 3},
        'Thanks for calling Openline. ' +
        'Please press 1 to dial an extension. ' +
        'Press 2 to talk to an Openline representative.',
    );

    return voiceResponse.toString();
};

exports.menu = function menu(digit) {
    const optionActions = {
        '1': dialExtension,
        '2': huntGroup,
    };

    return (optionActions[digit])
        ? optionActions[digit]()
        : redirectWelcome();
};

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
        {voice: 'alice', language: 'en-GB'},
        'Please introduce the extension you would like to dial',
    );

    return twiml.toString();
}

exports.dial = function dial(digits) {
    const twiml = new VoiceResponse();

    twiml.dial().sip(`sip:+${digits}@kamailio.openline.ninja`);

    return twiml.toString();
};

/**
 * Returns a TwiML to interact with the client
 * @return {String}
 */
async function huntGroup() {
    const res = await fetch()
    let extensionLevelMapping = res.rows.reduce(reducer, {});
    return hungGroupHandler(extensionLevelMapping);
}

const reducer = (map, row) => {
    if (map[row.priority] == null) {
         map[row.priority] = [{'call_type': row.call_type, 'e164': row.e164}];
    } else {
         map[row.priority].push({'call_type': row.call_type, 'e164': row.e164});
    }
    return map;
};

function hungGroupHandler(extension_level_mapping) {
    const twiml = new VoiceResponse();
    const levels = Object.keys(extension_level_mapping).sort();

    for (let i = 0; i < levels.length; i++) {
        const level = levels[i];
        let dial = twiml.dial({timeout: process.env.CALL_TIMEOUT});
        const entry = extension_level_mapping[level];
        for (let j = 0; j < entry.length; j++) {
            let entryElement = entry[j];
            let e164 = entryElement.e164;
            if (entryElement.call_type === 'sip') {
                dial.sip(`sip:${e164}@kamailio.openline.ninja`);
            } else if (entryElement.call_type === 'pstn') {
                dial.number(e164);
            } else {
                console.log("Invalid call type: " + entry.call_type)
            }
        }
    }
    console.log(twiml.toString())
    return twiml.toString();
}

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