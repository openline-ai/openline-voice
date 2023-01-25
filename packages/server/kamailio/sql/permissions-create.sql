CREATE TABLE IF NOT EXISTS kamailio_trusted (
    id SERIAL PRIMARY KEY NOT NULL,
    src_ip VARCHAR(50) NOT NULL,
    proto VARCHAR(4) NOT NULL,
    from_pattern VARCHAR(64) DEFAULT NULL,
    ruri_pattern VARCHAR(64) DEFAULT NULL,
    tag VARCHAR(64),
    priority INTEGER DEFAULT 0 NOT NULL
);

CREATE INDEX trusted_peer_idx ON kamailio_trusted (src_ip);

INSERT INTO version (table_name, table_version) values ('kamailio_trusted','6');

CREATE TABLE IF NOT EXISTS kamailio_address (
    id SERIAL PRIMARY KEY NOT NULL,
    grp INTEGER DEFAULT 1 NOT NULL,
    ip_addr VARCHAR(50) NOT NULL,
    mask INTEGER DEFAULT 32 NOT NULL,
    port SMALLINT DEFAULT 0 NOT NULL,
    tag VARCHAR(64)
);

CREATE INDEX kamailio_address_tag_idx ON kamailio_address (tag);
CREATE UNIQUE INDEX kamailio_address_ip_addr_idx ON kamailio_address (ip_addr);

INSERT INTO version (table_name, table_version) values ('kamailio_address','6');

INSERT INTO kamailio_address (grp, ip_addr, mask, port, tag) values
	(1, '54.172.60.0', 30, 0, 'twilio'),
	(1, '54.244.51.0', 30, 0, 'twilio'),
	(1, '54.171.127.192', 30, 0, 'twilio'),
	(1, '35.156.191.128', 30, 0, 'twilio'),
	(1, '54.65.63.192', 30, 0, 'twilio'),
	(1, '54.169.127.128', 30, 0, 'twilio'),
	(1, '54.252.254.64', 30, 0, 'twilio'),
	(1, '177.71.206.192', 30, 0, 'twilio');
