exports.fetch = () => {
    return {
        rows: [
            {priority: '1', call_type: 'sip', e164: '+102'},
            {priority: '2', call_type: 'sip', e164: '+103'},
            {priority: '3', call_type: 'pstn', e164: '+40744305855'}
        ]
    }
};