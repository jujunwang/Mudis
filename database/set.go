package database

import (
	HashSet "github.com/jujunwang/Mudis/datastruct/set"
	"github.com/jujunwang/Mudis/interface/database"
	"github.com/jujunwang/Mudis/interface/resp"
	"github.com/jujunwang/Mudis/lib/utils"
	"github.com/jujunwang/Mudis/resp/reply"
	"strconv"
)

func (db *DB) getAsSet(key string) (*HashSet.Set, reply.ErrorReply) {
	entity, exists := db.GetEntity(key)
	if !exists {
		return nil, nil
	}
	set, ok := entity.Data.(*HashSet.Set)
	if !ok {
		return nil, &reply.WrongTypeErrReply{}
	}
	return set, nil
}

func (db *DB) getOrInitSet(key string) (set *HashSet.Set, inited bool, errReply reply.ErrorReply) {
	set, errReply = db.getAsSet(key)
	if errReply != nil {
		return nil, false, errReply
	}
	inited = false
	if set == nil {
		set = HashSet.Make()
		db.PutEntity(key, &database.DataEntity{
			Data: set,
		})
		inited = true
	}
	return set, inited, nil
}

func execSAdd(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	members := args[1:]

	set, _, errReply := db.getOrInitSet(key)
	if errReply != nil {
		return errReply
	}
	counter := 0
	for _, member := range members {
		counter += set.Add(string(member))
	}
	db.addAof(utils.ToCmdLine3("sadd", args...))
	return reply.MakeIntReply(int64(counter))
}

func execSIsMember(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	member := string(args[1])

	set, errReply := db.getAsSet(key)
	if errReply != nil {
		return errReply
	}
	if set == nil {
		return reply.MakeIntReply(0)
	}

	has := set.Has(member)
	if has {
		return reply.MakeIntReply(1)
	}
	return reply.MakeIntReply(0)
}

func execSRem(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	members := args[1:]

	set, errReply := db.getAsSet(key)
	if errReply != nil {
		return errReply
	}
	if set == nil {
		return reply.MakeIntReply(0)
	}
	counter := 0
	for _, member := range members {
		counter += set.Remove(string(member))
	}
	if set.Len() == 0 {
		db.Remove(key)
	}
	if counter > 0 {
		db.addAof(utils.ToCmdLine3("srem", args...))
	}
	return reply.MakeIntReply(int64(counter))
}

func execSPop(db *DB, args [][]byte) resp.Reply {
	if len(args) != 1 && len(args) != 2 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'spop' command")
	}
	key := string(args[0])

	set, errReply := db.getAsSet(key)
	if errReply != nil {
		return errReply
	}
	if set == nil {
		return &reply.NullBulkReply{}
	}

	count := 1
	if len(args) == 2 {
		count64, err := strconv.ParseInt(string(args[1]), 10, 64)
		if err != nil || count64 <= 0 {
			return reply.MakeErrReply("ERR value is out of range, must be positive")
		}
		count = int(count64)
	}
	if count > set.Len() {
		count = set.Len()
	}

	members := set.RandomDistinctMembers(count)
	result := make([][]byte, len(members))
	for i, v := range members {
		set.Remove(v)
		result[i] = []byte(v)
	}

	if count > 0 {
		db.addAof(utils.ToCmdLine3("spop", args...))
	}
	return reply.MakeMultiBulkReply(result)
}

func execSCard(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])

	set, errReply := db.getAsSet(key)
	if errReply != nil {
		return errReply
	}
	if set == nil {
		return reply.MakeIntReply(0)
	}
	return reply.MakeIntReply(int64(set.Len()))
}

func execSMembers(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])

	set, errReply := db.getAsSet(key)
	if errReply != nil {
		return errReply
	}
	if set == nil {
		return &reply.EmptyMultiBulkReply{}
	}

	arr := make([][]byte, set.Len())
	i := 0
	set.ForEach(func(member string) bool {
		arr[i] = []byte(member)
		i++
		return true
	})
	return reply.MakeMultiBulkReply(arr)
}

