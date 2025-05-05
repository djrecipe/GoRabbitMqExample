# RabbitMQ Example for Go
This is a very simple RPC server using RabbitMQ and written in GoLang. It has a very simple API of `set {key} {value}` and `get {key}`
## Setup
1. [Install Docker](https://docs.docker.com/engine/install/) for your system
2. Run `docker run -d --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3-management`
3. Optionally, you can now access management portal via `http://localhost:15672/` using user/password "guest/guest" from your system
4. [Install GoLang](https://go.dev/doc/install) for system

## Running the Server
1. Run `go run RPC rpc_server.go`

## Running the Client
1. In a new window, run `go run rpc_client.go "set" {key} {value}` e.g. `go run rpc_client.go "set" 0 7`
2. Observe RPC result of "success"
3. Run `go run rpc_client.go "get" {key}` e.g. `go run rpc_client.go "get" 0`
4. Observe RPC result of "7"

## Notes
- What is 'go.mod'?

go.mod is just the project file for the Go language.

## External Resources
- [Rabbit MQ examples](https://github.com/rabbitmq/rabbitmq-tutorials)
- [Rabbit MQ Docker Setup](https://www.svix.com/resources/guides/rabbitmq-docker-setup-guide/)
