## 

---
1. [ ] 去掉自定义的 Transaction 直接使用 go-ethereum 的 struct 进行序列化和反序列化处理。

## 2024.1.9

---

1. [x] json 文件中去掉 version 和 rules 的字段。
2. [x] js 或者 json 文件，应该支持任意位置。
3. [x] Transaction 接口里面，判断所有的数值类型的字符串为 hex。
4. [x] 新增 MultiHSM 账户类型。
   
   yubihsm session 连接最大限制为 15，因此同一设备支持超过 15 个 ID 时，需要修改代码。 
    ```yaml
    type: MultiHSM
    provider: yubihsm-connector
    connector_url: localhost:12346
    private_key_ids: 1001-1003,1005
    ```

## 2024.1.3 ~ 2024.1.4

---

1. 增加动态 js 文件，动态验证逻辑。
    1. 配置文件使用动态加载，使用 --rule rule.js 指定加载配置文件。
    ```shell
    ./signer start --port 10000 --rule rule.json // 使用 json 作为规则匹配文件
    ./signer start --port 10001 --rule rule.js   // 使用 js   作为规则匹配文件
    ```
    2. 在不指定 rule 文件时，将会默认加载 rule.json 作为配置文件。如果配置文件不存在，则会退出服务。
    3. js 文件固定实现 check(message) 的方法，在 check 方法内实现的对应逻辑即可。
       ```js
       function check(message) {
        console.log("js message:", message)
        message = JSON.parse(message);
        const content = message.content;
     
        switch (message.type) {
        case "transaction":
        return content.transaction.from.toLowerCase() === content.transaction.to.toLowerCase() && (
        content.transaction.value === "0x0" || content.transaction.value === "0x00" ||
        content.transaction.value === "0");
        case "eip712":
        return content.data.domain.name === "Blend" && content.data.domain.version === "1.0" && content.data.domain.chainId === "0x1" && content.data.domain.verifyingContract === "0x29469395eaf6f95920e59f858042f0e28d98a20b";
        case "message":
        // 字符串中包含字串
        // return content.message.includes("12");
        // 字符串相等测试
        // return content.message === "1121212";
        // 测试 js 正则表达式
        return containsNumber(content.message);
        }
       }
     
       function containsNumber(str) {
         let regex = /\d/; // 这个正则表达式匹配任何数字
         return regex.test(str);
       }
       ```
2. 支持多个 hsm 设备。
    1. 动态指定连接 yubihsm 的 url，以便连接多个 hsm 设备。
   ```yaml
      type: HSM
      connector_url: localhost:12345 # 新增字段，当 provider == yubihsm-connector, 如果不填，默认为：localhost:3456
      provider: yubihsm-connector
      public_key_id:
      private_key_id: 1000
      use_last_pass: true
   ```   

## 2023.12.26

---

1. rule 中新增 value 的限制，限制 value 等于某个数值。比如 value == "0"
2. 删除 config.yaml 中的 whitelist 的配置，在 auth 中新增 ip 的字段，使用 auth 中的 ip 进行全局 router Ip 的判断。
3. 补充签名的日志。
4. 更新 readme.md 文件。

## 2023.6.27

---

- EvMnemonic 账户类型新增 `use_last_pass` 的配置，当用户配置该字段为 true 时，则自动使用上一个 key 的密码对配置信息进行解析。
  减少用户重复输入密码的次数，具体规则如下。
    - 在 EvMnemonic 中 keys 是有序的，即始终按照 keys 的 index 从小到大进行解析。
    - 当 `pass` 不为空 且 `use_last_pass` 为 true 时，优先使用 `pass` 的内容。
    - `use_last_pass` 未设置时，将默认为 false。

## V0.0.1

-----

- 支持 HSM
- 账户配置仅支持单账户模式