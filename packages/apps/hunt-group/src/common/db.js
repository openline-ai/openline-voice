const {Pool} = require('pg')

const pool = new Pool({
    max: process.env.SQL_POOL_MAX_CONN,
    idleTimeoutMillis: process.env.SQL_POOL_MAX_IDLE_CONN,
    connectionTimeoutMillis: process.env.SQL_POOL_CONNECTIONTIMEOUTMILLIS,
})

module.exports = {
    query: (text) => pool.query(text),
}