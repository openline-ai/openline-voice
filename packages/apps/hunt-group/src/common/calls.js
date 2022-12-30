const calls = {}

exports.saveCall = (hunt_group_id, call_sid, parent_sid, priority) => {
    calls[call_sid] = {hunt_group_id, parent_sid, priority}
};

exports.getCalls = (callSid) => {
    let call = calls[callSid];
    if (call) {
        const parentSid = call.parent_sid

        return Object.keys(calls).reduce((map, key) => {
            if (calls[key].parent_sid === parentSid) {
                map[key] = calls[key]
            }
            return map
        }, {})
    }
};

exports.deleteCall = (callSid) => {
    delete calls[callSid]
};

exports.getCallSids = () => {
    return Object.keys(calls)
}