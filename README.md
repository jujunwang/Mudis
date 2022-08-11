# Mudis

---

项目介绍：Mudis(Multi-Reactor Redis)是一个用go语言实现的Redis服务端程序，区别于Redis的单线程存储引擎，Mudis的存储引擎是并行的，并且在多线程的前提下，能够保证提供的各种操作是线程安全的，因为充分利用了多核性能，理论上Mudis相较于Redis更适合需要处理大量数据的场景（如数据排序）

---

## 为什么这样设计

### 理论基础一：

golang的goroutine非常轻量，初始只需要 2-4k 的栈空间，并且利用golang runtime调度器对于协程的出色调度能力，可以极大地减少系统级线程的上下文切换时间，

> 虽然线程比较轻量，但是在调度时也有比较大的额外开销。每个线程会都占用 1M 以上的内存空间，在切换线程时不止会消耗较多的内存，恢复寄存器中的内容还需要向操作系统申请或者销毁资源，每一次线程上下文的切换都需要消耗 ~1us 左右的时间[1](https://draveness.me/golang/docs/part3-runtime/ch06-concurrency/golang-goroutine/#fn:1)，但是 Go 调度器对 Goroutine 的上下文切换约为 ~0.2us，减少了 80% 的额外开销

这也是本项目决定采取并行存储架构的理论基础之一

### 理论基础二：

redis虽然在6.0版本之后引入了多线程，但是多线程主要引入在socket读写、协议解析方面，数据处理依旧采用单线程，对于海量数据处理的场景下，redis并没有充分发挥CPU的多核性能。我认为redis不考虑在数据处理方面引入多线程的原因在于避免锁竞争，针对这个问题，Mudis采用了分段锁的方案，通过降低锁的粒度，来降低加锁对性能的影响。

### Mudis单机版架构

![000](https://github.com/jujunwang/picture/blob/master/Mudis000.png?raw=true)

对于每一个客户端连接，Mudis均会分配一个goroutine处理连接请求，因此对于请求的处理是并行的，既然是并发的，那么就必须要考虑一个问题——**如何并发安全性**

## 如何保证并发安全性

方案一：分段锁

基于分段锁设计实现ConcurrentHashMap：我们将 key 分散到固定数量的 shard 中避免 rehash 操作。shard 是有锁保护的 map, 当 shard 进行 rehash 时会阻塞shard内的读写，但不会对其他 shard 造成影响。

代码在[github.com/jujunwang/Mudis/datastruct/dict/sync_dict_concurrenthashmap](https://github.com/jujunwang/Mudis/blob/master/datastruct/dict/sync_dict_concurrenthashmap.go)

方案二：sync.Map

sync.Map是golang官方在1.9版本引入的一个并发安全的map，适合读多写少的场景。因为在 m.dirty 刚被提升后会将 m.read 复制到新的 m.dirty 中，在数据量较大的情况下复制操作会阻塞所有协程，会造成严重的性能问题。

因此，Mudis最终采用分段锁这样的低粒度锁的方案，来解决并发问题，并尽可能地降低对读写性能的影响。
