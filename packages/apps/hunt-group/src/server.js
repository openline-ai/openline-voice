const express = require('express')
require('dotenv').config()
const {initdb} = require('./common/initdb')

if (process.env.SERVER_PORT == null) {
    console.error('Missing env variable: SERVER_PORT');
    process.exit(1);
}

const app = express()
app.use(express.json());
app.use(
    express.urlencoded({
        extended: true,
    })
);

const {welcome_openline, events} = require('./openline_handler')

// POST: /welcome
app.post('/welcome_openline', async (req, res) => {
    res.type('text/xml');
    res.send(await welcome_openline(req.body.From, req.body.CallSid));
});

app.post('/welcome_gaspos', (req, res) => {
    res.type('text/xml');
    res.send(welcome());
});

// POST: /menu
app.post('/menu', async (req, res) => {
    res.type('text/xml');
    return res.send(await menu(req.body.Digits, req.body.From, req.body.CallSid));
});

app.post('/voicemail', (req, res) => {
    res.type('text/xml');
    return res.send(voicemail(req.body.Digits));
});

// POST: /dial
app.post('/dial', (req, res) => {
    const digit = req.body.Digits;
    res.type('text/xml');
    res.send(dial(digit));
});


app.post('/dial_queue', async (req, res) => {
    res.type('text/xml');
    return res.send(dial_queue(req.body));
});

app.post('/events', (req, res) => {
    res.type('text/xml');
    return res.send(events(req.body));
});

app.post('/recording_events', (req, res) => {
    res.type('text/xml');
    return res.send(recording_events(req.body));
});

initdb()
    .then(() => console.log('DB init successfully'))
    .catch(e => console.error(e.stack))

app.listen(process.env.SERVER_PORT, () => {
    console.log(`Hunt Group server listening on port ${process.env.SERVER_PORT}`)
})