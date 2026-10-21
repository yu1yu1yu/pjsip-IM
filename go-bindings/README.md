# Go 开发环境与 PJSIP 绑定

本分支 `codex/go-development` 使用 Go 开发应用层，通过 cgo 调用本仓库的 PJSUA C 库。
Python 开发版本保留在原来的 `master` 分支；本分支继承的 Python 示例仅作历史参考，不参与 Go 构建。
PJSIP 本身仍是 C/C++ 库，Go 绑定不等于把协议栈改写成 Go。

## 工具与环境

Windows x64，Go 1.23 或以上、CMake 3.28 或以上、Ninja、MinGW-w64 GCC/G++ x64。
将这些工具的可执行文件目录加入当前终端的 PATH，并确认：

```powershell
go version
gcc -dumpmachine
g++ --version
cmake --version
ninja --version
```

GCC 目标应为 `x86_64-w64-mingw32`，与 Go 的 `windows/amd64` 匹配。
本机使用现有的 Go 1.23.2、GCC 13.2.0、CMake 3.31.0-rc1，无需 Python、Conda、SWIG，也没有外部 Go 模块依赖。
本流程使用 MinGW，不能混用 MSVC 编译的 `.lib` 或其他平台的 Python `.so`。

## 从源码构建

在仓库根目录执行：

```powershell
.\go-bindings\build.ps1 -Test
```

