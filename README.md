# Relay

Relay is a real-time messaging backend built with Go. It provides
authenticated REST APIs for users and conversations, persistent
messaging with PostgreSQL, and WebSocket-based real-time communication.
Redis Pub/Sub distributes live events between Relay server instances so
users connected to different instances can receive events.

### Authentication & Users

- User registration
- Login with email and password
- Password hashing with bcrypt
- JWT access tokens
- Refresh-token support
- Protected authenticated routes
- Structured validation errors
- Consistent HTTP error responses
- Duplicate-email handling

### Conversations

- Create and manage conversations
- Conversation participant management
- Participant authorization
- Protection against users accessing conversations they do not belong
  to

### Messaging

- Send messages
- Persist messages in PostgreSQL
- Retrieve messages for a conversation
- Conversation-level authorization
- Real-time message delivery through WebSockets
- Support for multiple active WebSocket connections per user
- Redis Pub/Sub event distribution across Relay instances
- Cross-instance delivery for message, read-receipt, typing, and presence events

### Real-Time Features

Relay uses WebSockets for live communication.

Supported events:

Event Description Persistent

---

`message` Delivers a new message in real time Yes
`typing` Indicates that a participant is typing No
`presence` Indicates online/offline status No
`read_receipt` Notifies participants that a message was read Yes

#### Presence

Relay tracks whether a user is currently connected.

If a user has multiple WebSocket connections, they remain online until
their last connection closes.

#### Typing Indicators

Typing events are delivered only to currently connected participants.
They are intentionally not persisted because typing is temporary state.

#### Read Receipts

Read receipts are persisted in PostgreSQL and also delivered in real
time.

This means an offline participant can miss the WebSocket event but will
still see the correct read state when messages are fetched later.

Example:

```json
{
  "type": "read_receipt",
  "payload": {
    "message_id": 24,
    "conversation_id": 1,
    "user_id": 2,
    "read_at": "2026-08-13T16:20:43.31277+03:00"
  }
}
```

Read receipts can be triggered through both:

```text
HTTP API
   ↓
MarkAsRead
   ↓
PostgreSQL
   ↓
WebSocket Hub
   ↓
Connected participants
```

and:

```text
WebSocket
   ↓
HandleReadReceipt
   ↓
MarkAsRead
   ↓
PostgreSQL
   ↓
WebSocket Hub
```

Both paths use the same message service logic.

---

## Architecture

Relay follows a layered architecture:

```text
                 ┌─────────────────────┐
                 │      HTTP / WS      │
                 │      Handlers       │
                 └──────────┬──────────┘
                            │
                            ▼
                 ┌─────────────────────┐
                 │      Services       │
                 │  Business Logic     │
                 └──────────┬──────────┘
                            │
                 ┌──────────┴──────────┐
                 ▼                     ▼
        ┌─────────────────┐    ┌─────────────────┐
        │  Repositories   │    │ WebSocket Hub   │
        │   PostgreSQL    │    │ Live Connections│
        └────────┬────────┘    └────────┬────────┘
                 │                      │
                 ▼                      ▼
        ┌─────────────────┐    ┌─────────────────┐
        │   PostgreSQL    │    │ Connected Users │
        └─────────────────┘    └─────────────────┘
```

### Core layers

#### Handlers

Responsible for:

- HTTP request parsing
- Authentication context
- URL parameters
- HTTP status codes
- Response formatting

#### Services

Responsible for:

- Business rules
- Authorization checks
- Validation
- Coordinating repositories
- Triggering real-time notifications where appropriate

#### Repositories

Responsible for:

- PostgreSQL queries
- Persistence
- Database-specific operations

#### Event publisher and Redis subscriber

The message service depends on an `EventPublisher` interface rather than
publishing directly to a particular WebSocket Hub.

- The Redis publisher JSON-encodes broadcast events and publishes them
  to the `relay:events` channel.
- Each Relay instance subscribes to that channel.
- A subscriber forwards received events to that instance's local Hub.
- The local Hub routes events to matching connected users.

