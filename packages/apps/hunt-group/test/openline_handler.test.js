const {welcome_openline, openline_events, status} = require('../src/openline_handler');
require('dotenv').config()

jest.mock('../src/common/cpaas')
jest.mock('../src/common/fetch');

describe('OpenlineHandler#Welcome', () => {
    const cpaas = require('../src/common/cpaas');
    const calls = require('../src/common/callTracker');

    const spyCreateCall = jest.spyOn(cpaas, 'createCall');
    const spyRecordVoicemail = jest.spyOn(cpaas, 'recordVoicemail');
    const spyCancelCall = jest.spyOn(cpaas, 'cancelCall');

    afterEach(() => {
        spyCreateCall.mockClear();
        spyCancelCall.mockClear();
        for(let callSid of calls.getAllTrackedCallSids()) {
            delete calls.removeCall(callSid);
        }
    });

    it('should serve TwiML with welcome message and dial huntgroup', async () => {
        let from = '+1';
        const twiml = await welcome_openline(from, 'CA1')
        const count = countWord(twiml);

        // TwiML verbs
        expect(count('Say')).toBe(2);
        // TwiML content
        expect(twiml).toContain('Thanks for calling Openline. Please wait while we find someone for you to talk to. ');

        expect(spyCreateCall).toHaveBeenCalledWith('openline', 'sip:+901@kamailio-development.openline.ninja', from);
        expect(spyCreateCall).toHaveBeenCalledWith('openline', 'sip:+902@kamailio-development.openline.ninja', from);

        let callSids = calls.getAllTrackedCallSids();

        expect(callSids.length).toBe(2);

        let firstSid = callSids[0];
        let secondSid = callSids[1];

        await openline_events({CallSid: firstSid, CallStatus: 'in-progress'})
        expect(spyCancelCall).toHaveBeenCalledWith(secondSid);
        expect(Object.keys(calls.getActiveHuntGroupCalls(secondSid)).length).toBe(1);

        await openline_events({CallSid: firstSid, CallStatus: 'completed'})
        expect(Object.keys(calls.getActiveHuntGroupCalls(firstSid)).length).toBe(0);
    });

    it('should serve dial entire hunt group and go to voicemail', async () => {
        await welcome_openline('+1', 'CA1');

        expect(spyCreateCall).toHaveBeenCalledWith('openline', 'sip:+901@kamailio-development.openline.ninja', '+1');
        expect(spyCreateCall).toHaveBeenCalledWith('openline', 'sip:+902@kamailio-development.openline.ninja', '+1');

        let callSids = calls.getAllTrackedCallSids();
        await openline_events({To: '+901', CallSid: callSids[0], From: '+1', CallStatus: 'no-answer'})
        spyCreateCall.mockClear();
        await openline_events({To: '+902', CallSid: callSids[1], From: '+1', CallStatus: 'no-answer'})

        expect(spyCreateCall).toHaveBeenCalledWith('openline', 'sip:+903@kamailio-development.openline.ninja', '+1');
        expect(spyCreateCall).toHaveBeenCalledWith('openline', 'sip:+904@kamailio-development.openline.ninja', '+1');

        jest.mock('../src/common/fetch', () => {
            return {
                getNextExtensionInTenant: jest.fn().mockImplementation(() => {
                    return Promise.resolve({rows: []})
                })
            }
        });

        callSids = calls.getAllTrackedCallSids();
        await openline_events({To: '+903', CallSid: callSids[2], From: '+1', CallStatus: 'no-answer'})
        await openline_events({To: '+904', CallSid: callSids[3], From: '+1', CallStatus: 'no-answer'})

        expect(spyRecordVoicemail).toHaveBeenCalledTimes(1);
    });

    it('should stop a hunt group on close event', async () => {
        let from = '+1';
        let parentSid = 'CA1'
        const twiml = await welcome_openline(from, parentSid)
        const count = countWord(twiml);

        // TwiML verbs
        expect(count('Say')).toBe(2);
        // TwiML content
        expect(twiml).toContain('Thanks for calling Openline. Please wait while we find someone for you to talk to. ');

        expect(spyCreateCall).toHaveBeenCalledWith('openline', 'sip:+901@kamailio-development.openline.ninja', from);
        expect(spyCreateCall).toHaveBeenCalledWith('openline', 'sip:+902@kamailio-development.openline.ninja', from);

        let callSids = calls.getAllTrackedCallSids();

        expect(callSids.length).toBe(2);

        await status({CallSid: parentSid, CallStatus: 'completed'})
        expect(spyCancelCall).toHaveBeenCalledTimes(2);
        expect(calls.getAllTrackedCallSids().length).toBe(0);
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