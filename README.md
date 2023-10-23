# X_IM

一个分布式即时通讯系统的框架。

![技术选型](./assets/technicalSelect.png)

### 项目目录结构

#### assets

md依赖的图片文件等

#### container

容器层，服务托管

#### examples

各种mock、benchmark测试示例

#### gateway

网关层，负载均衡，管理长连接

#### logger

系统的日志模块，使用logrus库

#### naming

容器层，服务注册与发现

抽象通用的注册中心接口，可以更便捷地接入其他注册中心

使用consul，DNS做服务发现

#### tcp

通信层，处理tcp的数据

#### websocket

通信层，处理websocket的数据

#### wire

协议层，字节序、proto文件、自定义序列化等

##### channel.go

通信层，channel实现

##### channels.go

通信层

##### client.go

##### default_server.go

##### server.go

定义了通信层的接口，如channel