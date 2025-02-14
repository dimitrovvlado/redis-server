## redis-server
A redis server/CLI implementation, compatible with the [RESP](https://redis.io/docs/latest/develop/reference/protocol-spec/) protocol.

Start the server by running `make run-server`

Run the CLI by executing `make run-cli`

### Available commands

**PING**
```
PING [message]
```

**ECHO**
```
ECHO message
```

**SET**
```
SET key value [EX seconds | PX milliseconds | EXAT unix-time-seconds | PXAT unix-time-milliseconds]
```

**GET**
```
GET key
```

**DEL**
```
DEL key [key ...]
```

**EXISTS**
```
EXISTS key [key ...]
```
