const {welcome, menu, events} = require('../src/common/handler');
jest.mock('../src/twilio_wrapper', () => {
    const {twilio_wrapper_mock} = require("../src/common/__mocks__/twilio_wrapper");
    return twilio_wrapper_mock();
});

jest.mock('../src/fetch', () => {
    const {mock_fetch} = require('../src/common/__mocks__/fetch')
    return mock_fetch()
});

jest.mock('../src/calls', () => {
    const {mock_calls} = require('../src/common/__mocks__/calls')
    return mock_calls()
});

describe('Handler#Menu', () => {
    it('should serve TwiML with enqueue and call first and second line support', () => {
        const twilio_wrapper = require('../src/common/twilio_wrapper');
        const calls = require('../src/common/calls');

        const spyTwCreateCall = jest.spyOn(twilio_wrapper, 'twCreateCall');
        const spyInsertCalls = jest.spyOn(calls, 'insertCall');

        // given that the user called us from +1 with callId = CA1 and pressed 1
        menu('1', '+1', 'CA1').then((twiml) => {
            const countFirst = countWord(twiml);
            expect(countFirst('Enqueue')).toBe(2);
            // we should have called the first line support
            expect(spyTwCreateCall).toHaveBeenCalledWith('sip:+101@kamailio.openline.ninja', '+1');
            expect(spyTwCreateCall).toHaveBeenCalledWith(  'sip:+102@kamailio.openline.ninja', '+1');
            // the calls were saved to our database
            expect(spyInsertCalls).toHaveBeenCalledTimes(2);
        });


        menu('2', '+1', 'CA2').then((twiml) => {
            const countSecond = countWord(twiml);
            expect(countSecond('Enqueue')).toBe(2);
            // we should have called the first line support
            expect(spyTwCreateCall).toHaveBeenCalledTimes(5);
            // the calls were saved to our database
            expect(spyInsertCalls).toHaveBeenCalledTimes(5);
        });
    });

    it('should serve cancel other calls after someone picked up the call', () => {
        const twilio_wrapper = require('../src/common/twilio_wrapper');
        const calls = require('../src/common/calls');
        const spyHangup = jest.spyOn(twilio_wrapper, 'twHangUpCall');
        const spyDelete = jest.spyOn(calls, 'deleteCall');

        // given that the user called from +1 with callId = CA1 and pressed 1
        menu('1', '+1', 'CA1').then((twiml) => {
            const countSecond = countWord(twiml);
            expect(countSecond('Enqueue')).toBe(2);
            // +101 has picked up the call. We should cancel the other calls
            events({To: '+101', CallSid: 'CA2', CallStatus: 'in-progress'}).then( () => {
                expect(spyHangup).toHaveBeenCalledTimes(2);
                expect(spyDelete).toHaveBeenCalledTimes(2);
            })
        });
    });
});

describe('Handler#Welcome', () => {
    it('should serve TwiML with gather', () => {
        const twiml = welcome();
        const count = countWord(twiml);

        // TwiML verbs
        expect(count('Gather')).toBe(2);
        expect(count('Say')).toBe(2);

        // TwiML options
        expect(twiml).toContain('action="/menu"');
        expect(twiml).toContain('numDigits="1"');
        expect(twiml).toContain('loop="3"');

        // TwiML content
        expect(twiml).toContain('Thanks for calling Openline. Please press 1 to talk to Sales. Press 2 to talk to Support. ');
    });
});

describe('Handler#Menu', () => {
    it('should redirect to welcomes with digits other than 1 or 2', () => {
        menu().then((twiml) => {
            const count = countWord(twiml);

            // TwiML verbs
            expect(count('Say')).toBe(2);
            expect(count('Say')).toBe(2);

            // TwiML content
            expect(twiml).toContain('welcome');
        });
    });

    it('should serve TwiML with gather', () => {
        const twiml = menu('1');
        const count = countWord(twiml);

        expect(count('Gather')).toBe(2);
        expect(count('Say')).toBe(2);

        // TwiML options
        expect(twiml).toContain('action="/dial"');
        expect(twiml).toContain('numDigits="3"');
        expect(twiml).toContain('method="POST"');

        // TwiML content
        expect(twiml).toContain('Please introduce the extension you would like to dial');

    });

    it('should serve TwiML with dial sip and number', () => {
        const twiml = menu('2');
        const count = countWord(twiml);

        // TwiML verbs
        expect(count('Dial')).toBe(6);
        expect(count('Sip')).toBe(4);
        expect(count('Number')).toBe(2);
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