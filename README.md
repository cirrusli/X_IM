# X_IM

一个分布式即时通讯系统的框架。

![技术选型](./assets/technicalSelect.png)

## 项目启动

### docker部署相关依赖

MySQL、Redis、Consul

```bash
docker-compose -f "docker-compose.yml" up -d --build
```

### 服务启动

Gateway、Router、Logic、Occult
    
```bash
cd cmd
go run main.go [gateway/router/logic/occult]
```
### 项目目录结构

```bash
├─assets
│  └─data
├─cmd
│  └─data
├─docs
├─examples
│  ├─benchmark
│  ├─fuzz
│  ├─mock
│  └─ut
├─internal
│  ├─gateway
│  │  ├─conf
│  │  └─serv
│  ├─logic
│  │  ├─client
│  │  ├─conf
│  │  ├─handler
│  │  └─server
│  ├─occult
│  │  ├─conf
│  │  ├─database
│  │  └─handler
│  └─router
│      ├─api
│      ├─conf
│      ├─data
│      └─ip
├─pkg
│  ├─container
│  ├─ip
│  ├─kafka
│  ├─logger
│  ├─middleware
│  ├─naming
│  │  └─consul
│  ├─storage
│  ├─tcp
│  ├─timingwheel
│  │  └─delayqueue
│  ├─token
│  ├─websocket
│  ├─wire
│  │  ├─common
│  │  ├─endian
│  │  ├─pkt
│  │  ├─protofiles
│  │  └─rpc
│  └─x
└─scripts

```
