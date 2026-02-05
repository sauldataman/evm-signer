# Signer

> 提供安全的交易签名服务
>
> 依赖环境：Go >= 1.19

## 配置流程

### 创建配置文件

```shell
$ cp conf/config.yaml.example conf/config.yaml
$ cp conf/rule.json.example conf/rule.json
$ mkdir .keystore
```

### 配置 config.yaml 文件

#### 必填项

```markdown
1. listen.port // 配置服务启动的端口号
2. auth.ip // IP 白名单 127.0.0.1,16.163.90.199,2407:cdc0:b02d::1039,110.184.67.202
3. account
    * 签名机解析账户类型支持 6 种，分别为 Keystore, EvMnemonic,
      EncryptedMnemonic, PlainMnemonic, PlainPrivateKey, HSM
    * 其中建议 PlainMnemonic, PlainPrivateKey 仅供测试使用
    * Keystore 为单个私钥加密出的 keystore 文件
    * EvMnemonic 虚拟助记词，为一组私钥加密出的 keystore 文件
    * EncryptedMnemonic 为加密后的助记词
    * HSM 则为 yubikey 的配置选项
    * account 中的 type 为常量且区分大小写。填写时请务必注意！
```

### rule 的匹配文件

使用 JSON 格式的规则文件进行交易验证，参照 `conf/rule.json.example` 配置。

#### json 规则验证

主要填写 conditions 的内容，conditions 是个数组，数组间是 and 关系，即：必须同时满足才会判定规则匹配。

```json
[
  {
    "name": "gmx exchange router",
    "chain_id": 42161,
    "conditions": [
      {
        "field": "to",
        "symbol": "==",
        "value": "0x7C68C7866A64FA2160F78EEaE12217FFbf871fa8"
      }
    ]
  }
]
```

### Account 类型

```markdown
1. PlainMnemonic
   明文助记词，支持 12|15|18|21|24 个单词长度的助记词
2. PlainPrivateKey
   明文私钥，带0x前缀共66个字符
3. Keystore
   私钥加密后的json文件，请在key字段填写密钥文件路径
4. EncryptedMnemonic
   使用keystore相同算法生成的的加密助记词json文件，请在key字段填写密钥文件路径
   建议使用 CS Wallet 对助记词进行非对称加密
5. HSM
   yubikey/AWS CloudHSM 硬件，配置参考hsm_yubikey和hsm_aws文档
6. EvMnemonic
   虚拟助记词，由以上5种账户类型组合而成，可将多个不同格式的账户对外暴露为一个助记词账户
```

#### Account 配置示例以及解释

##### PlainMnemonic

```yaml
type: PlainMnemonic
key:  # 明文助记词
```

##### PlainPrivateKey

```yaml
type: PlainPrivateKey
key: # 明文私钥
```

##### EncryptedMnemonic

```yaml
type: EncryptedMnemonic
key: # 加密后的助记词存放路径
pass: # 解密密码，选填项，如果不填则需要在程序启动时，在命令行输入密码
index: 0-10,11,20 # 使用助记词的指定下标，其中 - 表示连续的区间，前闭后开。, 分割的为单个 index
```

##### Keystore

```yaml
type: Keystore
key: # keystore 存放路径
pass: # 解密密码，选填项，如果不填则需要在程序启动时，在命令行输入密码
```

##### HSM

```yaml
type: HSM
provider: aws
pin: hsm_user:9g*KJ09TxMcbq#ns # 可测试使用的 pin code
private_key_id: 262158 # 可测试使用的 id
public_key_id: 262159 # 可测试使用的 id
```

##### HSM

```yaml
type: HSM # HSM
connector_url: localhost:12345 # 新增字段，当 provider == yubihsm-connector, 如果不填，默认为：localhost:3456
provider: "yubihsm-connector"
pin: 2:password2 # 可测试使用的 pin code
private_key_id: 1000 # 可测试使用的 id
```

##### MultiHSM

支持同一 HSM 设备多个 object 的连接。

```yaml
type: MultiHSM
provider: yubihsm-connector
connector_url: localhost:12345
private_key_ids: 1001-1020
```

##### EvMnemonic

key的数量可以任意多个，支持 `use_last_pass`，在该场景下 keys 解析是有顺序的，即始终按照 key 由小到大进行解析。
当不设置 `use_last_pass` 时默认为 true。当 `pass` 不为空 且 `use_last_pass` 为 true 时，优先使用 `pass` 的内容。

