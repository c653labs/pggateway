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
WARN[2018-04-15T08:44:43-04:00] listening for connections: ":5433"
INFO[2018-04-15T08:44:44-04:00] new client session                            client="127.0.0.1:49531" database=app session_id=501600aa-0a36-4e39-a42b-db393937aa17 ssl=true target="127.0.0.1:5432" user=test
INFO[2018-04-15T08:44:44-04:00] server response                               client="127.0.0.1:49531" database=app message="map[Type:AuthenticationRequest Payload:map[Method:5 Salt:[121 28 29 30]]]" session_id=501600aa-0a36-4e39-a42b-db393937aa17 ssl=true target="127.0.0.1:5432" user=test
INFO[2018-04-15T08:44:44-04:00] client request                                client="127.0.0.1:49531" database=app message="map[Type:PasswordMessage Payload:map[Password:[109 100 53 55 48 54 51 98 99 49 57 49 53 99 100 50 57 99 50 56 49 49 49 56 49 52 50 98 102 50 50 56 54 54 102]]]" session_id=501600aa-0a36-4e39-a42b-db393937aa17 ssl=true target="127.0.0.1:5432" user=test
INFO[2018-04-15T08:44:44-04:00] server response                               client="127.0.0.1:49531" database=app message="map[Type:AuthenticationRequest Payload:map[Method:0 Salt:[]]]" session_id=501600aa-0a36-4e39-a42b-db393937aa17 ssl=true target="127.0.0.1:5432" user=test
INFO[2018-04-15T08:44:44-04:00] server response                               client="127.0.0.1:49531" database=app message="map[Type:ParameterStatus Payload:map[Value:psql Name:application_name]]" session_id=501600aa-0a36-4e39-a42b-db393937aa17 ssl=true target="127.0.0.1:5432" user=test
INFO[2018-04-15T08:44:44-04:00] server response                               client="127.0.0.1:49531" database=app message="map[Type:ParameterStatus Payload:map[Name:client_encoding Value:UTF8]]" session_id=501600aa-0a36-4e39-a42b-db393937aa17 ssl=true target="127.0.0.1:5432" user=test
INFO[2018-04-15T08:44:44-04:00] server response                               client="127.0.0.1:49531" database=app message="map[Type:ParameterStatus Payload:map[Name:DateStyle Value:ISO, MDY]]" session_id=501600aa-0a36-4e39-a42b-db393937aa17 ssl=true target="127.0.0.1:5432" user=test
INFO[2018-04-15T08:44:44-04:00] server response                               client="127.0.0.1:49531" database=app message="map[Type:ParameterStatus Payload:map[Value:on Name:integer_datetimes]]" session_id=501600aa-0a36-4e39-a42b-db393937aa17 ssl=true target="127.0.0.1:5432" user=test
INFO[2018-04-15T08:44:44-04:00] server response                               client="127.0.0.1:49531" database=app message="map[Type:ParameterStatus Payload:map[Name:IntervalStyle Value:postgres]]" session_id=501600aa-0a36-4e39-a42b-db393937aa17 ssl=true target="127.0.0.1:5432" user=test
INFO[2018-04-15T08:44:44-04:00] server response                               client="127.0.0.1:49531" database=app message="map[Type:ParameterStatus Payload:map[Name:is_superuser Value:off]]" session_id=501600aa-0a36-4e39-a42b-db393937aa17 ssl=true target="127.0.0.1:5432" user=test
INFO[2018-04-15T08:44:44-04:00] server response                               client="127.0.0.1:49531" database=app message="map[Type:ParameterStatus Payload:map[Name:server_encoding Value:UTF8]]" session_id=501600aa-0a36-4e39-a42b-db393937aa17 ssl=true target="127.0.0.1:5432" user=test
INFO[2018-04-15T08:44:44-04:00] server response                               client="127.0.0.1:49531" database=app message="map[Type:ParameterStatus Payload:map[Value:10.3 Name:server_version]]" session_id=501600aa-0a36-4e39-a42b-db393937aa17 ssl=true target="127.0.0.1:5432" user=test
INFO[2018-04-15T08:44:44-04:00] server response                               client="127.0.0.1:49531" database=app message="map[Type:ParameterStatus Payload:map[Name:session_authorization Value:test]]" session_id=501600aa-0a36-4e39-a42b-db393937aa17 ssl=true target="127.0.0.1:5432" user=test
INFO[2018-04-15T08:44:44-04:00] server response                               client="127.0.0.1:49531" database=app message="map[Payload:map[Name:standard_conforming_strings Value:on] Type:ParameterStatus]" session_id=501600aa-0a36-4e39-a42b-db393937aa17 ssl=true target="127.0.0.1:5432" user=test
INFO[2018-04-15T08:44:44-04:00] server response                               client="127.0.0.1:49531" database=app message="map[Type:ParameterStatus Payload:map[Value:US/Eastern Name:TimeZone]]" session_id=501600aa-0a36-4e39-a42b-db393937aa17 ssl=true target="127.0.0.1:5432" user=test
INFO[2018-04-15T08:44:44-04:00] server response                               client="127.0.0.1:49531" database=app message="map[Type:BackendKeyData Payload:map[PID:5511 Key:506073006]]" session_id=501600aa-0a36-4e39-a42b-db393937aa17 ssl=true target="127.0.0.1:5432" user=test
INFO[2018-04-15T08:44:44-04:00] server response                               client="127.0.0.1:49531" database=app message="map[Type:ReadyForQuery Payload:map[Status:Idle]]" session_id=501600aa-0a36-4e39-a42b-db393937aa17 ssl=true target="127.0.0.1:5432" user=test
INFO[2018-04-15T08:44:45-04:00] client request                                client="127.0.0.1:49531" database=app message="map[Type:SimpleQuery Payload:map[Query:select 1;]]" session_id=501600aa-0a36-4e39-a42b-db393937aa17 ssl=true target="127.0.0.1:5432" user=test
INFO[2018-04-15T08:44:45-04:00] server response                               client="127.0.0.1:49531" database=app message="map[Type:RowDescription Payload:map[Fields:[map[TableOID:0 ColumnIndex:0 TypeOID:23 ColumnLength:4 TypeModifier:-1 Format:Text ColumnName:[63 99 111 108 117 109 110 63]]]]]" session_id=501600aa-0a36-4e39-a42b-db393937aa17 ssl=true target="127.0.0.1:5432" user=test
INFO[2018-04-15T08:44:45-04:00] server response                               client="127.0.0.1:49531" database=app message="map[Type:DataRow Payload:map[Fields:[1]]]" session_id=501600aa-0a36-4e39-a42b-db393937aa17 ssl=true target="127.0.0.1:5432" user=test
INFO[2018-04-15T08:44:45-04:00] server response                               client="127.0.0.1:49531" database=app message="map[Type:CommandCompletion Payload:map[Tag:SELECT 1]]" session_id=501600aa-0a36-4e39-a42b-db393937aa17 ssl=true target="127.0.0.1:5432" user=test
INFO[2018-04-15T08:44:45-04:00] server response                               client="127.0.0.1:49531" database=app message="map[Type:ReadyForQuery Payload:map[Status:Idle]]" session_id=501600aa-0a36-4e39-a42b-db393937aa17 ssl=true target="127.0.0.1:5432" user=test
INFO[2018-04-15T08:44:46-04:00] client session end                            client="127.0.0.1:49531" database=app session_id=501600aa-0a36-4e39-a42b-db393937aa17 ssl=true target="127.0.0.1:5432" user=test
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
  - bind: '127.0.0.1:5433'
    # Pass all authentication along to the target server
    authentication:
      passthrough:
        # PostgreSQL server to forward requests to
        target:
          host: '127.0.0.1'
          port: 5432
          # Databases we will accept requests for,
          # empty list matching any database
          databases:
            - "test"
    # Log messages from this listener to stdout
    logging:
      file:
        level: 'info'
        out: '-'
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
  - bind: ':5433'
    authentication:
      passthrough:
