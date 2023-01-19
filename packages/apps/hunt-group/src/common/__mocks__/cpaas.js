const uuid = require("uuid");

exports.createCall = jest.fn().mockImplementation((to, from) => {
    return Promise.resolve({
        sid: 'CA' + uuid.v4().replaceAll("-", ""),
        status: 'initiated',
        to: to,
        from: from
    });
});

exports.cancelCall = jest.fn().mockImplementation((callSid) => {
    return Promise.resolve({sid: callSid, status: 'canceled'});
})

exports.recordVoicemail = jest.fn().mockResolvedValue({});