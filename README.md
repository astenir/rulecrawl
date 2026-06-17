# RuleCrawl

RuleCrawl 是一个基于规则的 Go 爬虫框架原型。它把爬虫任务拆成“任务配置、请求调度、页面抓取、规则解析、结果存储”几个相对独立的环节，适合用来验证爬虫框架设计、编写站点采集规则，或作为后续服务化改造的基础。

当前项目仍处于原型阶段，已经具备可运行的核心链路，但还没有完成生产级生命周期管理、任务 API 和部署封装。

## 功能特性

- 基于 `spider.Task` 和 `spider.Rule` 的任务/规则模型
- 支持根请求生成、深度控制、去重和失败重试
- 并发 worker 调度，支持普通队列和优先请求队列
- 支持任务级限流，可组合多个限流器
- HTTP 抓取支持超时、Cookie、代理、随机 User-Agent 和字符集转换
- 解析结果通过 `spider.DataCell` 输出
- MySQL 存储支持自动建表和批量插入
- 内存存储 `storage/memstorage` 可用于测试、本地 demo 和 dry-run
- 规则级响应校验：可在 `spider.Rule.Validate` 中定义站点/页面特定的校验逻辑
- 示例任务注册与核心 engine 解耦，当前 Douban 示例位于 `examples/douban`
- 包含一个 go-micro/gRPC gateway 演示服务

## 目录结构

```text
.
├── collect/                 # HTTP fetcher
├── engine/                  # 调度器、worker、任务注册表
├── examples/douban/         # 当前 Douban 示例任务注册入口
├── limiter/                 # 组合限流器
├── log/                     # zap 日志封装
├── parse/                   # 示例解析规则
├── proto/greeter/           # go-micro/gRPC gateway demo
├── spider/                  # 任务、请求、规则、结果模型
├── sqldb/                   # MySQL SQL 生成和执行
├── storage/memstorage/      # 内存存储
├── storage/sqlstorage/      # MySQL 存储
├── config.toml              # 示例配置
└── main.go                  # 应用启动入口
```

## 环境要求

- Go `1.23.3`，以 `go.mod` 为准
- MySQL：使用 `storage/sqlstorage` 时需要
- etcd：运行 `main.go` 中的 go-micro 服务注册时需要

如果只是运行单元测试，通常不需要启动 MySQL 或 etcd。

## 快速开始

先按本地环境修改 `config.toml`：

- `storage.sqlURL` 改成本地 MySQL DSN
- `fetcher.proxy` 可留空，表示不使用代理
- `Tasks` 中的 `Name` 必须匹配已注册任务名，例如 `douban_book_list`
- `GRPCServer.RegistryAddress` 需要指向可用的 etcd 地址

运行测试和构建：

```bash
go test ./...
go build ./...
```

使用默认配置文件启动：

```bash
go run .
```

显式指定配置文件：

```bash
go run . -config ./config.toml
```

也可以通过环境变量指定：

```bash
CRAWLER_CONFIG=./config.toml go run .
```

## 配置说明

应用默认读取当前目录下的 `config.toml`。配置路径优先级如下：

1. 命令行参数：`-config /path/to/config.toml`
2. 环境变量：`CRAWLER_CONFIG=/path/to/config.toml`
3. 默认路径：`./config.toml`

主要配置项：

- `logLevel`：日志级别，例如 `debug`、`info`
- `Tasks`：启动任务列表，`Name` 必须匹配已注册任务
- `Tasks[].WaitTime`：请求前随机等待的最大秒数
- `Tasks[].Reload`：是否允许重复抓取同一请求
- `Tasks[].MaxDepth`：请求最大深度
- `Tasks[].Fetcher`：当前示例使用 `browser`
- `Tasks[].Limits`：任务限流配置
- `Tasks[].Cookie`：任务请求 Cookie
- `fetcher.timeout`：HTTP 请求超时时间，单位毫秒
- `fetcher.proxy`：代理列表，留空则直连
- `storage.sqlURL`：MySQL DSN
- `GRPCServer`：go-micro/gRPC gateway 监听和注册配置

## 示例任务

核心 `engine` 包不会自动注册具体站点任务。应用启动入口当前通过 `examples/douban` 注册 Douban 示例：

```go
doubanExamples.Register(engine.Store)
```

当前示例包括：

- `douban_book_list`
- Douban 小组示例
- Douban 小组 JavaScript 规则示例

新增 Go 任务时，通常需要定义：

- `Options.Name`：任务名，配置文件通过它引用任务
- `Rule.Root`：生成根请求
- `Rule.Trunk`：规则名称到解析规则的映射
- `Rule.ItemFields`：需要落库存储的字段列表
- `Rule.Validate`：可选，响应内容校验
- `Rule.ParseFunc`：解析页面并返回新请求或数据项

然后注册到任务表：

```go
engine.Store.Add(myTask)
```

## 存储

### MySQL 存储

`storage/sqlstorage` 会根据 `spider.DataCell` 中的任务名和规则字段自动建表，并按批次写入数据。

当前 SQL 层会校验并转义表名/列名，允许 Unicode 字母、数字和下划线，因此中文字段名可以正常使用；空格、点号、反引号等危险字符会被拒绝。

### 内存存储

`storage/memstorage` 实现了 `spider.Storage` 接口，适合用于测试和本地演示：

```go
storage := memstorage.New()
```

它提供：

- `Save`
- `All`
- `ByTask`
- `Len`
- `Reset`

## 测试

项目已有针对关键路径的基础测试：

- HTTP fetcher 非 200 响应处理
- `spider.Request` 默认行为和缺失 fetcher 错误
- SQL 建表/插入语句生成和标识符校验
- MySQL storage 批量 flush 行为
- 内存存储读写
- JS 任务模型属性传播
- engine 单请求处理链路：`fetch -> parse -> DataCell -> storage`
- engine 失败路径：fetch/validate/parse 失败

运行：

```bash
go test ./...
```

如果在容器环境中构建并遇到 VCS stamping 问题，可以临时关闭：

```bash
go build -buildvcs=false ./...
```

## 当前限制

这个仓库仍是原型，不建议直接当作生产爬虫平台使用。继续完善前需要重点关注：

- engine 还没有完整的 `context.Context` 生命周期管理和优雅退出
- SQLStorage 还缺少统一的退出前最终 `Flush`
- go-micro/gRPC 部分目前仍是 `Greeter` demo，不是正式 crawler 控制 API
- 任务状态、队列长度、成功/失败计数等运行时指标还不完整
- 示例任务仍依赖真实站点页面结构，可能随站点变化而失效
- 配置文件中包含本地示例 DSN 和代理地址，实际使用前必须调整

## 后续方向

建议后续按以下顺序推进：

1. 为 engine 增加 `context.Context` 生命周期、停止信号和优雅退出
2. 为存储层补 `Flush`/`Close` 接口，确保退出前写完最后一批数据
3. 将 go-micro `Greeter` demo 替换成真正的任务控制 API
4. 增加任务运行状态和统计信息
5. 增加最小本地 demo，避免依赖真实外网站点即可跑通完整链路
6. 增加 GitHub Actions，自动执行 `go test ./...` 和 `go build ./...`
