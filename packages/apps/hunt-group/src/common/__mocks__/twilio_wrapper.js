const uuid = require("uuid");
exports.twCreateCall = jest.fn().mockImplementation((to, from) => {
    return Promise.resolve({
        sid: 'CA' + uuid.v4().replaceAll("-", ""),
        status: 'queued',
        to: to,
        from: from
    });
});

exports.twHangUpCall = jest.fn().mockImplementation((callSid) => {
    return Promise.resolve({callsid: callSid});
})