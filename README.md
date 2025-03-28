# My architecture skills with Golang
The purpose of this repository is to show my mastery of the following topics:
- Docker
- Graceful shutdown
- Horizontal scaling
- Hexagonal pattern
- Circuit breaker pattern
- Transactional outbox pattern
<!-- TODO: - Caching (LRU) -->

To demonstrate this I built an API to manage user-related operations.

### Use cases

1. Create a user and notify it to other systems.
2. Update a user and notify it to other systems.
3. Make a soft deletion of a user and notify other systems.
4. Obtain information of a user by ID.

### Components
![Component diagram](docs/diagrams/components.svg)

As the diagram shows, this project is compiled in two parts:
1. `users-http`  It is the binary in charge of manage the http requests (service/sender)
2. `users-relay` It is the binary in charge of read and send the messages (message relay)

### Quick start

[See the OpenAPI specification](docs/OpenAPI.json)

[See the required environment variables](.env.example)

[See the database schema described by the .sql files](scripts/sql)

###### How to run

```shell
export $(grep -v ^# .env.example)
make up
```

### Architecture decisions
###### Go project layout standard
I decided to follow the [Go project layout standard](https://github.com/golang-standards/project-layout).
###### Package tree
I built the package tree following the concepts of the [hexagonal architecture pattern](https://alistair.cockburn.us/hexagonal-architecture/).
```
.
├── cmd
├── internal
│   ├── app
│   │   ├── business (Use cases, business rules, data models and ports)
│   │   ├── input    (Everything related to "drive" adapters)
│   │   └── output   (Everything related to "driven" adapters)
│   └── container (DI container)
└── pkg (Public and global code, potencially libraries)
```
###### Compile only what is required
According to the theory of hexagonal architecture, it is possible to have *n* adapters for different external signals (http, gRPC, command line, kafka, serverless).

So I decided to compile a binary to handle each signal. [See why this decision]()