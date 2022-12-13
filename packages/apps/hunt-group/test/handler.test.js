const {welcome, menu} = require('../src/handler');
jest.mock('../src/fetch');

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
        expect(twiml).toContain('Thanks for calling Openline. Please press 1 to dial an extension. Press 2 to talk to an Openline representative.');
    });
});

describe('Handler#Menu', () => {
    it('should redirect to welcomes with digits other than 1 or 2', () => {
        const twiml = menu();
        const count = countWord(twiml);

        // TwiML verbs
        expect(count('Say')).toBe(2);
        expect(count('Say')).toBe(2);

        // TwiML content
        expect(twiml).toContain('welcome');
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