脚本自动配置 CMake，以三个并行任务编译 PJSIP 和桥接 DLL，再编译 Go 示例；`-Test` 运行 Go 测试和 `go vet`。
再次执行为增量构建。构建输出和 Go 缓存仅放在 `go-bindings/build/`、`go-bindings/.cache/`，不修改全局 Go 设置。
如果系统执行策略阻止脚本，可使用当前进程级策略执行：

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File .\go-bindings\build.ps1 -Test
```

主要产物：

| 路径 | 用途 |
| --- | --- |
| `build/bin/sip-im.exe` | Go 消息收发程序 |
| `build/bin/libpjgo.dll` | PJSIP 原生桥接库，须与 exe 一起保留 |
| `build/lib/libpjgo.dll.a` | Go/cgo 链接时使用的导入库 |
| `build/pjproject/` | PJSIP 静态库、生成头文件和编译缓存 |

以上路径相对 `go-bindings/`。产物不提交 Git。修改底层源码后重新运行构建脚本；更换编译器或架构时使用新的构建目录或先备份现有构建目录。

## 两个终端收发中文消息

均从仓库根目录运行。先在终端 B 启动 Bob：

```powershell
.\go-bindings\build\bin\sip-im.exe -port 5070 -username bob
```

终端 A 启动 Alice 并发消息：

```powershell
.\go-bindings\build\bin\sip-im.exe -port 5060 -username alice -to 'sip:bob@127.0.0.1:5070' -text '你好 Bob'
```

Bob 输出 `kind: message` 及中文正文；Alice 输出 `kind: status`、相同的发送 ID、`code: 200` 和 `body: OK`。
`200 OK` 表示对端接受消息，不代表已读。每个程序启动时可发一条消息，之后保持接收；按 Ctrl+C 退出。
可加 `-duration 10s` 定时退出，或使用 `-version` 查看实际链接的 PJSIP 版本。
Go 层事件为逐行 JSON；PJSIP 初始化及错误日志可能混在标准输出中，机器消费时应筛选 JSON 行。
收件事件的 `peer` 为原始 SIP From 身份，不能直接用作回复地址；本机模式应向 `reply_to` 发送回复，该字段包含实际 UDP 来源端口。

## 继续开发 Go 代码

根目录的 `go.work` 将 `go-bindings` 设为工作区模块，编辑器可从仓库根目录打开。
Go 包路径为 `pjsip-im/go-bindings/pjsip`；这是仓库内模块名，尚未作为公开模块发布。

```go
ep, err := pjsip.New(pjsip.Config{Port: 5060, Username: "alice"})
if err != nil {
    return err
}
defer ep.Close()
id, err := ep.Send("sip:bob@127.0.0.1:5070", "你好")
// 检查 err；持续消费 ep.Events()，用 id 关联 status 事件。
```

| API | 说明 |
| --- | --- |
| `Version()` | 实际原生库版本 |
| `New(Config)` | 创建并启动端点 |
| `Send(uri, body)` | 提交消息，返回消息 ID；提交成功不等于送达 |
| `Events()` | 接收 `message` 和 `status` 事件 |
| `Done()`、`Err()` | 端点终止通知和终止错误 |
| `Close()` | 等待销毁完成，可重复调用 |

所有原生调用通过固定系统线程执行，禁用 PJSUA 信令与媒体轮询工作线程；应用可从不同 goroutine 调用 Send/Close。
每个进程仅允许一个活跃端点，多个身份使用不同进程。Go 不向 C 保存 Go 指针；字符串在调用期间分配，返回后释放，回调数据复制到 Go。
必须持续消费事件；原生队列或 Go 事件通道满时端点以错误退出，避免静默丢失消息。

构建后手动运行 Go 工具：

```powershell
$env:CGO_ENABLED = '1'
$env:CC = (Get-Command gcc).Source
$env:PATH = "$(Resolve-Path .\go-bindings\build\bin);$env:PATH"
$env:GOCACHE = "$PWD\go-bindings\.cache\go-build"
Set-Location go-bindings
go test -count=1 ./...
go vet ./...
go build -o build/bin/sip-im.exe ./cmd/sip-im
```

测试进程和 `go run` 的临时 exe 需要通过 PATH 找到 DLL；成品 exe 则从自身目录加载同目录 DLL。

## 当前边界与排错

已在本机 Windows x64 验证：从源码编译原生库和 Go 程序、两进程中文与 emoji 往返及双方 `200 OK`、消息 ID 关联、端口占用失败后重建、单进程端点限制、重复关闭与关闭后发送检查；`go test -count=1 ./...` 和 `go vet ./...` 通过。

- 首版仅绑定端点生命周期、UDP MESSAGE、收件通知与发送回执，尚未包装完整 PJSUA/PJSUA2 API。
- 只监听 `127.0.0.1`，无 SIP 服务器注册或密码认证；未验证局域网、公网、Linux、macOS。
- 单条正文最多 2048 字节有效 UTF-8；关闭自动切换 TCP。本地短文本以外的用途需扩展传输和报文限制。
- 构建关闭 TLS、SRTP、音视频设备和视频功能；仍会编译 PJSUA 依赖的媒体基础库。当前不能据此宣称支持语音通话或加密通信。
- `cannot find -lpjgo`：先执行构建脚本，确认 `build/lib` 导入库存在。
- 启动提示缺少 DLL 或退出码 `0xc0000135`：确认 `libpjgo.dll` 在 exe 目录或 PATH；如果换用其他 MinGW，需检查其额外运行库依赖。
- `Address already in use`：换端口，或退出占用端口的端点。
- `only one SIP endpoint per process`：关闭旧端点，或使用独立进程。
- PJSIP 错误码通过 `pj_strerror` 转成文本；收到非 200 的 `status` 时，检查目的 URI、接收端状态和网络。
- 排查传输问题时，可在启动前设置 `$env:PJGO_LOG_LEVEL = '4'`（允许 0..5，默认 1）；详细日志会包含消息正文，排查后移除此环境变量。

本次还修复了上游 CMake 的 Windows 位数判断：使用目标指针宽度设置 `PJ_WIN64`，避免 MinGW x64 下错误使用 32 位 `pj_ssize_t`。

参考：[Go cgo 文档](https://pkg.go.dev/cmd/cgo)、[LockOSThread](https://pkg.go.dev/runtime#LockOSThread)。PJSUA API 以本仓库 `pjsip/include/pjsua-lib/pjsua.h` 为准。许可证沿用仓库的 [COPYING](../COPYING)。
