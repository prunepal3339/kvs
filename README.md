## KVS - Key Value Server

An attempt to clone redis server from scratch in Go while keeping it minimal.

### HOW TO USE?
1. Use this [docs](https://redis.io/docs/latest/develop/tools/cli/) to install and learn more about redis cli.
2. Run the server using `go run main.go`.

**NOTE**: You need to stop redis server if it is already running as our program uses the same port as redis server.

---

## Commands Supported

| Command | Description | Usage | Example |
|---------|------------|-------|---------|
| `PING [message]` | Check server connectivity. Returns `"PONG"` if no argument is given, or echoes the argument. | `PING`<br>`PING "Hello"` | `PING` → `"PONG"`<br>`PING "Hello World"` → `"Hello World"` |
| `GET key` | Fetch the value for the given key. Returns `nil` if key does not exist. | `GET mykey` | `SET username "puru"`<br>`GET username` → `"puru"`<br>`GET age` → `nil` |
| `SET key value` | Set the value for the given key. Overwrites if key exists. | `SET mykey "some value"` | `SET username "puru"`<br>`GET username` → `"puru"` |
| `HSET hash field value` | Set a field in a hash (nested map). Creates hash if it does not exist. | `HSET users user1 "Alice"` | `HSET users user1 "Alice"`<br>`HGET users user1` → `"Alice"` |
| `HGET hash field` | Fetch a value from a hash. Returns `nil` if hash or field does not exist. | `HGET users user1` | `HGET users user2` → `"Bob"`<br>`HGET users user3` → `nil` |
| `HGETALL hash` | Fetch all fields and values from a hash as a flat array `[field1, value1, ...]`. Returns empty array if hash does not exist. | `HGETALL users` | `HGETALL users` → `["user1", "Alice", "user2", "Bob"]`<br>`HGETALL unknown` → `[]` |
---
