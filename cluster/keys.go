package cluster

import (
	"github.com/jujunwang/Mudis/interface/resp"
	"github.com/jujunwang/Mudis/resp/reply"
)

// FlushDB 删除当前数据库的所有数据
func FlushDB(cluster *ClusterDatabase, c resp.Connection, args [][]byte) resp.Reply {
	replies := cluster.broadcast(c, args)
	var errReply reply.ErrorReply
	for _, v := range replies {
		if reply.IsErrorReply(v) {
			errReply = v.(reply.ErrorReply)
			break
		}
	}
	if errReply == nil {
		return &reply.OkReply{}
	}
	return reply.MakeErrReply("error occurs: " + errReply.Error())
}
