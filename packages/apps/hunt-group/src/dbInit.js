const db = require('./db')
const fs = require('fs');

exports.dbInit = async () => {
    const sql = fs.readFileSync('sql/hunt-group.sql').toString();
    return await db.query(sql)
};