```yaml
type: EvMnemonic
keys:
  1:
    type: Keystore
    key: ./0xaaaa.json
    pass: "******"
    use_last_pass: true
  2:
    type: Keystore
    key: ./0xaaaa.json
    pass: "******"
    use_last_pass: true
  3:
    type: PlainPrivateKey
    key: xxxxxxxxxxxx
  4:
    type: PlainMnemonic
    key: "word1 word2 word3 ... word12"
    index: "0"
  5:
    type: Keystore
    key: ./0xaaaa.json
    pass: "******"
    use_last_pass: false
  6:
    type: HSM
    connector_url: localhost:12345 # 新增字段，当 provider == yubihsm-connector, 如果不填，默认为：localhost:3456
    provider: "yubihsm-connector"
    pin: "2:password1"
    private_key_id: 1000
    use_last_pass: true
```

### 配置 rule.json 文件

根据交易的基本信息，配置允许签名的规则。
可配置的字段验证如下：

```markdown
1. to
   除转账外就是要调用的合约地址，即对应transactions里面的 to 字段；
2. data_selector
   调用的目标合约的方法，对应transactions里面的data字段的前4个字节，例如 `0x12345678`
2. eip712.domain.name/eip712.domain.version/eip712.domain.chainId/eip712.domain.verifyingContract
   请查询eip712协议
```

## Build

```shell
$ export CGO_ENABLED=1 #必须确保该配置打开，否则关于 aws cloudhsm 部分将会编译错。
$ go build -o signer
```

## 启动服务

```shell
# 规则文件从 conf 目录加载（与 config.yaml 相同的搜索路径）
# 默认加载 conf/rule.json

./signer start --port 8080

# 指定不同的规则文件名（仍从 conf 目录加载）
./signer start --port 8080 --rule rule-prod.json
```

## 签名机暴露公网的方法(与 composer 配合使用)

### ngork

1. 登录ngrok 官网  https://ngrok.com/
2. 在dashboard找到 ngrok客户端, 下载
3. 找到Key, Getting Started - Your Authtoken
4. 按照文档步骤, 配置本机的ngrok
5. 最后启动ngrok, ./ngrok http 8080 , 8080是你签名机端口号

### frp

1. [下载适合你机器的frp包](https://github.com/fatedier/frp/releases)
2. 只需要配置客户端, 解压下载后的tar包, 找到frpc.ini
3. 修改本地签名机地址和子域名
   ```
   [common]
    server_addr = frp.csiodev.com
    server_port = 7000
    log_file = ./frpc.log
    log_level = info
    log_max_days = 3
    token = 6g9MbClGFQfK0T34VH1k
    protocol = tcp
    
    
    [signer-noven]
    type = http
    
    #你自己签名机的ip
    local_ip = 127.0.0.1
    
    #你自己签名机的端口号
    local_port = 8080
    
    use_encryption = true
    use_compression = true
    
    #改成你自己的子域名
    #公网访问路径 http://[subdomain].frp.csiodev.com 即 http://noven.frp.csiodev.com
    subdomain = noven
   ```
4. 启动代理 ./frpc -c ./frpc.ini 即可
5. 最后, 将http://noven.frp.csiodev.com填入到Composor-Account即可

## FAQ

### 如何生成公私钥对？

使用子命令 `key generate`

```shell
./signer key generate
```

### 如果遇到 `undefined: pkcs11.Ctx` 之类的问题，是什么原因导致的？

原因：CGO_ENABLED 被设置成为了 0，需要将其更改为1。
以下为复现

```text
$ CGO_ENABLED=1 go build -o signer_test
# cs-evm-signer/pkg/hsm/aws
pkg/hsm/aws/cloudhsm.go:28:27: undefined: pkcs11.Ctx
pkg/hsm/aws/cloudhsm.go:29:27: undefined: pkcs11.SessionHandle
pkg/hsm/aws/cloudhsm.go:76:10: undefined: pkcs11.ObjectHandle
pkg/hsm/aws/cloudhsm.go:77:13: undefined: pkcs11.Attribute
pkg/hsm/aws/cloudhsm.go:77:30: undefined: pkcs11.NewAttribute
```

### 如何查看 CGO_ENABLED 的配置

在终端输入，如果看到 CGO_ENABLED="1"，则说明该配置已被打开。

```shell
go env
```

### 应该借助什么生成加密私钥的keystore 文件？

    请使用 CS-Wallet 

### 应该借助什么工具生成加密助记词文件？

    请使用 CS-Wallet

### 如何确认账户是否加载成功？

不同类型的账户下，都会打印出 account 下的第一个地址，因此如果有address信息展示即为加载成功。

## 签名机安全

1. 如果为云服务器，可添加 ssh 登陆白名单，签名机端口请求白名单。当账户使用完成后，删除相关配置，释放该云服务器。白名单地址为
   IPV4 格式。
2. 如果为 ngrok 做机器代理，则 IP 白名单地址为 IPV6 格式，即应该配置为 `::1`
3. 如果不想设置白名单，则删除whitelist里的配置即可；
4. 当有新增交易场景时，需要设置rule的规则；