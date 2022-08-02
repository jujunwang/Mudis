package cluster

import (
	"github.com/jujunwang/Mudis/interface/resp"
	"github.com/jujunwang/Mudis/resp/reply"
)

// Del从集群中原子地移除给定的writekey, writekey可以分布在任何节点上
// 如果给定的writekey分布在不同的节点上，Del将使用try-commit-catch删除它们
func Del(cluster *ClusterDatabase, c resp.Connection, args [][]byte) resp.Reply {
	replies := cluster.broadcast(c, args)
	var errReply reply.ErrorReply
	var deleted int64 = 0
	for _, v := range replies {
		if reply.IsErrorReply(v) {
			errReply = v.(reply.ErrorReply)
			break
		}
		intReply, ok := v.(*reply.IntReply)
		if !ok {
			errReply = reply.MakeErrReply("error")
		}
		deleted += intReply.Code
	}

	if errReply == nil {
		return reply.MakeIntReply(deleted)
	}
	return reply.MakeErrReply("error occurs: " + errReply.Error())
}
