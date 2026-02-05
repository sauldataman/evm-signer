## Yubikey HSM 配置流程
1. 下载官方的 sdk，分别提供了 yubihsm-connector, 以及 yubihsm-shell 的工具。https://developers.yubico.com/YubiHSM2/Releases/
    以 mac 为例，安装命令如下：
   ```shell
    brew install --cask yubihsm2-sdk
    ```
2. 在终端里面输入 `yubihsm-connector -d` 启动连接 hsm。 连接成功会打印如下信息
    ``` text
    DEBU[0000] preflight complete cert= config= key= pid=10512 seccomp=false serial= syslog=false timeout=0s version=3.0.3
    DEBU[0000] takeoff TLS=false listen="localhost:12345" pid=10512
   ```
   同时在浏览器里面输入：http://127.0.0.1:12345/connector/status
   先试一下信息：
   ```text
    status=OK
    serial=*
    version=3.0.3
    pid=10512
    address=localhost
    port=12345
    ```
   也可以在终端输入 `curl http://127.0.0.1:12345/connector/status`
   显示内容同上
    
   **注意**
    如果 status != OK，说明 hsm 硬件连接有问题，请尝试重新插入对应的接口并重试。
3. 在终端中输入 `yubihsm-shell` 进入命令行操作模式。其本质也是命令行连接并保持和 `yubihsm-connector` 的对话。
   因此在 `yubihsm-shell` 中执行的每条命令，都可以在 `yubihsm-connector` 中看到对应输出的日志。
    ``` shell
   yubihsm-shell 
   connect # 与 yubihsm-connector 建立连接
   keepalive on # 与 yubihsm-connector 保持长连接。且[默认开启]。
   session open 1 password # session open authKey password. 其中 authKey 
    # 为用户创建的，同时设备自动创建了 authKey == 1 且 password == password 的默认登陆
    # 账户和密码用于测试使用。⚠️生产环境必须重新创建一个新的用户以及登陆密码，具体步骤见下文。
    # 终端返回 Created session 0，即创建了一个 sessionId == 0 的句柄。
   list objects 0 # 打印出 sessionId == 0 对象的列表，内容如下：
    # Found 256 object(s) 已存在的对象数。
    # objectId,   objectType,               objectSequence
    # id: 0x0001, type: authentication-key, sequence: 0
    # id: 0x0002, type: authentication-key, sequence: 0
    # id: 0x0008, type: asymmetric-key, sequence: 0
   ```

### HSM 添加新的身份以及密钥
1. 使用 put 命令添加新的身份
   ```shell
   put authkey 0 2 <your_label> 1,2,3 generate-asymmetric-key,export-wrapped,get-pseudo-random,put-wrap-key,import-wrapped,delete-asymmetric-key,sign-ecdsa sign-ecdsa,exportable-under-wrap,export-wrapped,import-wrapped <your_password>
   # 0: 当前 sessionId
   # 2: 要创建的 auth_key
   # <your_label>: 当前 ID 的 label
   # <your_password>: 为设置的登陆密码
   ```
   - 如果当前 ID 已经存在，则会展示如下错误。
      
      Failed to store authkey: An Object with that ID already exists
   - 创建成功则提示：

     Stored Authentication key 0x0003
2. 查看设置身份的详细信息
   ```shell
   get objectinfo 0 0x0002 authentication-key
   # id: 0x0002, type: authentication-key, algorithm: aes128-yubico-authentication, label: "yubico", length: 40, domains: 1:2:3, sequence: 0, origin: imported, capabilities: delete-asymmetric-key:export-wrapped:generate-asymmetric-key:get-pseudo-random:import-wrapped:put-wrap-key:sign-ecdsa, delegated_capabilities: export-wrapped:exportable-under-wrap:import-wrapped:sign-ecdsa
   ```
3. 测试新的管理身份登入 yubikey
   ```shell
   session open 2 <your_password>
   # 提示：Created session 1 则为成功
   ```

### HSM 中私钥生成步骤
1. 首先确定 Ethereum 私钥生成的算法是 ECC secp256k1。在 Yubikey 的
对照表中算法为 eck256。生成命令如下：
   ```shell
   generate asymmetric 0 1000 label_ecdsa_sign 1,2,3 exportable-under-wrap,sign-ecdsa eck256
   ```
2. 删除不再使用的 key pair
    ```shell
    delete 0 1000 asymmetric-key
    ```

### 如何查看 HSM 私钥生成后的地址 
1. 在 conf/config.yaml 文件中配置 hsm 的相关参数。配置完成后，启动签名机
即可打印出对应的地址。
2. 具体查看 conf/config.yaml.example 文件。
```text
account type: [HSM], index: [3], address: [0xf6994200339FB759de34dfC26052295Dfb922EB0]
```

### HSM 对象管理
| yubikey 中的对象创建数量和存储大小是有限制的。YubiHSM 2 设备都可以存储多达 256 个对象。 它们的总大小不能超过 126 KB。
| 当数量超过时可以手动管理对象的，删除部分不是用的对象。
| 在设备初始化时，用户的数据也是一个对象。因此在有资产的清空用户对象或者重置整个 hsm 设备是特别危险的，在没有
备份的情况下，将无法找回用户资产！

1. 查看对象列表以及数量
```shell
list objects 0 # 0 为当前 sessionId
```
2. 新增对象
```shell
generate asymmetric 0 1000 test_object 1,2 exportable-under-wrap,sign-ecdsa ed25519
```
3. 删除对象
```shell
delete 0 0x03e8 asymmetric-key
```

### 注意事项
1. 关于用户的 auth key，在 yubikey 中也被定义为 object。因此在删除
object 的时候，仅删除创建私钥的 id 即可。
2. 在物理级别恢复出厂设置时，需要用力按压金属边缘最少 10s。

### 参考
1. [yubikey 官方文档](https://developers.yubico.com/YubiHSM2/)
2. [yubikey 简要说明](https://zhuanlan.zhihu.com/p/545530838)
3. [yubikey 算法对照表](https://developers.yubico.com/YubiHSM2/Concepts/Algorithms.html)
4. [yubikey 设备重置](https://developers.yubico.com/YubiHSM2/Usage_Guides/Factory_reset.html)