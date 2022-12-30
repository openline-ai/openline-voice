const db = require('./db')

exports.fetchHuntGroupMappings = (tenant_name, extension) => {
    const query = {
        name: 'hungGroupMappings',
        text: 'SELECT m.hunt_group_id, m.call_type, m.e164 ' +
            'FROM hunt_group_mappings m ' +
            'INNER JOIN hunt_groups g ON m.hunt_group_id = g.id ' +
            'INNER JOIN hunt_group_tenants t ON g.tenant_id = t.id ' +
            'WHERE t.name = $1 AND g.extension = $2',
        values: [tenant_name, extension]
    }
    return db.query(query)
};