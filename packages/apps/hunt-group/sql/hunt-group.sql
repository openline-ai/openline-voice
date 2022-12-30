CREATE TABLE IF NOT EXISTS hunt_group_tenants
(
    id   SERIAL PRIMARY KEY NOT NULL,
    name VARCHAR(50)        NOT NULL
);

CREATE TABLE IF NOT EXISTS hunt_groups
(
    id        SERIAL PRIMARY KEY NOT NULL,
    tenant_id INTEGER            NOT NULL,
    name      VARCHAR(50)        NOT NULL,
    extension VARCHAR(50)        NOT NULL,
    UNIQUE (tenant_id, name, extension),
    CONSTRAINT fk_tenant_id
        FOREIGN KEY (tenant_id)
            REFERENCES hunt_group_tenants (id)
);


CREATE TABLE IF NOT EXISTS hunt_group_mappings
(
    id            SERIAL PRIMARY KEY NOT NULL,
    hunt_group_id INTEGER            NOT NULL,
    call_type     VARCHAR(50)        NOT NULL,
    e164          VARCHAR(50)        NOT NULL,
    priority      INTEGER            NOT NULL,
    CONSTRAINT fk_hunt_group_id
        FOREIGN KEY (hunt_group_id)
            REFERENCES hunt_groups (id)
);