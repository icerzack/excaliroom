# How to use Excaliroom

This directory contains guidance for using and integrating the `Excaliroom` server with your existing project.

## Table of contents

- [Getting started](#getting-started)
    - [Pre-requisites](#pre-requisites)
    - [How Excaliroom works](#how-excaliroom-works)
- [API reference](#api-reference)
- [Examples](#examples)
- [FAQ](#faq)

## Getting started

### Pre-requisites

First of all, it is implied that you have `Backend` and `Frontend` applications:
- `Backend`: It may contain some of your business logic, and it may be written in any programming language.
- `Frontend`: It is your JavaScript application that will communicate with the `Excaliroom` server.

The `Excaliroom` WebSocket server is only responsible for handling real-time communication between your clients on the same board.
It does not store any data (except for the current state of the board and connected clients) and does not have any business logic.

However, `Excaliroom` server requires a `JWT` token to authenticate and authorize users.
The `JWT` token should be:
- Validated by your `Backend` application.
- Provided by your `Frontend` application.

The `Excaliroom` server also requires a URL to validate user access to the board.

The neccessary URLs for `JWT` validation and board validation should be provided in the `Excaliroom` server configuration file.
See the [Configuration](../README.md#configuration) section for more information.

### How Excaliroom works

The `Excaliroom` server uses WebSockets to communicate between clients and broadcast the changes to all connected clients.
It is a real-time collaboration server that allows multiple users to draw on the same board.

Currently, the sharing mechanism is simple:
- When a first user connects to the `Excaliroom`, it sends the current state of the board to the WebSocket server. This happens only if the user is the first one to connect to the board. After receiving the board state, the `Excaliroom` creates a new room and stores the board state and the user in the room.
- With the next user connecting to the same board, the `Excaliroom` adds the user to the existing room and broadcasts the current room state to all connected users.
- By default, no one can modify the board state. `Excaliroom` can handle board updates only from the _**Leader**_ of the room. By default, after creating a new room, no one is the _**Leader**_ of the room. The _**Leader**_ is the user who can modify the board state. The _**Leader**_ can be dropped by the _**Leader**_ itself. If the _**Leader**_ leaves the room, the _**Leader**_ role is reset so anyone can become the _**Leader**_.
- When the _**Leader**_ sends a new board state to the `Excaliroom`, the server broadcasts the new board state to all connected users. In other words, the _**Leader**_ is the only user who can modify the board state, while all other users can only view the board state.
- When the last user leaves the room, the room is deleted from the `Excaliroom`.

The `Excaliroom` sends and receives messages in JSON format. The message format is described in the [API reference](#api-reference) section.

## API reference

Each JSON message contains `event` field that describes the type of the message. The `event` field can have the following values:
- `connect`: The message is sent by `Frontend` when the user requests to connect to the board.
- `userConnected`: The message is sent by `Excaliroom` to all connected users when a new user connects to the board.
- `userDisconnected`: The message is sent by `Excaliroom` to all connected users when a user disconnects from the board.
- `setLeader`: The message is sent by `Frontend` when the user requests to become the _**Leader**_ of the room and sent by `Excaliroom` to all connected users when the _**Leader**_ changes.
- `newData`: The message is sent by `Frontend` when the user sends new board data to the server and sent by `Excaliroom` to all connected users when the _**Leader**_ sends new board data.

The JSON message format is as follows:
1. `connect` event:
```json
{
    "event": "connect",
    "board_id": "<BOARD_ID>",
    "jwt": "<JWT_TOKEN>"
}
```
- `board_id`: The unique identifier of the board.
- `jwt`: The JWT token that is used to authenticate and authorize the user. The `Excaliroom` server will use `jwt_validation_url` to validate the JWT token on your `Backend` and `jwt_header_name` to set the JWT to the header. After validating the JWT token, the `Excaliroom` server will use `board_validation_url` to validate the access to the board. See the [Configuration](../README.md#jwt-and-board-urls) section for more information.

2. `userConnected` event:
```json
{
    "event": "userConnected",
    "user_ids": ["<USER_ID_1>", "<USER_ID_2>", ...],
    "leader_id": "<LEADER_ID>"
}
```
- `user_ids`: The list of user identifiers that are connected to the board.
- `leader_id`: The identifier of the _**Leader**_ of the room. If the _**Leader**_ is not set, the `leader_id` will be `0`.

3. `userDisconnected` event:
```json
{
    "event": "userDisconnected",
    "user_ids": ["<USER_ID_1>", "<USER_ID_2>", ...],
    "leader_id": "<LEADER_ID>"
}
```
- `user_ids`: The list of user identifiers that are connected to the board.
- `leader_id`: The identifier of the _**Leader**_ of the room. If the _**Leader**_ is not set, the `leader_id` will be `0`.

4. `setLeader` event (request):
```json
{
    "event": "setLeader",
    "board_id": "<BOARD_ID>",
    "jwt": "<JWT_TOKEN>"
}
```
- `board_id`: The unique identifier of the board.
- `jwt`: The JWT token that is used to authenticate and authorize the user. The `Excaliroom` server will use `jwt_validation_url` to validate the JWT token on your `Backend` and `jwt_header_name` to set the JWT to the header. After validating the JWT token, the `Excaliroom` server will use `board_validation_url` to validate the access to the board. See the [Configuration](../README.md#jwt-and-board-urls) section for more information.

5. `setLeader` event (response):
```json
{
    "event": "setLeader",
    "board_id": "<BOARD_ID>",
    "user_id": "<USER_ID>"
}
```
- `board_id`: The unique identifier of the board.
- `user_id`: The identifier of the _**Leader**_ of the room.

6. `newData` event (request):
```json
{
    "event": "newData",
    "board_id": "<BOARD_ID>",
    "jwt": "<JWT_TOKEN>",
    "data": {
        "elements": "EXCALIDRAW_ELEMENTS_JSON",
        "appState": "EXCALIDRAW_APP_STATE_JSON"
    }
}
```
- `board_id`: The unique identifier of the board.
- `jwt`: The JWT token that is used to authenticate and authorize the user. The `Excaliroom` server will use `jwt_validation_url` to validate the JWT token on your `Backend` and `jwt_header_name` to set the JWT to the header. After validating the JWT token, the `Excaliroom` server will use `board_validation_url` to validate the access to the board. See the [Configuration](../README.md#jwt-and-board-urls) section for more information.
- `data`: The board data that is sent by the _**Leader**_ of the room.
    - `elements`: The JSON string of the Excalidraw `elements`.
    - `appState`: The JSON string of the Excalidraw `appState`.

See the [Excalidraw Docs](https://docs.excalidraw.com/docs/@excalidraw/excalidraw/api/props/initialdata) documentation for more information.

7. `newData` event (response):
```json
{
    "event": "newData",
    "board_id": "<BOARD_ID>",
    "data": {
        "elements": "EXCALIDRAW_ELEMENTS_JSON",
        "appState": "EXCALIDRAW_APP_STATE_JSON"
    }
}
```
- `board_id`: The unique identifier of the board.
- `data`: The board data that is sent by the _**Leader**_ of the room.
    - `elements`: The JSON string of the Excalidraw `elements`.
    - `appState`: The JSON string of the Excalidraw `appState`.

See the [Excalidraw Docs](https://docs.excalidraw.com/docs/@excalidraw/excalidraw/api/props/initialdata) documentation for more information.

## Examples

_Later_

## FAQ

_Later_

