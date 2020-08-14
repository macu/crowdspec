# CrowdSpec

© Matt Cudmore 2020

## Set up Postgres

```
$ psql postgres
postgres=# CREATE DATABASE crowdspec;
postgres=# CREATE ROLE dev WITH LOGIN ENCRYPTED PASSWORD 'devpw2020';
postgres=# GRANT ALL PRIVILEGES ON DATABASE crowdspec TO dev;
postgres=# ALTER DEFAULT PRIVILEGES FOR USER dev IN SCHEMA public GRANT ALL ON TABLES TO dev;
postgres=# ALTER DEFAULT PRIVILEGES FOR USER dev IN SCHEMA public GRANT ALL ON SEQUENCES TO dev;
postgres=# SELECT table_catalog, table_schema, table_name, privilege_type FROM information_schema.table_privileges WHERE grantee = 'dev';
```

### Access database from command line

```
$ psql crowdspec
postgres=# \dt
postgres=# \d spec
```

## Set up `env.json` for local development

```
{
	"dbUser": "testuser",
	"dbPass": "testpass",
	"dbName": "crowdspec",
	"httpPort": "2020",
	"versionStamp": ""
}
```

Version stamp is updated automatically when running `restart.sh`.

## Build and run

On first use, or to re-initialize database:
```
$ sh init.sh
```

To rebuild client and server, and run:
```
$ sh restart.sh
```

## Publish

```
$ gcloud app deploy
```

Update cron jobs:

```
$ gcloud app deploy cron.yaml
```

## Access dev database through GCloud

```
gcloud sql connect crowdspec-dev --user=postgres --quiet
postgres=# \c crowdspec
postgres=# \i <migration>.pgsql
```

## Vacuum

```
postgres=# SELECT pg_size_pretty(pg_database_size('postgres'));
postgres=# SELECT pg_size_pretty(pg_database_size('crowdspec'));
postgres=#
```

## TODO

- display dependencies' licenses on site
