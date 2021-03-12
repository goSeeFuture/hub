# hub

chan集线器，聚合多个通道到单个协程处理。

以下这些方法参数，都被Group协程调用，它们之间可以访问相同的数据，串行执行，无需考虑数据竞争。

| 方法           | 参数       | 说明                                                     |
| -------------- | ---------- | -------------------------------------------------------- |
| GroupHandles() | processors | 设置Group（自定义）处理数据处理器队列                    |
| ListenEvent()  | handler    | 监听事件                                                 |
| ListenCall()   | handler    | 监听调用                                                 |
| SlowCall()     | callback   | 慢调用，使用额外协程执行fn函数，并将结果送回到callback中 |
| AfterFunc()    | fn         | 延时调用，类似`time.AfterFunc`，但fn在Group协程中执行    |
| Tick()         | fn         | 定时调用，间隔固定时间回调fn函数                         |

## 功能

### Group（组）

聚合通道数据到单协程处理。意图在于

- 将多协程编程变为单协程，降低编程难度
  - 附加 Attach
  - 处理器队列 Processors
- 提供常用的三种同步模型，降低同步编程难度
  - 慢调用 SlowCall
  - 事件 Event
  - 调用 Call
- 可配置的异常恢复
- 定时器函数

[Group使用说明](GROUP.md)
