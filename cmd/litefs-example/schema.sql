-- INSTALL sqlite; LOAD sqlite;
-- ATTACH '/litefs/db.sqlite' (TYPE SQLITE); USE litefs;
-- CREATE SEQUENCE id_sequence START 1;
CREATE TABLE IF NOT EXISTS db.persons (
--        id INTEGER DEFAULT nextval('duck.id_sequence') PRIMARY KEY,
        id INTEGER PRIMARY KEY,
        name TEXT NOT NULL,
        phone TEXT NOT NULL,
        company TEXT NOT NULL
);
