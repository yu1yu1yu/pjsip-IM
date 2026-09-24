# uv 虚拟环境与 PJSUA2 Python 绑定教程

本教程适用于 macOS Apple Silicon。已验证 Python 3.13.13、SWIG 4.5.0 和当前 PJPROJECT 2.17-dev 源码。Windows 需要独立构建，不能复制 Mac 的 `.so` 使用。

## 1. 安装构建工具

需要 Xcode Command Line Tools 和 Homebrew。已安装的工具无需重复安装：

```sh
xcode-select --install
```

完成系统安装窗口中的操作后，通过 Homebrew 安装工具：

```sh
brew install uv cmake openssl@3
uv --version
cmake --version
clang --version
```

CMake 至少需要 3.28。未安装 Homebrew 时按 [Homebrew 官方说明](https://brew.sh/)安装。
SWIG 会安装进 Python 虚拟环境，不需要另行用 Homebrew 安装。

## 2. 获取源码

```sh
git clone https://github.com/yu1yu1yu/pjsip-IM.git
cd pjsip-IM
```

如果已有项目，直接进入自己的仓库根目录。后续命令全部从仓库根目录执行，不依赖作者电脑的绝对路径。

## 3. 用 uv 创建虚拟环境

```sh
uv venv --python 3.13.13 --managed-python .venv-pjsua2
uv pip install --python .venv-pjsua2/bin/python -r python-bindings/requirements-build.txt
```

uv 会按需下载所需 Python。环境已存在时跳过创建命令；不用 Conda，也不用全局 pip。
`.venv-pjsua2` 与常见的 `.venv` 是同一种虚拟环境，仅名称不同。本项目后续 Python 通信服务可以继续使用它。

激活并检查解释器：

```sh
source .venv-pjsua2/bin/activate
python -c 'import sys; print(sys.executable); print(sys.version)'
```

解释器应位于当前项目的 `.venv-pjsua2/bin/python`。退出环境执行 `deactivate`。

## 4. 首次配置 PJSIP

新克隆项目不包含本机配置和编译缓存，先创建本地配置头文件（已有文件保持不变）：

```sh
test -f pjlib/include/pj/config_site.h || touch pjlib/include/pj/config_site.h
```

首次配置采用基础音频、TLS/SRTP，关闭视频、Opus 和回声消除：

```sh
cmake -S . -B build \
  -DBUILD_TESTING=OFF \
  -DPJLIB_WITH_SSL=openssl \
  -DPJSIP_WITH_TLS=ON \
  -DPJMEDIA_WITH_AUDIODEV=ON \
  -DPJMEDIA_WITH_AUDIODEV_COREAUDIO=ON \
  -DPJMEDIA_WITH_SRTP=ON \
  -DSRTP_WITH_OPENSSL=ON \
  -DPJMEDIA_WITH_VIDEO=OFF \
  -DPJMEDIA_WITH_OPUS_CODEC=OFF \
  -DPJMEDIA_WITH_SPEEX_AEC=OFF \
  -DPJMEDIA_WITH_WEBRTC_AEC=OFF \
  -DPJMEDIA_WITH_WEBRTC_AEC3=OFF \
  -DOPENSSL_ROOT_DIR="$(brew --prefix openssl@3)"
```

这一命令生成配置，不需要再单独构建或安装系统级 PJSIP。已有自定义配置时先核对选项，命令会更新同名选项。
上游 CMake 支持仍标记为实验性，本仓库使用这条已验证的 macOS 路线。

## 5. 构建并安装绑定

```sh
.venv-pjsua2/bin/python python-bindings/build_binding.py
```

脚本会读取 `build/CMakeCache.txt` 中的 PJ/SRTP 选项，在 `python-bindings/build/` 独立构建：

1. 编译 PJSIP/PJSUA2 静态库。
2. 调用 SWIG 读取官方 `pjsip-apps/src/swig/pjsua2.i`，生成 Python 包装和 C++ 桥接代码。
3. 编译 `_pjsua2` 原生模块，安装到当前虚拟环境。
4. 自动执行 `smoke_test.py`，成功时输出 `"result": "PASS"`。

不需要 `sudo`，不需要向 `/usr/local` 安装库，也不要执行 `uv pip install pjsua2` 替代此步骤。
当前构建脚本依赖本机 OpenSSL 的自动发现；如果 Homebrew 使用非标准路径，应检查绑定构建日志中的 OpenSSL 路径。

| 生成位置 | 内容 |
| --- | --- |
| `python-bindings/build/generated/pjsua2.py` | SWIG 生成的 Python 接口 |
| `python-bindings/build/generated/pjsua2_wrap.cpp` | SWIG 生成的 C++ 桥接 |
| `.venv-pjsua2/lib/python3.13/site-packages/pjsua2.py` | 实际导入的 Python 文件 |
| 同一 `site-packages` 内的 `_pjsua2*.so` | Python 原生扩展 |

绑定由 CMake 安装，不是带发行包元数据的 wheel，因此 `uv pip list` 不列出 `pjsua2` 也不代表未安装。

## 6. 验证和日常使用

```sh
source .venv-pjsua2/bin/activate
python -c 'import sys, pjsua2; print(sys.executable); print(pjsua2.__file__)'
python python-bindings/smoke_test.py
python sip-demo/endpoint-demo.py --help
```

不激活环境也可使用 `.venv-pjsua2/bin/python` 运行脚本。消息测试见[示例教程](../sip-demo/README.md)。

`smoke_test.py` 只验证模块加载、库生命周期、事件处理与编解码器访问，使用空音频设备，不打开麦克风或连接服务器。
另外已完成本机两个进程的中文 MESSAGE 双向收发及 SIP 200 测试。实际音频、注册鉴权、TLS 握手和公网尚未验证。

## 7. VS Code 配置

1. 打开项目文件夹，安装并启用 Python 和 Pylance 扩展。
2. 使用 `Python: Select Interpreter`，选择项目的 `.venv-pjsua2/bin/python`。
3. 若运行正常但无法跳转，执行 `Developer: Reload Window` 刷新语言服务。
4. 在 `pjsua2` 上“转到定义”应进入 Python 包装文件；原生 `.so` 无法像 Python 源码一样继续跳转。

## 8. 常见问题

| 问题 | 处理方式 |
| --- | --- |
| `No module named pjsua2` | 输出 `sys.executable`，确认运行入口使用上述环境；重建环境后必须重新构建绑定 |
| 找不到 `_pjsua2` / 原生库加载失败 | 不要只复制 `pjsua2.py`；核对 Python 版本、CPU 架构和 OpenSSL 动态库，重新构建 |
| 找不到 `pj/config_site.h` | 按第 4 步创建本地头文件，已有内容不要覆盖 |
| `Address already in use` | 退出占用端口的旧示例，或给两个客户端选择不同端口 |
| 初始化时 segmentation fault | 检查顺序：`libCreate` → `libInit` → `libStart`，不能省略 `libCreate` |
| 能运行但编辑器无法跳转 | 检查解释器、Python/Pylance 扩展并 Reload Window |
| 收不到消息 | 先启动接收端；URI 中的端口必须与对方监听端口一致；当前仅支持本机测试 |

## 9. 重建与提交约定

修改源码后重新执行 `build_binding.py` 即可。更换 Python 版本、CPU 架构或移动仓库后，应重新创建环境；如 CMake 提示缓存路径不匹配，将旧 `python-bindings/build/` 移到仓库外备份后再构建。
如果移除了根目录 `build/CMakeCache.txt`，先重新执行第 4 步，避免回到上游默认功能配置。

提交源码、构建脚本、依赖版本与文档；不提交 `.venv-pjsua2/`、桥接生成文件、编译缓存和二进制。
注意根目录 `build/` 同时保存上游脚本，不能整体删除。当前模块动态依赖 Homebrew OpenSSL，尚不是可直接分发的 Electron 安装包。

官方资料：[uv 虚拟环境](https://docs.astral.sh/uv/pip/environments/)、[PJSUA2 构建](https://docs.pjsip.org/en/latest/pjsua2/building.html)。
