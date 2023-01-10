CREATE TABLE IF NOT EXISTS openline_carrier (
    id SERIAL PRIMARY KEY NOT NULL,
    carrier_name VARCHAR(50) NOT NULL,
    username VARCHAR(50) NOT NULL,
    realm VARCHAR(50) NOT NULL,
    ha1 VARCHAR(50) NOT NULL,
    domain VARCHAR(250) NOT NULL
);

CREATE UNIQUE INDEX carrier_name_idx ON openline_carrier (carrier_name);

CREATE TABLE IF NOT EXISTS openline_number_mapping (
    id SERIAL PRIMARY KEY NOT NULL,
    e164 VARCHAR(50) NOT NULL,
    sipuri VARCHAR(250) NOT NULL,
    carrier_name VARCHAR(50) NOT NULL,
    alias VARCHAR(50) NOT NULL
);

CREATE UNIQUE INDEX number_e164_idx ON openline_number_mapping (e164);
CREATE UNIQUE INDEX number_sipuri_idx ON openline_number_mapping (sipuri);
