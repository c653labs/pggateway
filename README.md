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
