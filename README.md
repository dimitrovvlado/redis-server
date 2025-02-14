## redis-server
A redis server/CLI implementation, compatible with the [RESP](https://redis.io/docs/latest/develop/reference/protocol-spec/) protocol.

### Available commands

*SET*
```
SET key value [EX seconds | PX milliseconds | EXAT unix-time-seconds | PXAT unix-time-milliseconds]
```

*GET*
```
GET key
```