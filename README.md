# demo-hello

A tiny demo application written in Go that is shipped as a docker container.

## API 

| Verb | Path   | Description |
| ---- | ------ | ----------- |
| GET  | /health | Display a 'UP' message |
| GET  | /hello | Display a 'hello' message |
| GET  | /remote  | Call a remote service (get) using HELLO_REMOTE env variable |

# Docker

Docker images are published to: https://hub.docker.com/r/nocquidant/demo-hello/