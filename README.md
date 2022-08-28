# Mudis

---

项目介绍：Mudis是一个用go语言实现的类Redis缓存数据库，Mudis的命令请求处理器是并行的，并且在多线程的前提下，能够保证提供的各种操作是线程安全的，因为命令请求阶段是并行的，Mudis能够避免某些操作（比如BLPOP、BRPOP等）阻塞整个服务器。

---

## 为什么这样设计

### 理论基础一：

golang的goroutine非常轻量，初始只需要 2-4k 的栈空间，并且利用golang runtime调度器对于协程的出色调度能力，可以极大地减少系统级线程的上下文切换时间

> 虽然线程比较轻量，但是在调度时也有比较大的额外开销。每个线程会都占用 1M 以上的内存空间，在切换线程时不止会消耗较多的内存，恢复寄存器中的内容还需要向操作系统申请或者销毁资源，每一次线程上下文的切换都需要消耗 ~1us 左右的时间，但是 Go 调度器对 Goroutine 的上下文切换约为 ~0.2us，减少了 80% 的额外开销。 ---《go语言设计与实现》

借助golang的**netpoller网络编程模型**，用同步的代码逻辑可以实现内部异步的命令请求处理，netpoller内部基于Reactor模式（一个非阻塞事件轮训器+IO多路复用）实现goroutine的高效切换，降低系统级线程的阻塞时间

这也是本项目决定采取并行存储架构的理论基础之一

### 理论基础二：

redis虽然在6.0版本之后引入了多线程，但是多线程主要引入在socket读写、协议解析方面，数据处理依旧采用单线程，对于海量数据处理的场景下，redis并没有充分发挥CPU的多核性能。我认为redis不考虑在数据处理方面引入多线程的原因在于避免锁竞争，针对这个问题，Mudis采用了分段锁的方案，通过降低锁的粒度，来降低加锁对性能的影响。

## 与Redis的区别

redis虽然在文件事件处理器上应用了IO复用来提高文件事件的响应速率，但是IO复用仍需要在文件事件处理完成之后，才向文件事件分派器传送下一个事件

> 尽管多个文件事件可能会并发地出现， 但 I/O 多路复用程序总是会将所有产生事件的套接字都入队到一个队列里面， 然后通过这个队列， 以有序（sequentially）、同步（synchronously）、每次一个套接字的方式向文件事件分派器传送套接字： 当上一个套接字产生的事件被处理完毕之后（该套接字为事件所关联的事件处理器执行完毕）， I/O 多路复用程序才会继续向文件事件分派器传送下一个套接字                 ----《Redis设计与实现》

**Mudis的文件事件分派器不需要等待一个IO事件处理完成，换句话说IO事件的处理是并行的，而事件处理完成的快速响应则是依靠netpoller的IO多路复用机制来实现。**

**Mudis单机版架构**

![000](https://github.com/jujunwang/picture/blob/master/Mudis000.png?raw=true)

对于每一个客户端连接，Mudis均会分配一个goroutine处理连接请求，因此对于请求的处理是并行的，既然是并发的，那么就必须要考虑一个问题——**如何并发安全性**

## 如何保证并发安全性

方案一：分段锁

基于分段锁设计实现ConcurrentHashMap：我们将 key 分散到固定数量的 shard 中避免 rehash 操作。shard 是有锁保护的 map, 当 shard 进行 rehash 时会阻塞shard内的读写，但不会对其他 shard 造成影响。

方案二：sync.Map

sync.Map是golang官方在1.9版本引入的一个并发安全的map，适合读多写少的场景。因为在 m.dirty 刚被提升后会将 m.read 复制到新的 m.dirty 中，在数据量较大的情况下复制操作会阻塞所有协程，会造成严重的性能问题。

因此，Mudis最终采用分段锁这样的低粒度锁的方案，来解决并发问题，并尽可能地降低对读写性能的影响。

## 性能测试

环境:

Go version：1.18

System: macOS Monterey 12.3.1

CPU: 2 GHz 四核Intel Core i5

Memory: 16 GB 3733 MHz LPDDR4X

redis-benchmark 测试主要结果:

```
SET: 86714.87 requests per second
GET: 81735.77 requests per second
INCR: 98732.52 requests per second
LPUSH: 88894.45 requests per second
RPUSH: 95578.39 requests per second
LPOP: 87981.72 requests per second
RPOP: 89623.36 requests per second
```

