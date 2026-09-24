# PJSIP IM 课程项目

基于 PJPROJECT 学习 SIP 的即时通讯项目。目前使用 Python + PJSUA2 实现本机 UDP 文字消息收发；后续计划接入 Electron 桌面界面，并扩展登录、联系人、历史消息、语音和跨平台通信。

## 当前状态

- 已验证：macOS Apple Silicon、uv 虚拟环境、PJSUA2 Python 绑定构建。
- 已验证：两个本机 Python 进程互发中文 SIP MESSAGE、收到 `200 OK`、正常退出。
- 尚未实现：Electron 界面、服务器注册登录、联系人和消息持久化、语音/视频通话。
- 尚未验证：Windows 客户端、局域网双机和公网互通。

当前示例绑定 `127.0.0.1`，不需要 SIP 服务器；本地账号只是身份，不涉及密码鉴权。

## 文档入口

- [uv 环境创建、绑定构建与常见问题](python-bindings/README.md)
- [文字消息示例使用教程](sip-demo/README.md)

## 第一次运行

先按照[构建教程](python-bindings/README.md)安装工具、创建 `.venv-pjsua2` 并编译绑定。
以下命令均在项目根目录执行，两个终端分别激活虚拟环境。

终端 B，先启动接收者：

```sh
source .venv-pjsua2/bin/activate
python sip-demo/endpoint-demo.py --port 5070 --username bob
```

终端 A，发送一条消息：

```sh
source .venv-pjsua2/bin/activate
python sip-demo/endpoint-demo.py --port 5060 --username alice \
  --to "sip:bob@127.0.0.1:5070" --text "你好 Bob"
```

B 应显示消息正文，A 应显示 `发送结果：200 OK`。`200 OK` 表示对端接受消息，不代表已读。
每次启动最多发送一条消息，随后继续接收；按 `Ctrl+C` 退出。

## 项目结构

| 路径 | 说明 |
| --- | --- |
| `sip-demo/endpoint-demo.py` | 当前消息收发入口 |
| `sip-demo/message-demo.py` | 预留空文件，目前不使用 |
| `python-bindings/` | SWIG/CMake 构建入口、验证脚本与教程 |
| `.venv-pjsua2/` | uv 创建的 Python 虚拟环境，不提交 |
| `python-bindings/build/` | 自动生成的桥接代码和编译产物，不提交 |
| `pjlib/`、`pjmedia/`、`pjsip/`、`pjnath/` | 上游通信库源码 |
| `build/` | 含上游构建脚本，也包含被忽略的本机 CMake 产物；不要整体删除 |

`.venv-pjsua2` 是普通 Python 虚拟环境，其中安装了绑定；PJSIP 本身是 C/C++ 库。
自动生成的 `pjsua2.py` 和原生模块不手工修改。提交源码、构建脚本和文档即可。

## 后续开发顺序

1. 增加交互式消息输入与 Electron 进程通信。
2. 实现 Windows 绑定与局域网双机 MESSAGE。
3. 接入 SIP 服务器，学习 REGISTER、鉴权和消息路由。
4. 增加联系人、SQLite 历史消息和一对一语音。
5. 验证公网 NAT 穿透，最后扩展视频。

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
