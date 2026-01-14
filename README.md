Errormecha is an application for producing reliable errors and reporting them.

## How it works

This app builds a continuous stream of semi-random numbers that increase monotonically over time.
It writes the number to a configured PostgreSQL database, logging each successful write and all errors.
Every nine seconds, `errormecha` will interrupt and produce an error for that event instead of writing to the database, it looks like this in the logs:

```log
Inserted number: 233
2026/01/13 18:01:05 ERROR MECHA stole the write!
Inserted number: 239
Inserted number: 241
Inserted number: 250
Inserted number: 251
Inserted number: 259
Inserted number: 261
Inserted number: 267
Inserted number: 272
Inserted number: 272
2026/01/13 18:01:14 ERROR MECHA stole the write!
Inserted number: 273
```

## How to Use

### Kubernetes

> Requirements: Local Kubernetes cluster, local Docker engine

1. `docker build -t errormecha:latest .`
2. `kubectl apply -f postgres.yaml`
3. `kubectl apply -f mecha.yaml`

### Local Build

> Requirements: Local PostgreSQL instance

1. `go build -o errormecha` 
2. `./errormecha`

### Environment Variables

Five environment variables are available for authentication and database endpoint configurations.
Without any of these set, the app picks a set of dev environment defaults. They are:
- PG_DB = devdb
- PG_USER = devuser
- PG_PASS = devpass
- PG_HOST = postgres.local-db.svc.cluster.local
- PG_PORT = 5432

> These can be set in `mecha.yaml` for Kubernetes deployments

## Metrics

The app emits Prometheus metrics on port **8080** and is configured with a `ServiceMonitor` for collecting "Success" and "Error" counts.
When `errormecha` is operational, successful writes receive an "OK" count.
Any error, intentional or not, is counted as an "ERROR". 

This example views these metrics while `errormecha` is running as a local build:
```shell
$ curl -s localhost:8080/metrics | grep MECHA
# HELP MECHA_ERROR
# TYPE MECHA_ERROR counter
MECHA_ERROR 3
# HELP MECHA_WRITE_OK
# TYPE MECHA_WRITE_OK counter
MECHA_WRITE_OK 28
```

This app is also instrumented with standard Golang Runtime metrics, e.g.:
```shell
>>> curl localhost:8080/metrics
# HELP MECHA_ERROR
# TYPE MECHA_ERROR counter
MECHA_ERROR 6
# HELP MECHA_WRITE_OK
# TYPE MECHA_WRITE_OK counter
MECHA_WRITE_OK 56
# HELP go_cgo_go_to_c_calls_calls_total Count of calls made from Go to C by the current process. Sourced from /cgo/go-to-c-calls:calls.
# TYPE go_cgo_go_to_c_calls_calls_total counter
go_cgo_go_to_c_calls_calls_total 3
[...]
```