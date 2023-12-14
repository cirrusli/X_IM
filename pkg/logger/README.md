# 用法

在启动服务器时，通过以下示例初始化logger，即可正常使用
```go
    _ = logger.Init(logger.Settings{
        Level:    config.Level,
        Filename: logPath,
    })
```
