<img src="https://github.com/user-attachments/assets/46a5c546-7e9b-42c7-87f4-bc8defe674e0" width=250 />

DuckDB + LiteFS Example
==========================

This repository is an example of a toy application running on LiteFS. You can
test it out by deploying to Fly.io, or locally with a docker-compose setup.

* [Fly.io instructions](./fly-io-config)
* [Docker-compose instructions](./docker-config)

**Note: commands should be run from the top-level directory for both Fly.io and
local docker-compose (not from the location of the README files above).**


### DuckDB Version
This example is modified to use DuckDB instead of SQlite by leveraging the DuckDB sqlite extension and file format supporting read/write capabilities on the primary node and read-only on replicas.

#### Usage
Once the `/litefs` filesystem is mounted we can use it from DuckDB
```sql
--- Install the sqlite extension
INSTALL sqlite; LOAD sqlite;
--- Attach an sqlite file in the /litefs mount
ATTACH '/litefs/db.sqlite' (TYPE SQLITE); USE db;
--- Create a distributed table
CREATE TABLE IF NOT EXISTS db.persons (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    phone TEXT NOT NULL,
    company TEXT NOT NULL
);
--- Insert some data (or use the web demo)
WITH next_id AS (
    SELECT COALESCE(MAX(id), 0) + 1 AS id FROM db.persons
)
INSERT INTO db.persons (id, name, phone, company)
SELECT id, 'Jill', '1234', 'Jack'
FROM next_id;
```

> Data is automatically replicated to the read-only replica nodes. Replicas can become primary nodes.

#### Demo
![LiteFS-Example-ezgif com-optimize](https://github.com/user-attachments/assets/ae5ba93c-d784-4292-a1ab-11e9098e577e)
