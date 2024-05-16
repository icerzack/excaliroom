<p align="center">
  <h1 align="center">Excaliroom</h1>
  <p align="center">A simple WebSocket server for collaborative drawing with Excalidraw.</p>
</p>

---

>[!WARNING]
> Current project major version is _**0.x.**_ It means that the project is still in development and may have breaking changes in the future.

This is an _**unofficial**_ implementation of the [Excalidraw](https://excalidraw.com/) collaboration server. 
It uses WebSockets to communicate between clients and broadcast the changes to all connected clients.

## Table of Contents

- [Features](#features)
- [Configuration](#configuration)
    - [JWT and Board URLs](#jwt-and-board-urls)
    - [Storage](#storage)
- [Installation](#installation)
  - [Configuration](#configuration)
  - [Docker](#docker)
  - [Docker Compose](#docker-compose)
  - [Build Go binary](#build-go-binary)
- [How to use](#how-to-use)
- [Contributing](#contributing)
- [License](#license)

## Features

- Real-time collaboration with multiple users
- Authentication and validation with JWT
- Configurable storage (currently only supports **in-memory** storage)

## Configuration

The server uses **_.yaml_** configuration file to set up.
You can find an example configuration file [here](./config-example.yaml).

```yaml
apps:
  log_level: "DEBUG"
  rest:
    port: 8080
    validation:
      jwt_header_name: "<YOUR_JWT_HEADER_NAME>"
      jwt_validation_url: "<YOUR_JWT_VALIDATION_URL>"
      board_validation_url: "<YOUR_BOARD_VALIDATION_URL>"

storage:
  users:
    type: "in-memory"
  rooms:
    type: "in-memory"
```

Currently, the `apps` section contains the following configurations:
- `log_level`: The log level of the server. It can be one of the following: `DEBUG`, `INFO` (More levels will be added in the future).
- `rest`: The REST API configuration.
    - `port`: The port of the REST API.
    - `validation`: The JWT validation configuration.
        - `jwt_header_name`: The name of the header, in which `Excaliroom` will set the JWT token from client.
        - `jwt_validation_url`: The URL to validate the JWT token, which will be used to authenticate the user.
        - `board_validation_url`: The URL to validate the access to the board with the JWT token.
- `storage`: The storage configuration.
    - `users`: The user storage configuration.
        - `type`: The type of the storage. Currently, only `in-memory` is supported.
    - `rooms`: The room storage configuration.
        - `type`: The type of the storage. Currently, only `in-memory` is supported.

### JWT and Board URLs

To authenticate the user and validate the access to the board, you need to provide the URLs in the configuration file.

The `Excaliroom` server requires 2 URLs:
- `jwt_validation_url`: The URL to validate the JWT token. The `Excaliroom` server will send a `GET` request to this URL with the JWT token in the header. The server should return `200 OK` if the token is valid and the following JSON response:
    ```json
    {
      "id": "<USER_ID>"
    }
    ```
    The `id` will be used to identify the user.


- `board_validation_url`: The URL to validate the access to the board with the JWT token. The `Excaliroom` server will send a `GET` request to this URL with the JWT token in the header. The server should return `200 OK`.

### Storage

Currently, the server only supports `in-memory` storage for users and rooms.

## Installation

### Docker

>[!WARNING]
> Currently, the Docker image from the Docker Hub is only available for **linux/amd64** platform.
> 
> If you use another platform (e.g., **linux/arm64**), you can provide `--platform` flag to the `docker pull` command.

1. You can pull the Docker image from the Docker Hub:

    ```bash
    docker pull icerzack/excaliroom:latest
    ```

    Then, you can run the Docker container with the following command:

    ```bash
    docker run -d -p 8080:8080 -v path/to/config.yaml:/config.yaml -e CONFIG_PATH="/config.yaml" icerzack/excaliroom:latest
    ```

2. You can build the Docker image by yourself with the following command:

    ```bash
    docker build -f build/Dockerfile -e CONFIG_PATH="path/to/config.yaml" -t excaliroom .
    ```
        
    Then, you can run the Docker container with the following command:
    
    ```bash
    docker run -d -p 8080:8080 -v path/to/config.yaml:/config.yaml -e CONFIG_PATH="/config.yaml" excaliroom
    ```
   
### Docker Compose

You can use Docker Compose to run the server with the following `docker-compose.yml` file:

```yaml
services:
  app:
    image: icerzack/excaliroom:latest
    environment:
      - CONFIG_PATH=config.yaml
    ports:
      - "8080:8080"
    volumes:
      - ./config.yaml:/config.yaml
```

Then, you can run the server with the following command:

```bash
docker-compose up -d
```

### Build Go binary

Set the environment variable `CONFIG_PATH` to the path of the configuration file and build the binary:

```bash
export CONFIG_PATH="path/to/config.yaml"
go build -o excaliroom main.go
```

Then, you can run the binary:

```bash
./excaliroom
```

## How to use

Check the [docs](./docs) directory of this repo for the guides on how to use and integrate the `Excaliroom` server with your JavaScript application.

## Contributing

If you have a suggestion that would make this better, please fork the repo and create a pull request. You can also simply open an issue with the tag "enhancement". Don't forget to give the project a star! Thanks again!

1. Fork the Project
2. Create your Feature Branch (git checkout -b feature/AmazingFeature)
3. Commit your Changes (git commit -m 'Add some AmazingFeature')
4. Push to the Branch (git push origin feature/AmazingFeature)
5. Open a Pull Request

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](./LICENSE) file for details.