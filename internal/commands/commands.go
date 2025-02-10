package commands

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/dimitrovvlado/redis-server/internal/protocol"
)

func HandleCommand(resp protocol.Resp) (protocol.Resp, error) {
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
