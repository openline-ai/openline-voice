exports.insertCall = jest.fn().mockImplementation((callSid) => {
    return Promise.resolve({callsid: callSid});
});

exports.deleteCall = jest.fn().mockImplementation((callSid) => {
    return Promise.resolve({callsid: callSid});
})

exports.getCalls = jest.fn().mockImplementation((callSid) => {
    if (callSid === 'CA2') {
        return Promise.resolve({
                rows: [
                    {callsid: 'CA2'},
                    {callsid: 'CA3'},
                    {callsid: 'CA4'}
                ]
            }
        )
    } else if (callSid === 'CA3') {
        return Promise.resolve({
                rows: [
                    {callsid: 'CA3'},
                ]
            }
        )
    }
})
