const {welcome_openline, openline_events} = require('../src/openline_handler');
const calls = require("../src/common/calls");
require('dotenv').config()

jest.mock('../src/common/cpaas')
jest.mock('../src/common/fetch');

describe('OpenlineHandler#Welcome', () => {
    const cpaas = require('../src/common/cpaas');
    const calls = require('../src/common/calls');

    const spyCreateCall = jest.spyOn(cpaas, 'createCall');
    const spyRecordVoicemail = jest.spyOn(cpaas, 'recordVoicemail');
    const spyHangUpCall = jest.spyOn(cpaas, 'hangUpCall');

    afterEach(() => {
        spyCreateCall.mockClear();
        spyHangUpCall.mockClear();
        for(let callSid of calls.getCallSids()) {
            delete calls.deleteCall(callSid);
        }
    });

    it('should serve TwiML with welcome message and dial huntgroup', async () => {
        const twiml = await welcome_openline('+1', 'CA1')
        const count = countWord(twiml);

        // TwiML verbs
        expect(count('Say')).toBe(2);
        // TwiML content
        expect(twiml).toContain('Thanks for calling Openline. Please wait while we find someone for you to talk to. ');

        expect(spyCreateCall).toHaveBeenCalledWith('openline', 'sip:+901@kamailio-development.openline.ninja', '+1');
        expect(spyCreateCall).toHaveBeenCalledWith('openline', 'sip:+901@kamailio-development.openline.ninja', '+1');

        // +101 has picked up the call. We should cancel the other calls
        let callSids = calls.getCallSids();
        await openline_events({To: '+901', CallSid: callSids[0], CallStatus: 'in-progress'})
        expect(spyHangUpCall).toHaveBeenCalledTimes(1);
    });

    it('should serve dial entire hunt group and go to voicemail', async () => {
        await welcome_openline('+1', 'CA1');

        expect(spyCreateCall).toHaveBeenCalledWith('openline', 'sip:+901@kamailio-development.openline.ninja', '+1');
        expect(spyCreateCall).toHaveBeenCalledWith('openline', 'sip:+902@kamailio-development.openline.ninja', '+1');

        let callSids = calls.getCallSids();
        await openline_events({To: '+901', CallSid: callSids[0], From: '+1', CallStatus: 'no-answer'})

        spyCreateCall.mockClear();

        callSids = calls.getCallSids();
        await openline_events({To: '+902', CallSid: callSids[0], From: '+1', CallStatus: 'no-answer'})

        expect(spyCreateCall).toHaveBeenCalledWith('openline', 'sip:+903@kamailio-development.openline.ninja', '+1');
        expect(spyCreateCall).toHaveBeenCalledWith('openline', 'sip:+904@kamailio-development.openline.ninja', '+1');

        jest.mock('../src/common/fetch', () => {
            return {
                getNextExtensionInTenant: jest.fn().mockImplementation(() => {
                    return Promise.resolve({rows: []})
                })
            }
        });

        callSids = calls.getCallSids();
        await openline_events({To: '+903', CallSid: callSids[0], From: '+1', CallStatus: 'no-answer'})
        callSids = calls.getCallSids();
        await openline_events({To: '+904', CallSid: callSids[0], From: '+1', CallStatus: 'no-answer'})

        expect(spyRecordVoicemail).toHaveBeenCalledTimes(1);
    });
})


/**
 * Counts how many times a word is repeated
 * @param {String} paragraph
 * @return {String[]}
 */
function countWord(paragraph) {
    return (word) => {
        const regex = new RegExp(`\<${word}[ | \/?\>]|\<\/${word}?\>`);
        return (paragraph.split(regex).length - 1);
    };
}