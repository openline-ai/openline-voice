const {Pool} = require('pg')

const pool = new Pool({
    user: process.env.SQL_USER,
    host: process.env.SQL_HOST,
    database: process.env.SQL_DATABASE,
    password: process.env.SQL_PASSWORD,
    port: process.env.SQL_PORT,
    max: process.env.SQL_POOL_MAX_CONN,
    idleTimeoutMillis: process.env.SQL_POOL_MAX_IDLE_CONN,
    connectionTimeoutMillis: process.env.SQL_POOL_CONNECTIONTIMEOUTMILLIS,
})
console.log(pool)
module.exports = {
    query: (text) => pool.query(text),
}