const {fetchHuntGroupMappings} = require("./fetch");
const {createCall} = require("./cpaas");
const {trackCall} = require("./callTracker");

/**
 * Returns a TwiML to interact with the client
 * @return {String}
 */
exports.huntGroupStart = async (tenant_name, extension, priority, from, parentSid) => {
    return fetchHuntGroupMappings(tenant_name, extension, priority)
        .then((res) => {
            if (res.rows.length > 0) {
                for (let row of res.rows) {
                    let toDial = row.call_type === 'sip' ? `sip:${row.e164}@${process.env.KAMAILIO_DOMAIN}` : row.e164;
                    createCall(tenant_name, toDial, from)
                        .then((call) => trackCall(row.hunt_group_id, call.sid, parentSid, priority, call.status))
                        .catch((err) => console.log('Error while creating call: ' + err));
                }
            } else {
                console.log('No hunt group mappings found: ' + tenant_name + '->' + extension)
            }
        });
}