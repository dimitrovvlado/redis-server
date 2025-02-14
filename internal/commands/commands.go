package commands

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/dimitrovvlado/redis-server/internal/datastore"
	"github.com/dimitrovvlado/redis-server/internal/protocol"
)

func HandleCommand(resp protocol.Resp, ds *datastore.Datastore) (protocol.Resp, error) {
	switch resp.(type) {
	case protocol.Array:
		a := resp.(protocol.Array)
		cmd := (a.Items[0]).(protocol.BulkString)
		cmdS := strings.ToLower(protocol.Val(cmd.Data))
		args := (a.Items)[1:]
		switch cmdS {
		case "ping":
			return handlePingCommand(args), nil
		case "echo":
			return handleEchoCommand(args), nil
		case "set":
			return handleSetCommand(args, ds), nil
		case "get":
			return handleGetCommand(args, ds), nil
		case "del":
			return handleDelCommand(args, ds), nil
		case "exists":
			return handleExistsCommand(args, ds), nil
		default:
			return handleUnknownCommand(cmdS, args), nil
		}
	default:
		slog.Warn("Unexpected RESP type")
	}
	return nil, errors.New("unexpected RESP type")
}

func handlePingCommand(args []protocol.Resp) protocol.Resp {
	len := len(args)
	if len == 0 {
		return protocol.SimpleString{Data: "PONG"}
	} else if len == 1 {
		return protocol.BulkString{Data: protocol.Ptr(fmt.Sprintf("%s", args[0]))}
	} else {
		return protocol.Error{Data: "ERR wrong number of arguments for 'ping' command"}
	}
}

func handleEchoCommand(args []protocol.Resp) protocol.Resp {
	if len(args) != 1 {
		return protocol.Error{Data: "ERR wrong number of arguments for 'echo' command"}
	}
	return protocol.BulkString{Data: protocol.Ptr(fmt.Sprintf("%s", args[0]))}
}

func handleUnknownCommand(c string, args []protocol.Resp) protocol.Resp {
	argsArr := make([]string, len(args))
	for i, s := range args {
		argsArr[i] = fmt.Sprintf("'%s'", s)
	}
	return protocol.Error{Data: fmt.Sprintf("ERR unknown command '%s', with args beginning with: %s", c, strings.Join(argsArr, " "))}
}

func handleSetCommand(args []protocol.Resp, ds *datastore.Datastore) protocol.Resp {
	len := len(args)

	if len >= 2 {
		key := args[0].String()
		val := args[1].String()
		if len == 2 {
			ds.Set(key, val)
			return protocol.SimpleString{Data: "OK"}
		} else if len == 4 {
			expMode := strings.ToUpper(args[2].String())
			exp, err := strconv.ParseInt(args[3].String(), 10, 64)
			if err != nil || exp <= 0 {
				return protocol.Error{Data: "ERR value is not an integer or out of range"}
			}
			switch expMode {
			case "EX":
				//Set the specified expire time, in seconds
				ds.SetWithExpiry(key, val, exp*1000)
				return protocol.SimpleString{Data: "OK"}
			case "PX":
				//Set the specified expire time, in milliseconds
				ds.SetWithExpiry(key, val, exp)
				return protocol.SimpleString{Data: "OK"}
			case "EXAT":
				//Set the specified Unix time at which the key will expire, in seconds
				ds.SetWithExactExpiry(key, val, exp*1000)
				return protocol.SimpleString{Data: "OK"}
			case "PXAT":
				//Set the specified Unix time at which the key will expire, in milliseconds
				ds.SetWithExactExpiry(key, val, exp)
				return protocol.SimpleString{Data: "OK"}
			default:
				return protocol.Error{Data: "ERR syntax error"}
			}
		}
	}

	return protocol.Error{Data: "ERR wrong number of arguments for 'set' command"}
}

func handleGetCommand(args []protocol.Resp, ds *datastore.Datastore) protocol.Resp {
	if len(args) != 1 {
		return protocol.Error{Data: "ERR wrong number of arguments for 'get' command"}
	}
	key := args[0].String()
	val, err := ds.Get(key)
	if err != nil {
		return protocol.BulkString{Data: nil}
	}
	return protocol.BulkString{Data: protocol.Ptr(fmt.Sprintf("%s", val))}
}

func handleDelCommand(args []protocol.Resp, ds *datastore.Datastore) protocol.Resp {
	if len(args) < 1 {
		return protocol.Error{Data: "ERR wrong number of arguments for 'del' command"}
	}
	var cnt int64
	for _, k := range args {
		if err := ds.Delete(k.String()); err == nil {
			cnt += 1
		}
	}
	return protocol.Integer{Value: cnt}
}

func handleExistsCommand(args []protocol.Resp, ds *datastore.Datastore) protocol.Resp {
	if len(args) < 1 {
		return protocol.Error{Data: "ERR wrong number of arguments for 'exists' command"}
	}
	var cnt int64
	for _, k := range args {
		if _, err := ds.Get(k.String()); err == nil {
			cnt += 1
		}
	}
	return protocol.Integer{Value: cnt}
}
