const db = require('./db')

exports.insertCall = (hunt_group_id, call_sid, parent_sid) => {
    const query = {
        name: 'insert_calls',
        text: 'INSERT INTO hunt_group_calls(hunt_group_id, call_sid, parent_sid, calltime) VALUES ($1, $2, $3, now());',
        values: [hunt_group_id, call_sid, parent_sid],
    }
    return db.query(query)
};

exports.getCalls = (callSid) => {
    const query = {
        name: 'fetch_calls_with_same_parent',
        text: 'SELECT c1.call_sid, c2.hunt_group_id FROM hunt_group_calls c1 ' +
                'INNER JOIN hunt_group_calls c2 ON c2.call_sid = $1 AND c1.parent_sid = c2.parent_sid',
        values: [callSid],
    }
    return db.query(query)
};

exports.deleteCall = (callSid) => {
    const query = {
        name: 'delete_call',
        text: 'DELETE FROM hunt_group_calls WHERE call_sid = $1',
        values: [callSid],
    }
    return db.query(query)
};