```

#### VirtualUser

Example usage:

```yaml
  - bind: ':5433'
    authentication:
      virtualuser-authentication:
        - name: 'host1'
          users:
            host1user0: 'pass0' # plaintext password
            host1user1: 'md5eb93956e0e5654c7bacc18531f2b2982' # md5 hashed password
          target:
            host: '127.0.0.1'
            port: 5432
            user: 'test'
            password: 'test'
            databases:
              - "test"
              - "stats"
        - name: 'host2'
          users:
            host2user3: 'SCRAM-SHA-256$4096:GLBU5JUuBLn0t6gh8SurcA==$lvLKVbiO0LBbj7fU7sGVa61Hy/QjOnOyz9N+qsTaIEQ=:9rO0gSuecLXGw6ArRS6PfK49YCo3iYgGKtDAR36wK5E=' # scram secret for `pggateway` password
          target:
            host: '127.0.0.2'
            port: 2345
            user: 'test2'
            password: 'test2'
```

### Logging

#### CloudWatch logs

CloudWatch logs plugin will write log entries to a CloudWatch log group and stream.

Configuration options:

- `group` - Log group name to write to.
- `stream` - Log stream name to write to.
- `region` - AWS region of the log group to write to.
- `level` - Log level to emit: "info", "warn", "debug", "error", "fatal", default "warn"

The log stream will be created if it does not already exist, but the log group must already exist.

Example usage:

```yaml
listeners:
  ':5433':
    logging:
      # Write log entries to `my-log-group/my-log-stream` in the `us-east-1` region
      cloudwatchlogs:
        group: 'my-log-group'
        stream: 'my-log-stream'
        region: 'us-east-1'
        level: 'info'
```

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
