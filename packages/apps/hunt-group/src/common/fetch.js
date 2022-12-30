const db = require('./db')

exports.fetchHuntGroupMappings = (tenant_name, extension, priority) => {
    const query = {
        name: 'hungGroupMappings',
        text: 'SELECT m.hunt_group_id, m.call_type, m.e164, t.id as tenant_id ' +
            'FROM hunt_group_mappings m ' +
            'INNER JOIN hunt_groups g ON m.hunt_group_id = g.id ' +
            'INNER JOIN hunt_group_tenants t ON g.tenant_id = t.id ' +
            'WHERE t.name = $1 AND g.extension = $2 AND m.priority = $3',
        values: [tenant_name, extension, priority]
    }
    return db.query(query)
};

exports.getNextPriorityInGroup = (hunt_group_id, priority) => {
    const query = {
        name: 'nextPriorityInGroup',
        text: 'SELECT priority ' +
            'FROM hunt_group_mappings ' +
            'WHERE hunt_group_id = $1 AND priority > $2 ' +
            'ORDER BY priority ' +
            'LIMIT 1',
        values: [hunt_group_id, priority]
    }
    return db.query(query)
}