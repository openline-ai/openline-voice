const express = require('express')
const {dbInit} = require('./dbInit')

if (process.env.SERVER_PORT == null){
    console.error('Missing env variable: SERVER_PORT');
    process.exit(1);
}

dbInit().catch(ex => {
    console.log(process.env)
    console.error('Failed to initialize database');
    console.error(ex);
    process.exitCode = 1;
})

const app = express()
app.use(express.json());
app.use(
    express.urlencoded({
            extended: true,
    })
);

const {welcome, menu, dial} = require('./handler');
const {fetch} = require("./fetch");

// POST: /welcome
app.post('/welcome', (req, res) => {
    res.type('text/xml');
    res.send(welcome());
});

// POST: /menu
app.post('/menu', async (req, res) => {
    const digit = req.body.Digits;
    res.type('text/xml');
    return res.send(await menu(digit));
});

// POST: /dial
app.post('/dial', (req, res) => {
    const digit = req.body.Digits;
    res.type('text/xml');
    res.send(dial(digit));
});

app.listen(process.env.SERVER_PORT, () => {
    console.log(`Hunt Group server listening on port ${process.env.SERVER_PORT}`)
})