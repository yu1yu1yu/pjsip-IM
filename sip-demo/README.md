# 本机 SIP MESSAGE 使用教程

入口为 `endpoint-demo.py`。`message-demo.py` 目前是空文件，不要用它启动程序。
先完成 [uv 环境和 Python 绑定构建](../python-bindings/README.md)。以下命令均在仓库根目录执行。

## 1. 启动接收端 B

打开第一个终端：

```sh
source .venv-pjsua2/bin/activate
python sip-demo/endpoint-demo.py --port 5070 --username bob
```

看到 `账号已就绪：sip:bob@127.0.0.1:5070` 后保持运行。

## 2. 启动发送端 A

打开第二个终端并进入同一仓库根目录：

```sh
source .venv-pjsua2/bin/activate
python sip-demo/endpoint-demo.py --port 5060 --username alice \
  --to "sip:bob@127.0.0.1:5070" --text "你好 Bob，我是 Alice"
```

预期 B 打印发送者、正文和 `text/plain`；A 打印 `发送结果：200 OK`。
该状态表示对端 SIP 用户代理已接受消息，不表示用户已读或消息已保存到数据库。

## 3. 反向发送

先用 `Ctrl+C` 停止两个实例。再在两个终端依次启动：

```sh
# 终端 A：先等待接收
python sip-demo/endpoint-demo.py --port 5060 --username alice
```

```sh
# 终端 B：发送给 A
python sip-demo/endpoint-demo.py --port 5070 --username bob \
  --to "sip:alice@127.0.0.1:5060" --text "你好 Alice，我是 Bob"
```

每个终端都需要使用项目虚拟环境。当前程序每次启动最多发送一条消息，之后一直处理事件，不支持在运行中直接输入新消息。

## 参数

| 参数 | 默认值 | 说明 |
| --- | --- | --- |
| `--port` | `5060` | 本机 UDP 监听端口，两个实例不能相同 |
| `--username` | `alice` | 本地 SIP 身份名称，不是服务器登录 |
| `--to` | 无 | 对方完整 SIP URI；省略时仅等待接收 |
| `--text` | `你好，这是第一条 SIP 消息` | 指定 `--to` 后发送的正文 |

## 代码阅读顺序

1. `main()`：创建 Endpoint、配置 UDP、启动库。
2. `create_sip_account()`：创建不注册的本地账号。
3. `MyBuddy`：配置目标、关闭在线状态订阅。
4. `send_text()`：发起 MESSAGE 请求。
5. `Account.onInstantMessage()`：接收消息回调。
6. `Account.onInstantMessageStatus()`：发送结果回调，响应码字段是 `prm.code`。
7. `libHandleEvents()`：主线程持续处理事件。
8. `finally`：先释放 Buddy，再关闭 Account，最后销毁库。

发送方法返回不等于消息已经送达，必须等待结果回调。

## 调试与范围

- 抓包选择 Mac 的回环接口 `lo0`，Wireshark 显示过滤器可用 `sip`。
- 当前只绑定 `127.0.0.1`，不能用这份配置直接进行局域网双机通信。
- 没有注册服务器、鉴权、离线投递或历史存储。
- 使用空音频设备，不采集麦克风，也不进行语音通话。
- 关闭接收端后再发消息，可观察传输错误或超时，结果不会保证立即返回。
- 不要在事件循环中直接使用阻塞式 `input()`，否则会暂停 SIP 事件处理。