This allows a message created on one Relay instance to reach a recipient
connected to another instance.

#### WebSocket Hub

Responsible for:

- Tracking active WebSocket clients local to an instance
- Supporting multiple connections per user
- Routing events to connected users
- Avoiding new connections for every event

---

## Distributed Event Architecture

Relay uses **Redis Pub/Sub** to distribute live events between multiple
Relay server instances. Each instance has its own in-memory WebSocket
Hub, so a Hub cannot directly reach clients connected to another
instance. Redis provides the event-broadcast path between those Hubs.

```text
                         RELAY DISTRIBUTED ARCHITECTURE

 Client A                                                         Client B
     │                                                                ▲
     ▼                                                                │
┌──────────────────────┐                              ┌──────────────────────┐
│   Relay Instance A   │                              │   Relay Instance B   │
│                      │                              │                      │
│  HTTP / WS Handlers  │                              │  HTTP / WS Handlers  │
│          │           │                              │          ▲           │
│          ▼           │                              │          │           │
│     MessageService ──┼── publish event ──┐           │   Redis Subscriber  │
│          │           │                  │           │          │           │
│          ▼           │                  ▼           │          ▼           │
│     PostgreSQL       │          ┌────────────────┐  │       Hub B          │
│   (persistent data)  │          │     REDIS      │  │          ▲           │
│                      │          │   Pub/Sub      │  │          │           │
│       Hub A          │          │ relay:events   │  │   WebSocket Clients  │
│          ▲           │          └───────┬────────┘  │                      │
│          │           │                  │           │                      │
│   Redis Subscriber ◄─┼──────────────────┘           │                      │
└──────────────────────┘                              └──────────────────────┘
```

### Event flow

1. A client sends a message to Relay Instance A.
2. `MessageService` saves the message through the repository to
   PostgreSQL.
3. The service publishes a `BroadcastEvent` through the `EventPublisher`
   interface. The Redis publisher serializes the event as JSON and
   publishes it to the `relay:events` channel.
4. Each connected Relay instance's Redis subscriber receives the event.
5. The subscriber forwards the event to its local WebSocket Hub.
6. Each Hub sends the event to matching recipient connections on that
   instance. For example, Bob can receive Alice's message even when Bob
   is connected to Instance B.

The event contains recipient IDs, allowing each Hub to route the event
to the appropriate local users. The Redis subscriber and publisher are
separate from the WebSocket Hub: Redis carries events between instances,
while each Hub delivers them to its own connected clients.

**Important limitation:** Redis Pub/Sub is transient; it is not a durable
message queue. A subscriber that is disconnected can miss live events.
The persisted message and read state remain available from PostgreSQL,
which is still Relay's source of truth.

---

## WebSocket Architecture

A WebSocket connection is established once and registered with the Hub.

```text
User 1 ───────────────┐
User 2 ───────────────┼──> WebSocket Hub
User 3 ───────────────┤
User 4 ───────────────┘
```

When an event needs to be delivered:

```go
s.hub.SendToUser(userID, payload)
```

does **not** create a new network connection.

It finds the user's existing WebSocket client(s) and queues the event on
their `Send` channel.

A user can therefore have multiple active connections:

```text
                 User 1
                   │
          ┌────────┼────────┐
          ▼        ▼        ▼
       Browser   Phone    Laptop
          │        │        │
          └────────┼────────┘
                   ▼
                  Hub
```

---

## Offline Users

WebSocket events are only delivered to currently connected clients.

Persistent information is stored in PostgreSQL.

For example:

```text
User 1 sends message
        │
        ▼
PostgreSQL
        │
        ├── User 2 online  → WebSocket delivery
        │
        └── User 3 offline → no WebSocket delivery
```

When User 3 reconnects, the client can retrieve the conversation history
from the database.

This keeps WebSockets as a **real-time delivery mechanism**, rather than
using them as the source of truth.

---

## Validation

Relay uses structured validation errors.

Example:

```json
{
  "errors": {
    "email": "invalid email address format"
  }
}
```

