CREATE TABLE IF NOT EXISTS openline_hunt_group (
    id SERIAL PRIMARY KEY NOT NULL,
    name VARCHAR(50) NOT NULL,
    priority NUMERIC NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS hunt_group_name_idx ON openline_hunt_group (name);
CREATE UNIQUE INDEX IF NOT EXISTS hunt_group_priority_idx ON openline_hunt_group (priority);

CREATE TABLE IF NOT EXISTS openline_hunt_group_mapping (
    id SERIAL PRIMARY KEY NOT NULL,
    priority NUMERIC NOT NULL,
    call_type VARCHAR(50) NOT NULL,
    e164 VARCHAR(50) NOT NULL,
    CONSTRAINT fk_priority
        FOREIGN KEY(priority)
            REFERENCES openline_hunt_group(priority)
);
