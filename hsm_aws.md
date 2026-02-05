## Aws Cloud HSM 文档

### 环境搭建
1. 需要安装的内容有
   1. cloudhsm-client [cloudhsm 交互执行客户端]
   2. GCC/C++ 7.3.1-5
   3. OpenSSL 1.1.1t
   4. CMake > 3
   5. Botan > 2.6.0
   6. cloudhsm-client-pkcs11 [cloudhsm pkcs11 库] 
   7. Go1.19

2. 安装步骤
```shell
wget https://s3.amazonaws.com/cloudhsmv2-software/CloudHsmClient/EL7/cloudhsm-client-latest.el7.x86_64.rpm
sudo yum install ./cloudhsm-client-latest.el7.x86_64.rpm

# botan >= 2.6
1. https://github.com/randombit/botan
2. https://cloud.tencent.com/developer/article/1542655

# openssl >= 1.1.1
https://gist.github.com/Bill-tran/5e2ab062a9028bf693c934146249e68c

# gcc 版本升级
https://azdigi.com/blog/en/webserver-panel-en/directadmin-en/fix-the-error-cxxabi-1-3-9-not-found-on-centos-7-running-directadmin/

#  cloudhsm-client-pkcs11 有版本要求，最好安装 3.4.4 版本的，因为 5 版本的会和 cloudhsm client 冲突
wget https://s3.amazonaws.com/cloudhsmv2-software/CloudHsmClient/EL6/cloudhsm-client-pkcs11-3.4.4-1.el6.x86_64.rpm
# 查看 pkcs11.so 的依赖情况。
ldd /opt/cloudhsm/lib/libcloudhsm_pkcs11.so
```


3. cloudhsm-client 安装成功后，可以执行相关命令
```shell
# 登陆 cloudhsm 管理交互模式
/opt/cloudhsm/bin/cloudhsm_mgmt_util /opt/cloudhsm/etc/cloudhsm_mgmt_util.cfg

# 登陆 cloudhsm hsm 管理模式，在该模式下，操作单个公钥或者私钥对象
/opt/cloudhsm/bin/key_mgmt_util
# 登陆命令，登陆的用户身份必须为 CU。
loginHSM -u CU -s example_user -hpswd <password>
```
更多的内容：https://docs.aws.amazon.com/zh_cn/cloudhsm/latest/userguide/key_mgmt_util-getting-started.html


### HSM 添加新的身份
使用 cloudhsm_client 安装好的工具进行，流程如下：
```shell
/opt/cloudhsm/bin/cloudhsm_mgmt_util /opt/cloudhsm/etc/cloudhsm_mgmt_util.cfg
loginHSM CO admin -hpswd # 使用 admin 的账户登陆，管理用户，否则权限不够。
listUsers # 列出集群上的所有用户
createUser CU user_name user_password # 创建 CU 账户管理公私钥对。
```


### HSM 中私钥生成步骤
```shell
/opt/cloudhsm/bin/key_mgmt_util
loginHSM -u CU -s <your_username> -hpswd <password> # 测试密码：<your_password>

genECCKeyPair -i 16 -l ether_test
# 16 为 aws 的 secp256k1 曲线 id
#   Cfm3GenerateKeyPair:  public key handle: 262159    private key handle: 262158
# 公钥 ID 为 262159，私钥 ID 为 262158
```

### 如何查看 HSM 私钥生成后的地址
1. 在 conf/config.yaml 文件中配置 hsm 的相关参数。配置完成后，启动签名机
   即可打印出对应的地址。
2. 具体查看 conf/config.yaml.example 文件。
```text
account type: [HSM], index: [3], address: [0xf6994200339FB759de34dfC26052295Dfb922EB0]
```

### 如何快速测试 pkcs#11 可以正常交互
1. 使用 aws 官方给的 C 版本测试用例进行测试。具体步骤如下。
   ```shell
   git clone https://github.com/ThreeAndTwo/aws-cloudhsm-pkcs11-examples.git
   cd aws-cloudhsm-pkcs11-examples/
   mkdir build && cd build
   cmake .. -DHSM_USER=<your_username> -DHSM_PASSWORD=<your_password>
   sudo make
   # #正常打印出 hash 证明整个环境已经准备就行
   src/sign/sign --pin <your_username>:<your_password> 
   ```
2. 使用签名机代码配置 aws 的 HSM 测试。

### 注意事项
1. 注意版本工具的版本要求


### 参考资料
- [aws cloudhsm pkcs11](https://github.com/aws-samples/aws-cloudhsm-pkcs11-examples)