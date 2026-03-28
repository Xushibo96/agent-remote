# agent-remote

`agent-remote` 是一个面向 AI Agent 的远程开发工具，专门解决跨设备开发时的三类核心问题：

- 远程执行命令，不依赖运行时手工输入密码
- 本地与远端之间的文件上传、下载、双向同步
- 以结构化、低 token 占用的方式返回执行结果，避免上下文被日志快速打满

项目使用 Go 实现，当前提供：

- CLI 入口
- MCP 适配层
- SSH 远程命令执行
- SFTP 文件同步
- rsync 参数构造与后端选择逻辑
- 加密配置存储
- session 持久化与增量读取

## Why

很多 agent 在实际开发环境里会遇到这些问题：

- 远程执行命令需要用户临时输入密码，自动化链路断掉
- 文件同步依赖人工上传下载，开发闭环慢
- 命令输出太长，agent 上下文很快被占满
- 跨设备、多次调用之间缺少稳定 session，前一次结果下一次读不到

`agent-remote` 的目标就是把这些能力收敛成一个 agent 友好的本地代理工具。

## Features

- 非交互式目标配置保存
  - 用户输入密码时可以是明文
  - 落盘前会加密保存
- SSH 远程命令执行
  - `exec start`
  - `exec read`
  - `exec stop`
- 文件同步
  - `sync upload`
  - `sync download`
  - `sync bidir`
- 结果裁剪与结构化返回
  - 事件化输出
  - cursor 增量读取
  - 面向 agent 的紧凑 JSON
- 会话持久化
  - `exec start` 和 `exec read` 可以跨 CLI 进程工作

## Project Layout

```text
cmd/agent-remote/         CLI 入口
internal/app/             编排层
internal/connection/      SSH 连接复用与能力探测
internal/credential/      凭据管理与脱敏
internal/exec/            远程命令执行
internal/mcp/             MCP 适配层
internal/rsync/           rsync runner 与参数构造
internal/session/         session store 与 cursor
internal/sync/            SFTP 同步与双向差异规划
internal/secret/          keyring 与 AES-GCM
test/integration/         集成测试
test/perf/                benchmark
```

## Requirements

- Go 1.22+
- 可访问目标主机的 SSH 网络
- 本机可用 keyring
- 如果需要 rsync 优化路径：
  - 本机安装 `rsync`
  - 远端安装 `rsync`

## Quick Start

### 1. Build

```bash
make build
```

也可以直接调用脚本：

```bash
./scripts/build.sh
GOOS=linux GOARCH=amd64 ./scripts/build.sh
```

### 2. Add a target

```bash
./bin/agent-remote target add \
  --id router \
  --host 192.168.1.254 \
  --user admin \
  --password 'your-password' \
  --known-hosts-policy insecure
```

### 3. Execute a remote command

```bash
./bin/agent-remote exec start --target router --command 'ls'
./bin/agent-remote exec read --session <session-id>
```

### 4. Upload / download files

```bash
./bin/agent-remote sync upload \
  --target router \
  --local ./local-dir \
  --remote /tmp/remote-dir

./bin/agent-remote sync download \
  --target router \
  --local ./download-dir \
  --remote /tmp/remote-dir
```

### 5. Bidirectional sync

```bash
./bin/agent-remote sync bidir \
  --target router \
  --local ./workspace \
  --remote /srv/workspace \
  --conflict newer-wins
```

## CLI Commands

```bash
agent-remote target add
agent-remote target list
agent-remote sync upload
agent-remote sync download
agent-remote sync bidir
agent-remote exec start
agent-remote exec read
agent-remote exec stop
agent-remote job status
```

默认输出是结构化 JSON，适合脚本和 agent 直接消费。

## Config and Security

默认配置目录：

```text
macOS: ~/Library/Application Support/agent-remote
Linux: $XDG_CONFIG_HOME/agent-remote or ~/.config/agent-remote
```

当前会写入：

- `config.json`
- `sessions.json`

安全行为：

- 目标密码不会以明文保存在配置文件里
- 本地配置使用加密 envelope 持久化
- session 输出默认以结构化事件返回
- 错误路径会做敏感信息脱敏

## MCP

项目内已经包含 MCP 适配层，当前暴露的工具面包括：

- `target_add`
- `target_list`
- `sync_upload`
- `sync_download`
- `sync_bidir`
- `job_status`
- `exec_start`
- `exec_read`
- `exec_stop`

如果你要把它接到 agent runtime，可以直接复用 `internal/mcp` 下的适配器和 server。

## Development

### Run tests

```bash
go test ./...
```

### Package

```bash
make package
make release
```

对应脚本：

```bash
./scripts/package.sh
./scripts/release.sh
```

### Run benchmarks

```bash
go test ./test/perf -run '^$' -bench . -benchtime=1x
```

### Integration tests

```bash
go test ./test/integration
```

`test/integration/exec_integration_test.go` 默认需要显式开启环境变量后才会跑真实 SSH 集成路径。

## Current Status

当前版本已经可用于：

- 保存远程目标
- 执行远程命令
- 读取持久化 session 输出
- 通过 SFTP 进行同步

当前仍有继续增强空间：

- rsync 数据面进一步实装
- 更完整的 MCP 对外运行模式
- 长时间命令的独立后台执行模型
- 更细的输出预算策略

## License

如果你准备开源，建议补一个 `LICENSE` 文件。MIT 或 Apache-2.0 都比较合适。
