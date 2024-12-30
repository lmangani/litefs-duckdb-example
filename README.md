LiteFS + DuckDB Example Application
==========================

This repository is an example of a toy application running on LiteFS. You can
test it out by deploying to Fly.io, or locally with a docker-compose setup.

* [Fly.io instructions](./fly-io-config)
* [Docker-compose instructions](./docker-config)

**Note: commands should be run from the top-level directory for both Fly.io and
local docker-compose (not from the location of the README files above).**


### DuckDB
This example is modified to use DuckDB instead of SQlite by leveraging the DuckDB sqlite extension and file format supporting read/write capabilities on the primary node and read-only on replicas.

![LiteFS-Example-ezgif com-optimize](https://github.com/user-attachments/assets/ae5ba93c-d784-4292-a1ab-11e9098e577e)
