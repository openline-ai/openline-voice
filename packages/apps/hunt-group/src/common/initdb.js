const db = require("./db");
const fs = require('fs');

const sql = fs.readFileSync('sql/hunt-group.sql').toString();

exports.initdb = () => {
    return db.query(sql)
};