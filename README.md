# Mudis

---

项目介绍：这是一个用go语言重写的Redis服务端程序，区别于Redis的单线程存储引擎，Mudis的存储引擎是并行的，并且在多线程的前提下，能够保证提供的各种操作是线程安全的，且支持了可回滚事务的特性。

---

## 为什么这样设计

golang的goroutine非常轻量，初始只需要 2-4k 的栈空间，并且利用golang runtime调度器对于协程的出色调度能力，可以极大地减少系统级线程的上下文切换时间，

> 虽然线程比较轻量，但是在调度时也有比较大的额外开销。每个线程会都占用 1M 以上的内存空间，在切换线程时不止会消耗较多的内存，恢复寄存器中的内容还需要向操作系统申请或者销毁资源，每一次线程上下文的切换都需要消耗 ~1us 左右的时间[1](https://draveness.me/golang/docs/part3-runtime/ch06-concurrency/golang-goroutine/#fn:1)，但是 Go 调度器对 Goroutine 的上下文切换约为 ~0.2us，减少了 80% 的额外开销

这也是本项目决定采取并行存储架构的理论基础

### Mudis单机版架构

![mudis03](https://github.com/jujunwang/picture/blob/master/mudis03.png?raw=true)

对于每一个客户端连接，Mudis均会分配一个goroutine处理连接请求，因此对于请求的处理是并行的，既然是并发的，那么就必须要考虑一个问题——**如何并发安全性**

## 如何保证并发安全性

方案一：分段锁

基于分段锁设计实现ConcurrentHashMap：我们将 key 分散到固定数量的 shard 中避免 rehash 操作。shard 是有锁保护的 map, 当 shard 进行 rehash 时会阻塞shard内的读写，但不会对其他 shard 造成影响。

代码在[github.com/jujunwang/Mudis/datastruct/dict/sync_dict_concurrenthashmap](https://github.com/jujunwang/Mudis/datastruct/dict/sync_dict_concurrenthashmap)
