# PJSIP IM — Go 开发分支

本分支 `codex/go-development` 使用 Go 开发即时通讯应用，通过 cgo 绑定本仓库的 PJSUA C 接口。原来的 `master` 保留 Python + PJSUA2 开发版本。

当前实现 Windows x64 上的本机 UDP 中文 SIP MESSAGE 收发、发送结果回调和端点生命周期管理。无需 Python 或 SIP 服务器。Go 应用逻辑位于 `go-bindings/`；PJSIP 协议栈仍由 C/C++ 源码编译。

## 构建与运行

需要 Go 1.23+、CMake 3.28+、Ninja、MinGW-w64 GCC/G++ x64。仓库根目录执行：

```powershell
.\go-bindings\build.ps1 -Test
```

先在终端 B 启动接收者：

```powershell
.\go-bindings\build\bin\sip-im.exe -port 5070 -username bob
```

终端 A 发送中文消息：

```powershell
.\go-bindings\build\bin\sip-im.exe -port 5060 -username alice -to 'sip:bob@127.0.0.1:5070' -text '你好 Bob'
```

Bob 收到 `message` 事件，Alice 收到 `status` 事件和 `200 OK`；按 Ctrl+C 退出。消息接受回执不代表已读。

完整说明见 [Go 环境、源码构建、API 与排错](go-bindings/README.md)。

## 目录与分支

| 路径或分支 | 用途 |
| --- | --- |
| `codex/go-development` | 当前 Go 开发分支 |
| `master` | 原 Python 开发分支 |
| `go.work` | Go 工作区入口 |
| `go-bindings/pjsip/` | Go API、线程与生命周期管理、集成测试 |
| `go-bindings/native/` | PJSUA C 桥接层 |
| `go-bindings/cmd/sip-im/` | Go 消息收发入口 |
| `go-bindings/build.ps1` | 从底层库到 Go 程序的一键构建与验证 |
| `go-bindings/build/` | 原生编译产物与 exe，不提交 Git |
| `python-bindings/`、`sip-demo/` | 继承的 Python 源码与文档，不参与 Go 构建 |
| `pjlib/`、`pjmedia/`、`pjsip/`、`pjnath/` | 上游协议栈源码 |

首版只监听 `127.0.0.1`，每进程一个端点。尚未实现服务器注册登录、持久化、Electron 界面及语音视频通话；其他平台与跨机通信尚未验证。当前构建关闭 TLS、SRTP 和音视频设备。

切换语言开发线前先提交或保存当前修改，再使用 `git switch master` 或 `git switch codex/go-development`。构建目录被忽略，不随分支切换，不能将 Go 和 Python 的原生编译产物混用。

## 上游项目与许可

