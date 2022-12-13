const db = require('./db')

exports.fetch = async () => {
    return await db.query('SELECT priority, call_type, e164 from openline_hunt_group_mapping')
};