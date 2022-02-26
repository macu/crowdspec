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
postgres=# \d+ spec
```

## Set up `env.json` for local development

```
{
	"dbUser": "testuser",
	"dbPass": "testpass",
	"dbName": "crowdspec",
	"httpPort": "2020",
	"adminUserId": 1,
	"recaptchaSiteKey": "...",
	"recaptchaSecretKey": "...",
	"mailjetApiKey": "...",
	"mailjetSecretKey": "...",
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

Make two separate deploys for database upgrades,
setting `MAINTENANCE_MODE` to `"true"` in app.yaml for the first deploy
to disable the site temporarily.

```
$ gcloud app deploy
```

Update cron jobs:

```
$ gcloud app deploy cron.yaml
```

## Vacuum

```
postgres=# SELECT pg_size_pretty(pg_database_size('postgres'));
postgres=# SELECT pg_size_pretty(pg_database_size('crowdspec'));
postgres=#
```

## Changelog

### 2022-02

- [x] Upgrade to Vue 3
- [x] Make public specs publicly accessible

### 2022-01

- [x] Refactoring
- [x] Jump to subspec modal update
- [x] Edit block modal ref form updates

### 2021-02

- [x] Add maintenance mode
- [x] Signup requests
- [x] Add unreadOnly community setting
- [x] Use YouTube API for fetching video previews
- [x] Add Markdown support
- [x] Add community review page
- [x] Initialize add block modal style type according to siblings
- [x] Add code highlighting

### 2021-01

- [x] Add community space, comments, sub-comments, read records, and unread counts
- [x] Scroll viewport as needed after move or insert block
- [x] Multiselect bulk block move (within or across contexts)

### 2020-12

- [x] Add user settings
- [x] Add forgot password form and email service
- [x] Freely move blocks between spec/subspec contexts
- [x] Change routes structure
- [x] Add loading messages when navigating specs
- [x] Case insensitive username login
- [x] Add username highlight
- [x] Allow moving blocks into empty contexts

### 2020-08-31

- [x] spec/subspec last modified time displayed to visitors now reflects block updates
- [x] change password form
- [x] reCAPTCHA on login

### 2020-08-30

- [x] toggle spec public
- [x] public specs on dashboard
- [x] remove editing features for visitors
- [x] show content unavailable when block refs have been deleted
- [x] record and display last modified times
- [x] change all timestamp columns to TIMESTAMPTZ
