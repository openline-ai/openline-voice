const uuid = require("uuid");

exports.createCall = jest.fn().mockImplementation((to, from) => {
    return Promise.resolve({
        sid: 'CA' + uuid.v4().replaceAll("-", ""),
        status: 'queued',
        to: to,
        from: from
    });
});

exports.hangUpCall = jest.fn().mockImplementation((callSid) => {
    return Promise.resolve({callsid: callSid});
})

exports.recordVoicemail = jest.fn().mockResolvedValue({});