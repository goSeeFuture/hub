# hub

chan集线器，聚合多个通道到单个协程处理。

## 主要功能

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

[Group使用说明](group.md)
