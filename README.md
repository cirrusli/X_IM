# X_IM

一个分布式即时通讯系统的框架。

![技术选型](./assets/technicalSelect.png)

## 项目启动

### docker部署相关依赖

MySQL、Redis、Consul

```bash
docker-compose -f "docker-compose.yml" up -d --build
```
### 更改相关server的配置文件

注意MySQL、Redis等的连接账户密码配置
### MySQL中建立数据库

新建x_base、x_message两个DB，字符集为utf8mb4，排序规则为utf8mb4_general_ci
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
├─test
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
