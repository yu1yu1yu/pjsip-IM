# PJSUA2 Python 绑定（macOS）

已在 Apple Silicon / macOS 上通过基础验证。Python 环境由 uv 创建与管理。

## 环境

- 源码：当前仓库 PJPROJECT 2.17-dev，保留仓库已有改动。
- Python：uv 管理的 CPython 3.13.13 / arm64。
- 虚拟环境：项目根目录 `.venv-pjsua2`。
- SWIG：4.5.0，安装在虚拟环境中。
- 构建工具：本机 Apple Clang、CMake、Homebrew OpenSSL 3。
- 独立构建目录：`python-bindings/build`，不覆盖原有 `build`。

## 使用

```sh
cd /Users/yuyu/postgradu/network/pjproject
source .venv-pjsua2/bin/activate
python -c 'import pjsua2; print(pjsua2.__file__)'
python python-bindings/smoke_test.py
```

无需激活也可以直接运行 `.venv-pjsua2/bin/python`。后续 Electron 应启动这个
解释器运行通信服务，不能直接使用全局 Python，否则可能导入不到绑定。

## 重建

首次创建环境时执行（环境已存在时跳过第一行）：

```sh
uv venv --python 3.13.13 --managed-python .venv-pjsua2
uv pip install --python .venv-pjsua2/bin/python -r python-bindings/requirements-build.txt
.venv-pjsua2/bin/python python-bindings/build_binding.py
```

脚本读取原有 `build/CMakeCache.txt` 中的 PJ/SRTP 功能选项，在独立目录构建
底层静态库及官方 SWIG 接口，安装 `pjsua2.py` 和 `_pjsua2` 原生模块到虚拟环境。
若删除原有 CMakeCache，重建将使用上游默认选项，应重新核对 TLS、音频、SRTP。

这是对当前 CMake 源码构建的本地适配。上游 Python Makefile 依赖 `build.mak`，
本仓库已有构建没有该文件，因此没有直接执行 `make install`。
绑定由 CMake 安装，并非 PyPI wheel，`uv pip list` 不会把它列为已安装发行包。
重建或更换虚拟环境后，必须重新运行构建脚本；不要只复制 `pjsua2.py`。

## 验证范围

`smoke_test.py` 验证原生模块导入、Endpoint 创建/初始化/启动、事件处理、
编解码器枚举、MESSAGE 接口存在，以及正常关闭。
测试使用空音频设备，不打开麦克风，不创建网络传输，不联系 SIP 服务器。
尚未验证实际注册、MESSAGE 回调收发、音频通话、TLS 握手或公网穿透。

当前构建启用 CoreAudio、G.711/G.722 等、OpenSSL/TLS 和 SRTP；没有 Opus，视频关闭。
构建能力不代表相应功能已完成端到端测试。

## 范围与打包

本构建脚本仅适用于已验证的 macOS 路线；Windows 需独立构建验证。
原生模块动态依赖本机 Homebrew OpenSSL，当前不是可复制到其他电脑的独立安装包。
后续 Electron 分发时还需处理 Python 运行时、动态库、系统权限和签名。
PJSIP 对象应保持明确引用并有序销毁；Python 应自行驱动 SIP 事件，避免依赖垃圾回收。