func execSInter(db *DB, args [][]byte) resp.Reply {
	keys := make([]string, len(args))
	for i, arg := range args {
		keys[i] = string(arg)
	}

	var result *HashSet.Set
	for _, key := range keys {
		set, errReply := db.getAsSet(key)
		if errReply != nil {
			return errReply
		}
		if set == nil {
			return &reply.EmptyMultiBulkReply{}
		}

		if result == nil {
			// init
			result = HashSet.Make(set.ToSlice()...)
		} else {
			result = result.Intersect(set)
			if result.Len() == 0 {
				return &reply.EmptyMultiBulkReply{}
			}
		}
	}

	arr := make([][]byte, result.Len())
	i := 0
	result.ForEach(func(member string) bool {
		arr[i] = []byte(member)
		i++
		return true
	})
	return reply.MakeMultiBulkReply(arr)
}

func execSInterStore(db *DB, args [][]byte) resp.Reply {
	dest := string(args[0])
	keys := make([]string, len(args)-1)
	keyArgs := args[1:]
	for i, arg := range keyArgs {
		keys[i] = string(arg)
	}

	var result *HashSet.Set
	for _, key := range keys {
		set, errReply := db.getAsSet(key)
		if errReply != nil {
			return errReply
		}
		if set == nil {
			db.Remove(dest)
			return reply.MakeIntReply(0)
		}

		if result == nil {

			result = HashSet.Make(set.ToSlice()...)
		} else {
			result = result.Intersect(set)
			if result.Len() == 0 {
				db.Remove(dest)
				return reply.MakeIntReply(0)
			}
		}
	}

	set := HashSet.Make(result.ToSlice()...)
	db.PutEntity(dest, &database.DataEntity{
		Data: set,
	})
	db.addAof(utils.ToCmdLine3("sinterstore", args...))
	return reply.MakeIntReply(int64(set.Len()))
}

func execSUnion(db *DB, args [][]byte) resp.Reply {
	keys := make([]string, len(args))
	for i, arg := range args {
		keys[i] = string(arg)
	}

	var result *HashSet.Set
	for _, key := range keys {
		set, errReply := db.getAsSet(key)
		if errReply != nil {
			return errReply
		}
		if set == nil {
			continue
		}

		if result == nil {
			result = HashSet.Make(set.ToSlice()...)
		} else {
			result = result.Union(set)
		}
	}

	if result == nil {
		return &reply.EmptyMultiBulkReply{}
	}
	arr := make([][]byte, result.Len())
	i := 0
	result.ForEach(func(member string) bool {
		arr[i] = []byte(member)
		i++
		return true
	})
	return reply.MakeMultiBulkReply(arr)
}

func execSUnionStore(db *DB, args [][]byte) resp.Reply {
	dest := string(args[0])
	keys := make([]string, len(args)-1)
	keyArgs := args[1:]
	for i, arg := range keyArgs {
		keys[i] = string(arg)
	}

	var result *HashSet.Set
	for _, key := range keys {
		set, errReply := db.getAsSet(key)
		if errReply != nil {
			return errReply
		}
		if set == nil {
			continue
		}
		if result == nil {
			result = HashSet.Make(set.ToSlice()...)
		} else {
			result = result.Union(set)
		}
	}

	db.Remove(dest)
	if result == nil {
		return &reply.EmptyMultiBulkReply{}
	}

	set := HashSet.Make(result.ToSlice()...)
	db.PutEntity(dest, &database.DataEntity{
		Data: set,
	})

	db.addAof(utils.ToCmdLine3("sunionstore", args...))
	return reply.MakeIntReply(int64(set.Len()))
}