Validation errors are returned with HTTP `422 Unprocessable Entity`.

This allows clients to display field-specific validation messages.

---

## Error Handling

The API uses appropriate HTTP status codes for common failures.

Examples:

    Status Meaning

---

     `400` Invalid request
     `401` Authentication required or invalid credentials
     `403` Authenticated but not authorized
     `409` Resource conflict
     `422` Validation failure
     `500` Unexpected server error

---

## Logging

Relay includes request logging with request IDs.

Example:

```text
[bbea2ebf-fec7-4d8f-b837-782db26019f5] POST /users/login 200 229.203293ms
```

Request IDs make it easier to trace an individual request through logs.

---

## Technology Stack

- **Go**
- **Chi** HTTP router
- **PostgreSQL**
- **Redis** — Pub/Sub transport for distributing live events between
  Relay instances
- **Docker**
- **Gorilla WebSocket**
- **JWT**
- **bcrypt**
- SQL migrations

---

## Project Structure

The project is organized roughly as follows:

```text
relay/
├── cmd/
│   └── api/
│       └── main.go
│
├── internal/
│   ├── app/
│   ├── config/
│   ├── database/
│   ├── domain/
│   ├── handlers/
│   ├── middleware/
│   ├── models/
│   ├── repositories/
│   │   └── postgres/
│   ├── response/
│   ├── router/
│   ├── services/
│   ├── token/
│   ├── validation/
│   └── websocket/
│
├── migrations/
│
├── .env
├── go.mod
└── README.md
```

---

## Getting Started

### Prerequisites

Install:

- Go
- Docker
- PostgreSQL (or run PostgreSQL through Docker)
- Redis (or run Redis through Docker)

### Clone the project

```bash
git clone <repository-url>
cd relay
```

### Configure environment variables

Create a `.env` file at the project root.

Example configuration:

```env
APP_PORT=8080

DB_HOST=localhost
DB_PORT=5432
DB_USER=relay
DB_PASSWORD=your_password
DB_NAME=relay
DB_SSLMODE=disable

JWT_SECRET=your_secret
```

Use your project's actual configuration variable names if they differ.

### Start PostgreSQL and Redis

Both services are required by the current setup:

- **PostgreSQL** stores users, conversations, messages, and persistent
  read state.
- **Redis** distributes live broadcast events between Relay instances
  through the `relay:events` Pub/Sub channel.

Start the PostgreSQL and Redis containers before starting Relay. The
current Redis connection is configured in `cmd/api/main.go` as
`localhost:6379`; move this into environment-based configuration before
deploying to other environments.

### Run migrations

Run the project's migration command to create the required tables.

### Start the API

```bash
go run ./cmd/api
```

The server should start on:

```text
http://localhost:8080
```

---

## Testing

The project has unit tests for message-service behavior and HTTP handlers.
Redis cross-instance event paths have also been exercised for message,
read-receipt, typing, and presence events. For production confidence, add
automated integration tests that start multiple Relay instances against
the same PostgreSQL and Redis services.

Run the complete Go test suite:

```bash
go test ./...
```
---

## API Overview

The exact routes are defined by the router, but the application
currently provides functionality around:

### Users

```text
POST /users/register
POST /users/login
```

### Conversations

Conversation endpoints are protected and require an authenticated user.

### Messages

```text
POST /conversations/{conversationID}/messages
GET  /conversations/{conversationID}/messages
POST /conversations/{conversationID}/messages/{messageID}/read
```

The read endpoint returns:

```text
204 No Content
```

after successfully marking the message as read.

The operation also generates the real-time `read_receipt` event for
connected participants.

### WebSocket

The WebSocket endpoint requires authentication and upgrades the HTTP
connection into a persistent WebSocket connection.

Clients send events using the common structure:

```json
{
  "type": "event_type",
  "payload": {}
}
```

For example, a read receipt:

```json
{
  "type": "read_receipt",
  "payload": {
    "message_id": 24,
    "conversation_id": 1
  }
}
```

---

## WebSocket Event Reference

### Message

