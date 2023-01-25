# Openline Voice Network
![Octavian Tails On The Phone](../images/otter_phone.jpeg)

## Architecture

The Voice Network consists of Kamailio and Asterisk

![Network Diagram](../images/Full%20Voice%20Network.png)

### Kamailio
* receives and sends WebRTC calls directly to the browser via SIP over WebSockets
* authenticates WebRTC users using the ephemeral auth module of Kamailio
* supports digest auth for making outbound pstn calls
* Identifies the carrier used by ingress calls


### Asterisk
* media anchors all calls
* transcodes between webrtc & classic media
* records calls (Comming Soon)


## Building
* Kamailio currently can build to AWS or Docker
* Asterisk currently builds to AWS or Docker
* look at the asterisk and kamailio sub directories for more building instructions
