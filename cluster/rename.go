package cluster

import (
	"github.com/jujunwang/Mudis/interface/resp"
	"github.com/jujunwang/Mudis/resp/reply"
)

// Rename 重命名一个key，但需要保证重命名前后通过一致性哈希映射在一个节点
func Rename(cluster *ClusterDatabase, c resp.Connection, args [][]byte) resp.Reply {
	if len(args) != 3 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'rename' command")
	}
	src := string(args[1])
	dest := string(args[2])

	srcPeer := cluster.peerPicker.PickNode(src)
	destPeer := cluster.peerPicker.PickNode(dest)

	if srcPeer != destPeer {
		return reply.MakeErrReply("ERR rename must within one slot in cluster mode")
	}
	//转发到目标节点
	return cluster.relay(srcPeer, c, args)
}
