const {welcome_openline, events} = require('../src/openline_handler');
require('dotenv').config()

jest.mock('../src/common/twilio_wrapper')
jest.mock('../src/common/fetch')
jest.mock('../src/common/calls')
describe('OpenlineHandler#Welcome', () => {
    const twilio_wrapper = require('../src/common/twilio_wrapper');
    const calls = require('../src/common/calls');

    const spyTwCreateCall = jest.spyOn(twilio_wrapper, 'twCreateCall');
    const spyTwHangup = jest.spyOn(twilio_wrapper, 'twHangUpCall');
    const spyInsertCalls = jest.spyOn(calls, 'insertCall');
    const spyDelete = jest.spyOn(calls, 'deleteCall');

    afterEach(() => {
        spyTwCreateCall.mockClear();
        spyTwHangup.mockClear();
        spyInsertCalls.mockClear();
        spyDelete.mockClear();
    });

    it('should serve TwiML with welcome message and dial huntgroup', () => {
        return welcome_openline('+1', 'CA1').then((twiml) => {
            const count = countWord(twiml);

            // TwiML verbs
            expect(count('Say')).toBe(2);
            // TwiML content
            expect(twiml).toContain('Thanks for calling Openline. Please wait while we find someone to talk to. ');

            expect(spyTwCreateCall).toHaveBeenCalledWith('sip:+101@kamailio.openline.ninja', '+1');
            expect(spyTwCreateCall).toHaveBeenCalledWith('sip:+102@kamailio.openline.ninja', '+1');
            // the calls were saved to our database
            expect(spyInsertCalls).toHaveBeenCalledTimes(2);

            // +101 has picked up the call. We should cancel the other calls
            return events({To: '+101', CallSid: 'CA2', CallStatus: 'in-progress'}).then(() => {
                expect(spyTwHangup).toHaveBeenCalledTimes(2);
                expect(spyDelete).toHaveBeenCalledTimes(2);
            });
        });
    });

    it('should serve dial entire hunt group and go to voicemail', () => {
        return welcome_openline('+1', 'CA1').then((twiml) => {
            expect(spyTwCreateCall).toHaveBeenCalledWith('sip:+101@kamailio.openline.ninja', '+1');
            expect(spyTwCreateCall).toHaveBeenCalledWith('sip:+102@kamailio.openline.ninja', '+1');
            // +101 has picked up the call. We should cancel the other calls
            return events({To: '+101', CallSid: 'CA2', CallStatus: 'no-answer'}).then(() => {
                expect(spyDelete).toHaveBeenCalledTimes(1);
                return events({To: '+102', CallSid: 'CA3', From: '+1', CallStatus: 'no-answer'}).then(() => {

                    expect(spyDelete).toHaveBeenCalledTimes(2);

                    expect(spyTwCreateCall).toHaveBeenCalledWith('sip:+201@kamailio.openline.ninja', '+1');
                    expect(spyTwCreateCall).toHaveBeenCalledWith('sip:+202@kamailio.openline.ninja', '+1');
                    expect(spyTwCreateCall).toHaveBeenCalledWith('sip:+203@kamailio.openline.ninja', '+1');
                });
            });
        });
    });
});


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