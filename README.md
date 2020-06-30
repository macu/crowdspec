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
```

## Access database from command line

```
$ psql crowdspec
```

## Access dev database through GCloud

```
gcloud sql connect crowdspec-dev --user=postgres --quiet
postgres=# \c crowdspec
```
