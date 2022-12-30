exports.fetchHuntGroupMappings = jest.fn().mockImplementation((tenant_name, extension, priority) => {
    if (tenant_name === 'openline') {
        if (extension === '900') {
            if (priority === '1') {
                return Promise.resolve({
                        rows: [
                            {hunt_group_id: '1', call_type: 'sip', e164: '+901'},
                            {hunt_group_id: '1', call_type: 'sip', e164: '+902'}
                        ]
                    }
                )
            } else if (priority === '2') {
                return Promise.resolve({
                        rows: [
                            {hunt_group_id: '2', call_type: 'sip', e164: '+903'},
                            {hunt_group_id: '2', call_type: 'sip', e164: '+904'}
                        ]
                    }
                )
            }
        }
    }
})

exports.getNextPriorityInGroup = jest.fn().mockImplementation((hunt_group_id, priority) => {
    if (hunt_group_id === '1') {
        if (priority === '1') {
            return Promise.resolve({rows: [{priority: '2'}]})
        } else if (priority === '2') {
            return Promise.resolve({rows: []})
        }
    } else if (hunt_group_id === '2') {
        if (priority === '1') {
            return Promise.resolve({rows: [{priority: '2'}]})
        } else if (priority === '2') {
            return Promise.resolve({rows: []})
        }
    }
})
