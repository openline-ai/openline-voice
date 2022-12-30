exports.fetchHuntGroupMappings = jest.fn().mockImplementation((tenant_name, digit) => {
    if (tenant_name === 'openline') {
        if (digit === '1') {
            return Promise.resolve({
                    rows: [
                        {hunt_group_id: '1', call_type: 'sip', e164: '+101'},
                        {hunt_group_id: '1', call_type: 'sip', e164: '+102'}
                    ]
                }
            )
        } else if (digit === '2') {
            return Promise.resolve({
                    rows: [
                        {hunt_group_id: '2', call_type: 'sip', e164: '+201'},
                        {hunt_group_id: '2', call_type: 'sip', e164: '+202'},
                        {hunt_group_id: '2', call_type: 'sip', e164: '+203'}
                    ]
                }
            )
        }
    }
})
