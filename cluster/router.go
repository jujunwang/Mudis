package cluster

import "github.com/jujunwang/Mudis/interface/resp"

// CmdLine 代表一个命令
type CmdLine = [][]byte

func makeRouter() map[string]CmdFunc {
	routerMap := make(map[string]CmdFunc)
	routerMap["ping"] = ping

	routerMap["del"] = Del

	routerMap["exists"] = defaultFunc
	routerMap["type"] = defaultFunc
	routerMap["rename"] = Rename
	routerMap["renamenx"] = Rename

	routerMap["set"] = defaultFunc
	routerMap["setnx"] = defaultFunc
	routerMap["get"] = defaultFunc
	routerMap["getset"] = defaultFunc

	routerMap["flushdb"] = FlushDB

	return routerMap
}

// 将命令转发给负责的节点，并将其回复返回给客户端
func defaultFunc(cluster *ClusterDatabase, c resp.Connection, args [][]byte) resp.Reply {
	key := string(args[1])
	peer := cluster.peerPicker.PickNode(key)
	return cluster.relay(peer, c, args)
}
