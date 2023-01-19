const trackerMap = {}

exports.trackCall = (hunt_group_id, call_sid, parent_sid, priority, call_status) => {
    trackerMap[call_sid] = {hunt_group_id, parent_sid, priority, call_status}
};

/**
 * Returns all the call that were created by the hunt group
 * @param callSid
 * @returns {{}}
 */
exports.getActiveHuntGroupCalls = (callSid) => {
    let call = trackerMap[callSid];
    if (call) {
        const parentSid = call.parent_sid

        return Object.keys(trackerMap).reduce((map, key) => {
            let entry = trackerMap[key];
            if (entry.parent_sid === parentSid && (entry.call_status === 'initiated' || entry.call_status === 'ringing' || entry.call_status === 'in-progress')) {
                map[key] = trackerMap[key]
            }
            return map
        }, {})
    } else {
        return {}
    }
};

/**
 * Return the child call sids
 * @param parentSid
 * @returns {{}}
 */
exports.getChildCalls = (parentSid) => {
    return Object.keys(trackerMap).reduce((map, key) => {
        if (trackerMap[key].parent_sid === parentSid) {
            map[key] = trackerMap[key]
        }
        return map
    }, {})

};

/**
 * Removes the call from the tracker
 * @param callSid
 */
exports.removeCall = (callSid) => {
    if (trackerMap[callSid]) {
        delete trackerMap[callSid]
    }
};

/**
 * set status of the call
 * @param callSid
 */
exports.setCallStatus = (callSid, status) => {
    let trackerMapElement = trackerMap[callSid];
    if (trackerMapElement) {
        trackerMapElement.call_status = status
    }
};

/**
 * Returns the next priority in the hunt group
 * @returns {string[]}
 */
exports.getAllTrackedCallSids = () => {
    return Object.keys(trackerMap)
}

/**
 * Returns the next priority in the hunt group
 * @returns {string[]}
 */
exports.getTrackedCall = (callSid) => {
    return trackerMap[callSid];
}

/**
 *
 * @param callSid
 * @returns {boolean}
 */
exports.isTrackedCall = (callSid) => {
    return trackerMap[callSid] !== undefined
}