func execSDiff(db *DB, args [][]byte) resp.Reply {
	keys := make([]string, len(args))
	for i, arg := range args {
		keys[i] = string(arg)
	}

	var result *HashSet.Set
	for i, key := range keys {
		set, errReply := db.getAsSet(key)
		if errReply != nil {
			return errReply
		}
		if set == nil {
			if i == 0 {
				return &reply.EmptyMultiBulkReply{}
			}
			continue
		}
		if result == nil {
			result = HashSet.Make(set.ToSlice()...)
		} else {
			result = result.Diff(set)
			if result.Len() == 0 {
				return &reply.EmptyMultiBulkReply{}
			}
		}
	}

	if result == nil {
		return &reply.EmptyMultiBulkReply{}
	}
	arr := make([][]byte, result.Len())
	i := 0
	result.ForEach(func(member string) bool {
		arr[i] = []byte(member)
		i++
		return true
	})
	return reply.MakeMultiBulkReply(arr)
}

func execSDiffStore(db *DB, args [][]byte) resp.Reply {
	dest := string(args[0])
	keys := make([]string, len(args)-1)
	keyArgs := args[1:]
	for i, arg := range keyArgs {
		keys[i] = string(arg)
	}

	var result *HashSet.Set
	for i, key := range keys {
		set, errReply := db.getAsSet(key)
		if errReply != nil {
			return errReply
		}
		if set == nil {
			if i == 0 {
				db.Remove(dest)
				return reply.MakeIntReply(0)
			}
			continue
		}
		if result == nil {
			result = HashSet.Make(set.ToSlice()...)
		} else {
			result = result.Diff(set)
			if result.Len() == 0 {
				db.Remove(dest)
				return reply.MakeIntReply(0)
			}
		}
	}

	if result == nil {
		db.Remove(dest)
		return &reply.EmptyMultiBulkReply{}
	}
	set := HashSet.Make(result.ToSlice()...)
	db.PutEntity(dest, &database.DataEntity{
		Data: set,
	})

	db.addAof(utils.ToCmdLine3("sdiffstore", args...))
	return reply.MakeIntReply(int64(set.Len()))
}

func execSRandMember(db *DB, args [][]byte) resp.Reply {
	if len(args) != 1 && len(args) != 2 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'srandmember' command")
	}
	key := string(args[0])

	set, errReply := db.getAsSet(key)
	if errReply != nil {
		return errReply
	}
	if set == nil {
		return &reply.NullBulkReply{}
	}
	if len(args) == 1 {
		members := set.RandomMembers(1)
		return reply.MakeBulkReply([]byte(members[0]))
	}
	count64, err := strconv.ParseInt(string(args[1]), 10, 64)
	if err != nil {
		return reply.MakeErrReply("ERR value is not an integer or out of range")
	}
	count := int(count64)
	if count > 0 {
		members := set.RandomDistinctMembers(count)
		result := make([][]byte, len(members))
		for i, v := range members {
			result[i] = []byte(v)
		}
		return reply.MakeMultiBulkReply(result)
	} else if count < 0 {
		members := set.RandomMembers(-count)
		result := make([][]byte, len(members))
		for i, v := range members {
			result[i] = []byte(v)
		}
		return reply.MakeMultiBulkReply(result)
	}
	return &reply.EmptyMultiBulkReply{}
}

func init() {
	RegisterCommand("SAdd", execSAdd, -3)
	RegisterCommand("SIsMember", execSIsMember, 3)
	RegisterCommand("SRem", execSRem, -3)
	RegisterCommand("SPop", execSPop, -2)
	RegisterCommand("SCard", execSCard, 2)
	RegisterCommand("SMembers", execSMembers, 2)
	RegisterCommand("SInter", execSInter, -2)
	RegisterCommand("SInterStore", execSInterStore, -3)
	RegisterCommand("SUnion", execSUnion, -2)
	RegisterCommand("SUnionStore", execSUnionStore, -3)
	RegisterCommand("SDiff", execSDiff, -2)
	RegisterCommand("SDiffStore", execSDiffStore, -3)
	RegisterCommand("SRandMember", execSRandMember, -2)
}
