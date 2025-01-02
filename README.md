## Usage

### Run via docker:

To run app and it's dependencies in docker, use next command:

```bash
task -d scripts docker_prod -v
```

### Run via source files:

To run application via source files, use next commands:

```shell
go run ./cmd/server/server.go
```

## gRPC:

To setup protobuf, use next command:

```shell
task -d scripts setup_proto -v
```

To generate files from .proto, use next command:

```shell
task -d scripts grpc_generate -v
```

## Linters

To run linters, use next command:

```shell
 task -d scripts linters -v
```

## Tests

To run test, use next commands. Coverage info will be
recorded to ```tests/coverage``` folder:

```shell
task -d scripts tests -v
```

## Benchmarks

To run benchmarks, use next command:

```shell
task -d scripts bench -v
```

## Migrations

To create migration file, use next command:

```shell
task -d scripts makemigrations NAME={{migration name}}
```

To apply all available migrations, use next command:

```shell
task -d scripts migrate
```

To migrate up to a specific version, use next command:

```shell
task -d scripts migrate_to VERSION={{migration version}}
```

To rollback migrations to a specific version, use next command:

```shell
task -d scripts downgrade_to VERSION={{migration version}}
```

To rollback all migrations (careful!), use next command:

```shell
task -d scripts downgrade_to_base
```

To print status of all migrations, use next command:

```shell
task -d scripts migrations_status
```
