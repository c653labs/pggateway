# PGGateway
[![GoDoc](https://godoc.org/github.com/c653labs/pggateway?status.svg)](https://godoc.org/github.com/c653labs/pggateway)

**Note:** This project is still under active development and is not quite ready for prime time yet.

PGGateway is a PostgreSQL proxy service that allows you to use custom authentication and request logging plugins.

## Building
```bash
# clone
git clone git://github.com/c653labs/pggateway
cd ./pggateway

# via make
make

# build manually
CGO_ENABLED=0 go build -o pggateway -a -ldflags "-s -w" cmd/pggateway/main.go
```

## Running
```
$ pggateway -help
Usage of pggateway:
  -config string
        config file to load (default "pggateway.yaml")
```

```bash
# Run expecting default config file `pggateway.yaml`
pggateway

# Run explicitly setting config file
pggateway -config "server.yaml"
```

Once `pggateway` is running you can connect via your preferred PostgreSQL client, including `psql`.

```bash
psql "postgresql://user:password@127.0.0.1:5433/db_name""
```

<details>
  <summary>Example logging output from connecting and running `select 1;` and `\dt` via `psql`:</summary>
   <code><pre>
WARN[2018-04-02T20:31:49-04:00] listening for connections: ":5433"
INFO[2018-04-02T20:31:51-04:00] new client session                            client="127.0.0.1:58058" database=app session_id=ebb1959c-fc8e-4a1e-adb6-7be484a3980b ssl=true target="127.0.0.1:5432" user=test
INFO[2018-04-02T20:31:51-04:00] server response: AuthenticationRequest<Method=MD5, Salt=[121 29 119 240]>  client="127.0.0.1:58058" database=app session_id=ebb1959c-fc8e-4a1e-adb6-7be484a3980b ssl=true target="127.0.0.1:5432" user=test
INFO[2018-04-02T20:31:51-04:00] client request: PasswordMessage<Password="md5afbab2e5272d66a32c1182b45c4a4f95">  client="127.0.0.1:58058" database=app session_id=ebb1959c-fc8e-4a1e-adb6-7be484a3980b ssl=true target="127.0.0.1:5432" user=test
INFO[2018-04-02T20:31:51-04:00] server response: AuthenticationRequest<Method=OK, Salt=[]>  client="127.0.0.1:58058" database=app session_id=ebb1959c-fc8e-4a1e-adb6-7be484a3980b ssl=true target="127.0.0.1:5432" user=test
INFO[2018-04-02T20:31:51-04:00] server response: ParameterStatus<Name="application_name", Value="psql">  client="127.0.0.1:58058" database=app session_id=ebb1959c-fc8e-4a1e-adb6-7be484a3980b ssl=true target="127.0.0.1:5432" user=test
INFO[2018-04-02T20:31:51-04:00] server response: ParameterStatus<Name="client_encoding", Value="UTF8">  client="127.0.0.1:58058" database=app session_id=ebb1959c-fc8e-4a1e-adb6-7be484a3980b ssl=true target="127.0.0.1:5432" user=test
INFO[2018-04-02T20:31:51-04:00] server response: ParameterStatus<Name="DateStyle", Value="ISO, MDY">  client="127.0.0.1:58058" database=app session_id=ebb1959c-fc8e-4a1e-adb6-7be484a3980b ssl=true target="127.0.0.1:5432" user=test
INFO[2018-04-02T20:31:51-04:00] server response: ParameterStatus<Name="integer_datetimes", Value="on">  client="127.0.0.1:58058" database=app session_id=ebb1959c-fc8e-4a1e-adb6-7be484a3980b ssl=true target="127.0.0.1:5432" user=test
INFO[2018-04-02T20:31:51-04:00] server response: ParameterStatus<Name="IntervalStyle", Value="postgres">  client="127.0.0.1:58058" database=app session_id=ebb1959c-fc8e-4a1e-adb6-7be484a3980b ssl=true target="127.0.0.1:5432" user=test
INFO[2018-04-02T20:31:51-04:00] server response: ParameterStatus<Name="is_superuser", Value="off">  client="127.0.0.1:58058" database=app session_id=ebb1959c-fc8e-4a1e-adb6-7be484a3980b ssl=true target="127.0.0.1:5432" user=test
INFO[2018-04-02T20:31:51-04:00] server response: ParameterStatus<Name="server_encoding", Value="UTF8">  client="127.0.0.1:58058" database=app session_id=ebb1959c-fc8e-4a1e-adb6-7be484a3980b ssl=true target="127.0.0.1:5432" user=test
INFO[2018-04-02T20:31:51-04:00] server response: ParameterStatus<Name="server_version", Value="10.3">  client="127.0.0.1:58058" database=app session_id=ebb1959c-fc8e-4a1e-adb6-7be484a3980b ssl=true target="127.0.0.1:5432" user=test
INFO[2018-04-02T20:31:51-04:00] server response: ParameterStatus<Name="session_authorization", Value="test">  client="127.0.0.1:58058" database=app session_id=ebb1959c-fc8e-4a1e-adb6-7be484a3980b ssl=true target="127.0.0.1:5432" user=test
INFO[2018-04-02T20:31:51-04:00] server response: ParameterStatus<Name="standard_conforming_strings", Value="on">  client="127.0.0.1:58058" database=app session_id=ebb1959c-fc8e-4a1e-adb6-7be484a3980b ssl=true target="127.0.0.1:5432" user=test
INFO[2018-04-02T20:31:51-04:00] server response: ParameterStatus<Name="TimeZone", Value="US/Eastern">  client="127.0.0.1:58058" database=app session_id=ebb1959c-fc8e-4a1e-adb6-7be484a3980b ssl=true target="127.0.0.1:5432" user=test
INFO[2018-04-02T20:31:51-04:00] server response: BackendKeyData<PID=24985, Key=1461566709>  client="127.0.0.1:58058" database=app session_id=ebb1959c-fc8e-4a1e-adb6-7be484a3980b ssl=true target="127.0.0.1:5432" user=test
INFO[2018-04-02T20:31:51-04:00] server response: ReadyForQuery<Status=Idle>   client="127.0.0.1:58058" database=app session_id=ebb1959c-fc8e-4a1e-adb6-7be484a3980b ssl=true target="127.0.0.1:5432" user=test
INFO[2018-04-02T20:31:59-04:00] client request: SimpleQuery<Query="select 1;">  client="127.0.0.1:58058" database=app session_id=ebb1959c-fc8e-4a1e-adb6-7be484a3980b ssl=true target="127.0.0.1:5432" user=test
INFO[2018-04-02T20:31:59-04:00] server response: RowDescription<?column?<TableOID=0, ColumnIndex=0, TypeOID=23, ColumnLength=4, TypeModifier=-1, Format=Text>>  client="127.0.0.1:58058" database=app session_id=ebb1959c-fc8e-4a1e-adb6-7be484a3980b ssl=true target="127.0.0.1:5432" user=test
INFO[2018-04-02T20:31:59-04:00] server response: DataRow<"1">                 client="127.0.0.1:58058" database=app session_id=ebb1959c-fc8e-4a1e-adb6-7be484a3980b ssl=true target="127.0.0.1:5432" user=test
INFO[2018-04-02T20:31:59-04:00] server response: CommandCompletion<Tag="SELECT 1">  client="127.0.0.1:58058" database=app session_id=ebb1959c-fc8e-4a1e-adb6-7be484a3980b ssl=true target="127.0.0.1:5432" user=test
INFO[2018-04-02T20:31:59-04:00] server response: ReadyForQuery<Status=Idle>   client="127.0.0.1:58058" database=app session_id=ebb1959c-fc8e-4a1e-adb6-7be484a3980b ssl=true target="127.0.0.1:5432" user=test
INFO[2018-04-02T20:32:06-04:00] client request: SimpleQuery<Query="SELECT n.nspname as \"Schema\",\n  c.relname as \"Name\",\n  CASE c.relkind WHEN 'r' THEN 'table' WHEN 'v' THEN 'view' WHEN 'm' THEN 'materialized view' WHEN 'i' THEN 'index' WHEN 'S' THEN 'sequence' WHEN 's' THEN 'special' WHEN 'f' THEN 'foreign table' WHEN 'p' THEN 'table' END as \"Type\",\n  pg_catalog.pg_get_userbyid(c.relowner) as \"Owner\"\nFROM pg_catalog.pg_class c\n     LEFT JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace\nWHERE c.relkind IN ('r','p','')\n      AND n.nspname <> 'pg_catalog'\n      AND n.nspname <> 'information_schema'\n      AND n.nspname !~ '^pg_toast'\n  AND pg_catalog.pg_table_is_visible(c.oid)\nORDER BY 1,2;">  client="127.0.0.1:58058" database=app session_id=ebb1959c-fc8e-4a1e-adb6-7be484a3980b ssl=true target="127.0.0.1:5432" user=test
INFO[2018-04-02T20:32:06-04:00] server response: RowDescription<Schema<TableOID=2615, ColumnIndex=1, TypeOID=19, ColumnLength=64, TypeModifier=-1, Format=Text>, Name<TableOID=1259, ColumnIndex=1, TypeOID=19, ColumnLength=64, TypeModifier=-1, Format=Text>, Type<TableOID=0, ColumnIndex=0, TypeOID=25, ColumnLength=-1, TypeModifier=-1, Format=Text>, Owner<TableOID=0, ColumnIndex=0, TypeOID=19, ColumnLength=64, TypeModifier=-1, Format=Text>>  client="127.0.0.1:58058" database=app session_id=ebb1959c-fc8e-4a1e-adb6-7be484a3980b ssl=true target="127.0.0.1:5432" user=test
INFO[2018-04-02T20:32:06-04:00] server response: DataRow<"public", "pgbench_accounts", "table", "test">  client="127.0.0.1:58058" database=app session_id=ebb1959c-fc8e-4a1e-adb6-7be484a3980b ssl=true target="127.0.0.1:5432" user=test
INFO[2018-04-02T20:32:06-04:00] server response: DataRow<"public", "pgbench_branches", "table", "test">  client="127.0.0.1:58058" database=app session_id=ebb1959c-fc8e-4a1e-adb6-7be484a3980b ssl=true target="127.0.0.1:5432" user=test
INFO[2018-04-02T20:32:06-04:00] server response: DataRow<"public", "pgbench_history", "table", "test">  client="127.0.0.1:58058" database=app session_id=ebb1959c-fc8e-4a1e-adb6-7be484a3980b ssl=true target="127.0.0.1:5432" user=test
INFO[2018-04-02T20:32:06-04:00] server response: DataRow<"public", "pgbench_tellers", "table", "test">  client="127.0.0.1:58058" database=app session_id=ebb1959c-fc8e-4a1e-adb6-7be484a3980b ssl=true target="127.0.0.1:5432" user=test
INFO[2018-04-02T20:32:06-04:00] server response: DataRow<"public", "t", "table", "test">  client="127.0.0.1:58058" database=app session_id=ebb1959c-fc8e-4a1e-adb6-7be484a3980b ssl=true target="127.0.0.1:5432" user=test
INFO[2018-04-02T20:32:06-04:00] server response: DataRow<"public", "test", "table", "test">  client="127.0.0.1:58058" database=app session_id=ebb1959c-fc8e-4a1e-adb6-7be484a3980b ssl=true target="127.0.0.1:5432" user=test
INFO[2018-04-02T20:32:06-04:00] server response: CommandCompletion<Tag="SELECT 6">  client="127.0.0.1:58058" database=app session_id=ebb1959c-fc8e-4a1e-adb6-7be484a3980b ssl=true target="127.0.0.1:5432" user=test
INFO[2018-04-02T20:32:06-04:00] server response: ReadyForQuery<Status=Idle>   client="127.0.0.1:58058" database=app session_id=ebb1959c-fc8e-4a1e-adb6-7be484a3980b ssl=true target="127.0.0.1:5432" user=test
</pre></code>
</details>

## Configuration file
Basic example, proxying requests for any database from `127.0.0.1:5433` to `127.0.0.1:5432`.

```yaml
# Log server messages to stdout
logging:
  file:
    level: 'warn'
    out: '-'

listeners:
  # Listen for requests on port `5433`
  '127.0.0.1:5433':
    # PostgreSQL server to forward requests to
    target:
      host: '127.0.0.1'
      port: 5432

    # Pass all authentication along to the target server
    authentication:
      passthrough:

    # Log messages from this listener to stdout
    logging:
      file:
        level: 'info'
        out: '-'

    # Databases we will accept requests for,
    #   '*' is a special case matching any database
    databases:
      '*':
```

## Plugins
Authentication and logging plugins can be configured on a per-listener basis.

### Authentication
The following are the available built-in authentication plugins.

#### Passthrough
Passthrough authentication forwards all authentication requests to the target server.

There are no configuration options for `passthrough` authentication plugin.

Example usage:

```yaml
listeners:
  ':5433':
    authentication:
      passthrough:
```

### Logging
#### File
File logging writes log entries to a file or `stdout`.

Configuration options:

- `format` - Format of log entries: "text" or "json", default "text"
- `level` - Level of messages to emit: "info", "warn", "debug", "error", "fatal", default "warn"
- `out` - File to write log entries to: filename or "-" (stdout), default: "-"

Example usages:

```yaml
listeners:
  ':5433':
    logging:
      # Write log entries to /var/log/pggateway.log
      file:
        level: 'info'
        out: '/var/log/pggateway.log'
  ':5434':
    logging:
      # Write log entries formatted as JSON to stdout
      file:
        format: 'json'
        level: 'warn'
        out: '-'
```