```json
{
  "type": "message",
  "payload": {
    "id": 24,
    "conversation_id": 1,
    "sender_id": 1,
    "content": "Hello",
    "created_at": "2026-08-13T15:39:42+03:00"
  }
}
```

### Typing

Client sends:

```json
{
  "type": "typing",
  "payload": {
    "conversation_id": 1
  }
}
```

Other participants receive:

```json
{
  "type": "typing",
  "payload": {
    "conversation_id": 1,
    "user_id": 1
  }
}
```

### Presence

```json
{
  "type": "presence",
  "payload": {
    "user_id": 1,
    "online": true
  }
}
```

### Read Receipt

Client sends:

```json
{
  "type": "read_receipt",
  "payload": {
    "message_id": 24,
    "conversation_id": 1
  }
}
```

Other participants receive:

```json
{
  "type": "read_receipt",
  "payload": {
    "message_id": 24,
    "conversation_id": 1,
    "user_id": 2,
    "read_at": "2026-08-13T16:20:43+03:00"
  }
}
```

---

## Security Considerations

Relay currently includes several security-related protections:

- Passwords are stored using bcrypt hashes rather than plaintext.
- JWT authentication protects private endpoints.
- Conversation membership is checked before accessing conversation
  data.
- Users cannot mark messages in conversations they do not belong to.
- Message IDs are checked against their conversation.
- Invalid credentials return a generic error rather than revealing
  whether an account exists.
- Validation is performed before authentication-related database
  operations where appropriate.

For production deployment, additional hardening should include:

- Strong production JWT secrets
- Secure WebSocket origin validation
- HTTPS/WSS
- Rate limiting
- Secure cookie/token handling where applicable
- Database credential management through secrets
- Production logging and monitoring

---

## Design Principles

### Database as source of truth

Messages and read receipts are persisted.

### WebSocket as real-time transport

WebSockets notify connected clients but are not relied upon for
permanent storage.

### Service layer owns business operations

Handlers should remain thin and delegate business logic to services.

### Repository layer owns persistence

Services should not contain database-specific SQL.

### Hub owns active WebSocket connections

The Hub routes events through already-established connections.

### Authorization before data access

Conversation membership is verified before users can access or modify
conversation resources.

---

## Current Status

Core real-time messaging functionality is implemented and working:

- [x] User registration
- [x] User login
- [x] JWT authentication
- [x] Refresh tokens
- [x] Structured validation errors
- [x] Conversations
- [x] Conversation participant authorization
- [x] Persistent messages
- [x] Message history
- [x] WebSocket connections
- [x] Real-time messages
- [x] Typing indicators
- [x] Online/offline presence
- [x] Persistent read receipts
- [x] Real-time read receipts
- [x] HTTP-triggered real-time read receipts
- [x] Multiple connections per user
- [x] Request ID logging
- [x] MessageService unit tests
- [x] HTTP handler tests
- [x] Redis event publisher and subscriber wiring
- [x] Redis Pub/Sub channel (`relay:events`)
- [x] Cross-instance delivery for message events
- [x] Cross-instance delivery for read receipts
- [x] Cross-instance delivery for typing indicators
- [x] Cross-instance delivery for presence events
- [x] Automated Go test execution (`go test ./...`)

---

## Future Improvements

Potential future work includes:

- Last-seen timestamps
- Unread message counts
- Message delivery states (`sent`, `delivered`, `read`)
- Add retry/reconnection handling for the Redis subscriber; currently a subscriber error can stop that subscriber
- Move the Redis address and other infrastructure settings into environment-based configuration
- Add stronger integration tests that exercise multiple Relay instances end to end
- Improve offline synchronization and document client resynchronization behavior
- Add/verify WebSocket heartbeat and ping-pong handling
- Add automatic client reconnection
- Add message pagination
- Add rate limiting
- Add production observability, metrics, and health checks for Redis
- Consider Redis Streams or an outbox pattern if durable event delivery becomes a requirement
- Add API documentation with OpenAPI/Swagger

---

## License

MIT

---