本仓库基于 [pjsip/pjproject](https://github.com/pjsip/pjproject)。保留其源码、许可与下方文档信息；项目依赖的许可见仓库 COPYING 文件及[上游许可说明](https://www.pjsip.org/licensing.htm)。下方徽章反映上游项目状态，不代表本课程项目的测试结果。

---


[![CI Linux](https://github.com/pjsip/pjproject/actions/workflows/ci-linux.yml/badge.svg?branch=master)](https://github.com/pjsip/pjproject/actions/workflows/ci-linux.yml)
[![CI Mac](https://github.com/pjsip/pjproject/actions/workflows/ci-mac.yml/badge.svg?branch=master)](https://github.com/pjsip/pjproject/actions/workflows/ci-mac.yml)
[![CI Windows](https://github.com/pjsip/pjproject/actions/workflows/ci-win.yml/badge.svg?branch=master)](https://github.com/pjsip/pjproject/actions/workflows/ci-win.yml)
[![Bitrise iOS](https://img.shields.io/bitrise/70e79dc5-cae8-4cb7-a6cd-9a5bd3f3270f?token=tnXk2DZ71Zmd0qDMhFgiBg&label=CI%20iOS)](https://app.bitrise.io/app/70e79dc5-cae8-4cb7-a6cd-9a5bd3f3270f)
[![Bitrise Android](https://img.shields.io/bitrise/e4b6aade20ea9eb3?token=byZU0e1BJn_VYg2YuAs-cA&label=CI%20Android)](https://app.bitrise.io/app/e4b6aade20ea9eb3)
<BR>
[![OSS-Fuzz](https://oss-fuzz-build-logs.storage.googleapis.com/badges/pjsip.png)](https://oss-fuzz-build-logs.storage.googleapis.com/index.html#pjsip)
[![Coverity-Scan](https://scan.coverity.com/projects/905/badge.svg)](https://scan.coverity.com/projects/pjsip)
[![CodeQL](https://github.com/pjsip/pjproject/actions/workflows/codeql-analysis.yml/badge.svg?branch=master)](https://github.com/pjsip/pjproject/actions/workflows/codeql-analysis.yml)
[![docs.pjsip.org](https://readthedocs.org/projects/pjsip/badge/?version=latest)](https://docs.pjsip.org/en/latest/)


# PJSIP

PJSIP is a free and open source multimedia communication library written in C with high level API in C, C++, Java, C#, and Python languages. It implements standard based protocols such as SIP, SDP, RTP, STUN, TURN, and ICE. It combines signaling protocol (SIP) with rich multimedia framework and NAT traversal functionality into high level API that is portable and suitable for almost any type of systems ranging from desktops, embedded systems, to mobile handsets.

## Getting PJSIP

- Main repository: https://github.com/pjsip/pjproject
- Releases: https://github.com/pjsip/pjproject/releases


## Documentation

Main documentation site: https://docs.pjsip.org

Table of contents:

- Overview
  - [Overview](https://docs.pjsip.org/en/latest/overview/intro.html)
  - [Features (Datasheet)](https://docs.pjsip.org/en/latest/overview/features.html)
  - [License](https://docs.pjsip.org/en/latest/overview/license.html)
- **Getting started**
  - [Getting PJSIP](https://docs.pjsip.org/en/latest/get-started/getting.html)
  - [General Guidelines](https://docs.pjsip.org/en/latest/get-started/general_guidelines.html)
  - [Android](https://docs.pjsip.org/en/latest/get-started/android/index.html)
  - [iPhone](https://docs.pjsip.org/en/latest/get-started/ios/index.html)
  - [Mac/Linux/Unix](https://docs.pjsip.org/en/latest/get-started/posix/index.html)
  - [Windows](https://docs.pjsip.org/en/latest/get-started/windows/index.html)
  - [Windows Phone](https://docs.pjsip.org/en/latest/get-started/windows-phone/index.html)
- PJSUA2 - High level API guide
  - [Introduction](https://docs.pjsip.org/en/latest/pjsua2/intro.html)
  - [Building PJSUA2](https://docs.pjsip.org/en/latest/pjsua2/building.html)
  - [General concepts](https://docs.pjsip.org/en/latest/pjsua2/general_concept.html)
  - [Hello world!](https://docs.pjsip.org/en/latest/pjsua2/building.html)
  - [Using PJSUA2](https://docs.pjsip.org/en/latest/pjsua2/using/index.html)
  - [Sample applications](https://docs.pjsip.org/en/latest/pjsua2/samples.html)
- Specific guides
  - [Audio](https://docs.pjsip.org/en/latest/specific-guides/index.html#audio)
  - [Audio Troubleshooting](https://docs.pjsip.org/en/latest/specific-guides/index.html#audio-troubleshooting)
  - [Build and integration](https://docs.pjsip.org/en/latest/specific-guides/index.html#build-integration)
  - [Development and programming](https://docs.pjsip.org/en/latest/specific-guides/index.html#development-programming)
  - [Media](https://docs.pjsip.org/en/latest/specific-guides/index.html#media)
  - [Network and NAT](https://docs.pjsip.org/en/latest/specific-guides/index.html#network-nat)
  - [Performance and footprint](https://docs.pjsip.org/en/latest/specific-guides/index.html#performance-footprint)
  - [Security](https://docs.pjsip.org/en/latest/specific-guides/index.html#security)
  - [SIP](https://docs.pjsip.org/en/latest/specific-guides/index.html#sip)
  - [Video](https://docs.pjsip.org/en/latest/specific-guides/index.html#video)
  - [Other](https://docs.pjsip.org/en/latest/specific-guides/index.html#other)
- API reference
  - [PJSUA2](https://docs.pjsip.org/en/latest/api/pjsua2/index.html) - high level API (Java/C#/Python/C++/swig)
  - [PJSUA-LIB](https://docs.pjsip.org/en/latest/api/pjsua-lib/index.html) - high level API (C)
  - [PJSIP](https://docs.pjsip.org/en/latest/api/pjsip/index.html) - SIP stack
  - [PJMEDIA](https://docs.pjsip.org/en/latest/api/pjmedia/index.html) - media framework
  - [PJNATH](https://docs.pjsip.org/en/latest/api/pjnath/index.html) - NAT traversal helper
  - [PJLIB-UTIL](https://docs.pjsip.org/en/latest/api/pjlib-util/index.html) - utilities
  - [PJLIB](https://docs.pjsip.org/en/latest/api/pjlib/index.html) - portable library
