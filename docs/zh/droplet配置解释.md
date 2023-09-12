# droplet 的配置解释


### 常用配置项说明

这里我们对比较常用的配置项进行说明。

#### 链服务配置

- 包括：同步节点，消息节点，签名节点及授权节点。

```toml
[ChainService]
  Url =  "/ip4/192.168.200.21/tcp/45132"
  Token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiemwiLCJwZXJtIjoiYWRtaW4iLCJleHQiOiIifQ.3u-PInSUmX-8f6Z971M7JBCHYgFVQrvwUjJfFY03ouQ"
[Node]
  Url = "/ip4/192.168.200.21/tcp/3453"
  Token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiemwiLCJwZXJtIjoiYWRtaW4iLCJleHQiOiIifQ.3u-PInSUmX-8f6Z971M7JBCHYgFVQrvwUjJfFY03ouQ"
[Messager]
  Url = "/ip4/192.168.200.21/tcp/39812"
  Token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiemwiLCJwZXJtIjoiYWRtaW4iLCJleHQiOiIifQ.3u-PInSUmX-8f6Z971M7JBCHYgFVQrvwUjJfFY03ouQ"
[Signer]
  Type = "gateway"
  Url = "/ip4/192.168.200.21/tcp/45132"
  Token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiemwiLCJwZXJtIjoiYWRtaW4iLCJleHQiOiIifQ.3u-PInSUmX-8f6Z971M7JBCHYgFVQrvwUjJfFY03ouQ"
[AuthNode]
  Url = "http://192.168.200.21:8989"
  Token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiemwiLCJwZXJtIjoiYWRtaW4iLCJleHQiOiIifQ.3u-PInSUmX-8f6Z971M7JBCHYgFVQrvwUjJfFY03ouQ"
```

#### `API` 监听配置

`droplet` 默认监听端口为 `127.0.0.1:41235`, 为了支持不同网络的访问请求, 需要修改`API`的监听地址:

```yuml
[API]
ListenAddress = "/ip4/0.0.0.0/tcp/41235"
```

#### `PublishMsgPeriod` 配置

`droplet` 在收到 `droplet-client` 的订单时, 并不会马上就发布 `ClientDealProposal` 消息,会等待一定的时间, 由配置文件中的 `PublishMsgPeriod` 项来控制，在测试时可以将此项设置为较小值减少等待时间。下面的设置，将等待时间设置为10秒。

```yuml
PublishMsgPeriod = "10s"
```

#### `PieceStorage` 配置

目前 `droplet` 支持两种 `Piece` 数据的存储模式：
- 文件系统
- 对象存储

```yuml
[PieceStorage]
  [[PieceStorage.Fs]]
    Name = "local"
    Enable = true
    Path = "/mnt/pieces"
  [[PieceStorage.S3]]
    Name = "oss"
    Enable = false
    EndPoint = ""
    AccessKey = ""
    SecretKey = ""
    Token = ""
```

也可以通过命令配置，命令设置不需要重启进程。命令设置后会更新配置文件：

```bash
# 本地文件系统存储
./droplet piece-storage add-fs --path="/piece/storage/path" --name="local"

# 对象存储
./droplet piece-storage add-s3 --endpoint=<url> --name="oss"
```

#### `Miners` 配置

`droplet` 服务的矿工及每个矿工的参数，配置如下：

```
[[Miners]]
  Addr = "f01000"
  Account = "testuser01"
  
  ConsiderOnlineStorageDeals = true
  ConsiderOfflineStorageDeals = true
  ConsiderOnlineRetrievalDeals = true
  ConsiderOfflineRetrievalDeals = true
  ConsiderVerifiedStorageDeals = true
  ConsiderUnverifiedStorageDeals = true
  PieceCidBlocklist = []
  ExpectedSealDuration = "24h0m0s"
  MaxDealStartDelay = "336h0m0s"
  PublishMsgPeriod = "1h0m0s"
  MaxDealsPerPublishMsg = 8
  MaxProviderCollateralMultiplier = 2
  Filter = ""
  RetrievalFilter = ""
  TransferPath = ""
  MaxPublishDealsFee = "0 FIL"
  MaxMarketBalanceAddFee = "0 FIL"
  [CommonProviderConfig.RetrievalPricing]
    Strategy = "default"
    [CommonProviderConfig.RetrievalPricing.Default]
      VerifiedDealsFreeTransfer = true
    [CommonProviderConfig.RetrievalPricing.External]
      Path = ""
    [CommonProviderConfig.AddressConfig]
      DisableWorkerFallback = false
```

:::tip

如果有多个矿工，将上述配置拷贝一份即可。***如果矿工比较多，那配置文件会很长，后续会考虑优化***

:::


### 一份典型的 `droplet` 全量配置
```
# ****** 数据传输参数配置 ********
SimultaneousTransfersForStorage = 20
SimultaneousTransfersForStoragePerClient = 20
SimultaneousTransfersForRetrieval = 20

# ****** 全局基础参数配置 ********
[CommonProvider]
  ConsiderOnlineStorageDeals = true
  ConsiderOfflineStorageDeals = true
  ConsiderOnlineRetrievalDeals = true
  ConsiderOfflineRetrievalDeals = true
  ConsiderVerifiedStorageDeals = true
  ConsiderUnverifiedStorageDeals = true
  PieceCidBlocklist = []
  ExpectedSealDuration = "24h0m0s"
  MaxDealStartDelay = "336h0m0s"
  PublishMsgPeriod = "1h0m0s"
  MaxDealsPerPublishMsg = 8
  MaxProviderCollateralMultiplier = 2
  Filter = ""
  RetrievalFilter = ""
  TransferPath = ""
  MaxPublishDealsFee = "0 FIL"
  MaxMarketBalanceAddFee = "0 FIL"
  RetrievalPaymentAddress = ""
  DealPublishAddress = []
  [CommonProvider.RetrievalPricing]
    Strategy = "default"
    [CommonProvider.RetrievalPricing.Default]
      VerifiedDealsFreeTransfer = true
    [CommonProvider.RetrievalPricing.External]
      Path = ""
    

# 每个矿工可以有独立的基础参数，没有配置时使用全局配置，配置方式如下：

# ****** miner基础参数配置 ********
[[Miners]]
  Addr = "f01000"
  Account = "testuser01"
  
   ConsiderOnlineStorageDeals = true
   ConsiderOfflineStorageDeals = true
   ConsiderOnlineRetrievalDeals = true
   ConsiderOfflineRetrievalDeals = true
   ConsiderVerifiedStorageDeals = true
   ConsiderUnverifiedStorageDeals = true
   PieceCidBlocklist = []
   ExpectedSealDuration = "24h0m0s"
   MaxDealStartDelay = "336h0m0s"
   PublishMsgPeriod = "1h0m0s"
   MaxDealsPerPublishMsg = 8
   MaxProviderCollateralMultiplier = 2
   Filter = ""
   RetrievalFilter = ""
   TransferPath = "/mnt/transfer"
   MaxPublishDealsFee = "0 FIL"
   MaxMarketBalanceAddFee = "0 FIL"
   RetrievalPaymentAddress = ""
   DealPublishAddress = []
   [CommonProvider.RetrievalPricing]
     Strategy = "default"
     [CommonProvider.RetrievalPricing.Default]
       VerifiedDealsFreeTransfer = true
     [CommonProvider.RetrievalPricing.External]
       Path = ""

# ****** droplet 网络配置  ********
[API]
  ListenAddress = "/ip4/127.0.0.1/tcp/41235"
  RemoteListenAddress = ""
  Secret = "e647ee23cf95424162b974cd641b6a6479cbc7cb1209cc755f762c8248d50ba4"
  Timeout = "30s"

[Libp2p]
  ListenAddresses = ["/ip4/0.0.0.0/tcp/58418", "/ip6/::/tcp/0"]
  AnnounceAddresses = []
  NoAnnounceAddresses = []
  PrivateKey = "08011240d47934b6fccf8b79786335a55ccc04bdb9c92866cae2c0cea2fdefe0f2e7c18650dfbde5dd126c2a23a0d1c60686d3dedd064b67ba97c6161dd8007f0675e1a9"


# ****** venus 组件服务配置 ********
[ChainService]
  Url =  "/ip4/192.168.200.21/tcp/45132"
  Token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiemwiLCJwZXJtIjoiYWRtaW4iLCJleHQiOiIifQ.3u-PInSUmX-8f6Z971M7JBCHYgFVQrvwUjJfFY03ouQ"

[Node]
  Url = "/ip4/192.168.200.151/tcp/3453"
  Token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoidGVzdC11c2VyMDEiLCJwZXJtIjoic2lnbiIsImV4dCI6IiJ9.ETjNy3HMDS3ScZ3cax9xYb6AopNWYp4y71lZGCvYxMg"

[Messager]
  Url = "/ip4/127.0.0.1/tcp/39812"
  Token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoidGVzdC11c2VyMDEiLCJwZXJtIjoic2lnbiIsImV4dCI6IiJ9.ETjNy3HMDS3ScZ3cax9xYb6AopNWYp4y71lZGCvYxMg"

[Signer]
  Type = "gateway"
  Url = "/ip4/127.0.0.1/tcp/45132"
  Token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoidGVzdC11c2VyMDEiLCJwZXJtIjoic2lnbiIsImV4dCI6IiJ9.ETjNy3HMDS3ScZ3cax9xYb6AopNWYp4y71lZGCvYxMg"

[AuthNode]
  Url = "http://127.0.0.1:8989"
  Token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoidGVzdC11c2VyMDEiLCJwZXJtIjoic2lnbiIsImV4dCI6IiJ9.ETjNy3HMDS3ScZ3cax9xYb6AopNWYp4y71lZGCvYxMg"

#  ******** 数据库设置 ********
[Mysql]
ConnectionString = ""
MaxOpenConn = 100
MaxIdleConn = 100
ConnMaxLifeTime = "1m"
Debug = false

# ******** 扇区存储设置 ********
[PieceStorage]
S3 = []

[[PieceStorage.Fs]]
Name = "local"
ReadOnly = false
Path = "./.vscode/test"

# ******** 日志设置 ********
[Journal]
Path = "journal"

# ******** DAG存储设置 ********
[DAGStore]
RootDir = "/root/.droplet/dagstore"
MaxConcurrentIndex = 5
MaxConcurrentReadyFetches = 0
MaxConcurrencyStorageCalls = 100
GCInterval = "1m0s"
Transient = ""
Index = ""
UseTransient = false

# ******** 数据检索配置 ********
RetrievalPaymentAddress = ""

# ****** Metric 配置 ********
[Metrics]
  Enabled = false
  [Metrics.Exporter]
    Type = "prometheus"
    [Metrics.Exporter.Prometheus]
      RegistryType = "define"
      Namespace = ""
      EndPoint = "/ip4/0.0.0.0/tcp/4568"
      Path = "/debug/metrics"
      ReportingPeriod = "10s"
    [Metrics.Exporter.Graphite]
      Namespace = ""
      Host = "127.0.0.1"
      Port = 4568
      ReportingPeriod = "10s"
```

接下来，将这个配置分成基础参数，网络配置，Venus组件配置等多个部分进行讲解

## 全量配置说明
### 数据传输参数配置
```
# 存储订单的最大同时传输数目
# 整数类型 默认为：20
SimultaneousTransfersForStorage = 20

# 针对每一个客户端的存储订单最大同时传输数目
# 整数类型 默认为：20
SimultaneousTransfersForStoragePerClient = 20

# 获取数据最大同时传输数目
# 整数类型 默认为：20
SimultaneousTransfersForRetrieval = 20
```

### 基础参数配置

这部分的配置主要是决定了了 `droplet` 在进行工作时的偏好，满足定制化的需求，其中各项配置的作用如下：

``` 
# 决定是否接受线上存储订单
# 布尔值 默认为 true
ConsiderOnlineStorageDeals = true

# 决定是否接受线下存储订单
# 布尔值 默认为 true
ConsiderOfflineStorageDeals = true

# 决定是否接受线上数据获取订单
# 布尔值 默认为 true
ConsiderOnlineRetrievalDeals = true

# 决定是否接受线下数据获取订单
# 布尔值 默认为 true
ConsiderOfflineRetrievalDeals = true

# 决定是否接受经过验证的存储订单
# 布尔值 默认为 true
ConsiderVerifiedStorageDeals = true

# 决定是否接受未经过验证的存储订单
# 布尔值 默认为 true
ConsiderUnverifiedStorageDeals = true

# 订单数据黑名单
# 字符串数组 其中每一个字符串都是CID 默认为空
# CID在黑名单中的数据，不会被用于构建订单
PieceCidBlocklist = []

# 订单数据被封装完成的最大时间预期
# 时间字符串 默认为："24h0m0s"
# 时间字符串是由数字和时间单位组成的字符串，数字包括整数和小数，合法的单位包括 "ns", "us" (or "µs"), "ms", "s", "m", "h".
ExpectedSealDuration = "24h0m0s"

# 预期订单封装开始前等待时间
# 时间字符串 默认为："336h0m0s"
MaxDealStartDelay = "336h0m0s"

# 消息推送上链的周期
# 时间字符串 默认为："1h0m0s"
PublishMsgPeriod = "5m0s"

# 在一个消息推送周期内的最大订数量
# 整数类型 默认为8 
MaxDealsPerPublishMsg = 8

# 最大的存储供应商抵押乘法因子
# 整数类型 默认为：2
MaxProviderCollateralMultiplier = 2

# 通过外部执行器来筛选存储订单,是可执行的程序或脚本
Filter = ""

# 通过外部执行器来筛选检索订单,是可执行的程序或脚本
RetrievalFilter = ""

# 订单传输数据的存储位置
# 字符串类型 可选 为空值时默认使用`DROPLET_REPO`的路径
TransferPath = ""

# 发送订单消息的最大费用
# FIL类型 默认为："0 FIL"
# FIL类型字符串形式为 整数+" FIL"
MaxPublishDealsFee = "0 FIL"

# 发送增加抵押消息时花费的最大费用
# FIL类型 默认为："0 FIL"
MaxMarketBalanceAddFee = "0 FIL"

# 保留字段，当前配置无效
[RetrievalPricing]

# 使用的策略类型
# 字符串类型 可以选择"default"和"external"  默认为:"default"
# 前者使用内置的默认策略，后者使用外部提供的脚本自定义的策略
Strategy = "default"

[RetrievalPricing.Default]

# 对于经过认证的订单数据，是否定价为0
# 布尔值 默认为 "true"
# 只有Strategy = "default" 才会生效
VerifiedDealsFreeTransfer = true

[RetrievalPricing.External]
# 定义外部策略的脚本的路径
# 字符串类型 如果选择external策略时，必选
Path = ""

# 该设置为保留字段，当前无效
[AddressConfig]

# 是否降低使用woker地址发布消息的优先级，如果是，则只有在其他可选地址没有的情况下才会使用woker的地址发消息
# 布尔值 默认为 false
DisableWorkerFallback = false

[[AddressConfig.DealPublishControl]]

# 发布订单消息的地址
# 字符串类型 必选
Addr = ""

# 持有相应地址的账户
# 字符串类型 必选
Account =""
```

### droplet  网络配置

这部分的配置决定了 droplet 和外界交互的接口

#### [API]
droplet 对外提供服务的接口

```
[API]
# droplet 提供服务监听的地址
# 字符串类型，必选项，默认为:"/ip4/127.0.0.1/tcp/41235"
ListenAddress = "/ip4/127.0.0.1/tcp/41235"

# 保留字段
RemoteListenAddress = ""

# 密钥用于加密通信
# 字符串类型 可选项（没有则自动生成）
Secret = "878f9c1f88c6f68ee7be17e5f0848c9312897b5d22ff7d89ca386ed0a583da3c"

# 保留字段
Timeout = "30s"
```

#### [Libp2p]

droplet 在P2P网络中通信时使用的 通信地址
```
[Libp2p]
# 监听的网络地址
# 字符串数组 必选 默认为:["/ip4/0.0.0.0/tcp/58418", "/ip6/::/tcp/0"]
ListenAddresses = ["/ip4/0.0.0.0/tcp/58418", "/ip6/::/tcp/0"]

# 保留字段
AnnounceAddresses = []

# 保留字段
NoAnnounceAddresses = []

# 用于生成p2p节点的peerid
# 字符串 可选（没设置则自动生成）
PrivateKey = "08011240ae580daabbe087007d2b4db4e880af10d582215d2272669a94c49c854f36f99c35"
```

### venus 组件服务配置

当 `droplet` 接入venus组件使用时，需要配置相关组件的API。


#### [ChainService]
venus 链服务统一入口配置。
该配置项的 `Url` 和 `Token` 会成为后续配置项 `Node` , `Messager` 以及 `AuthNode` 的默认值
```toml
[ChainService]
  Url =  "/ip4/192.168.200.21/tcp/45132"
  Token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiemwiLCJwZXJtIjoiYWRtaW4iLCJleHQiOiIifQ.3u-PInSUmX-8f6Z971M7JBCHYgFVQrvwUjJfFY03ouQ"
```

#### [Node]
venus链同步节点接入配置
```
[Node]
# 链服务的入口
# 字符串类型 必选（也可以直接通过命令行的--node-url flag 进行配置）
Url = "/ip4/192.168.200.128/tcp/3453"

# venus 系列组件的鉴权token
# 字符串类型 必选（也可以直接通过命令行的 --auth-token flag 进行配置）
Token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiZm9yY2VuZXQtbnYxNiIsInBlcm0iOiJhZG1pbiIsImV4dCI6IiJ9.PuzEy1TlAjjNiSUu_tbHi2XPUritDLm9Xf5UW3MHRe8"

```


#### [Messager]

venus 消息服务接入配置

```
[Messager]
# 消息服务入口
# 字符串类型 可选（也可以直接通过命令行的 --messager-url flag 进行配置） 不接入链服务时可不填
Url = "/ip4/192.168.200.128/tcp/39812/"

# venus 系列组件的鉴权token
# 字符串类型 可选（也可以直接通过命令行的 --auth-token flag 进行配置） 不接入链服务时可不填
Token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiZm9yY2VuZXQtbnYxNiIsInBlcm0iOiJhZG1pbiIsImV4dCI6IiJ9.PuzEy1TlAjjNiSUu_tbHi2XPUritDLm9Xf5UW3MHRe8"
```


#### [Signer]

venus 提供签名服务的组件，它可以由两种类型：由venus-wallet直接提供的签名服务和由sophon-gateway提供的间接签名服务

```
[Signer]
# 签名服务组件的类型
# 字符串类型  枚举："gateway"，"wallet"，"lotusnode"
Type = "gateway"

# 签名服务入口
# 字符串类型 必选（也可以直接通过命令行的 --signer-url flag 进行配置）
Url = "/ip4/192.168.200.128/tcp/45132/"

# venus 系列组件的鉴权token
# 字符串类型 必选（也可以直接通过命令行的 --auth-token flag 进行配置）
Token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiZm9yY2VuZXQtbnYxNiIsInBlcm0iOiJhZG1pbiIsImV4dCI6IiJ9.PuzEy1TlAjjNiSUu_tbHi2XPUritDLm9Xf5UW3MHRe8"
```


#### [AuthNode]

venus 提供鉴权服务接入配置
```
[AuthNode]

# 鉴权服务入口
# 字符串类型 可选（也可以直接通过命令行的 --signer-url flag 进行配置） 不接入链服务时可不填
Url = "http://192.168.200.128:8989"

# venus 系列组件的鉴权token
# 字符串类型 可选（也可以直接通过命令行的 --auth-token flag 进行配置） 不接入链服务时可不填
Token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiZm9yY2VuZXQtbnYxNiIsInBlcm0iOiJhZG1pbiIsImV4dCI6IiJ9.PuzEy1TlAjjNiSUu_tbHi2XPUritDLm9Xf5UW3MHRe8"
```


### 矿工配置

预置矿工信息
```
[[Miners]]
# 矿工的地址
# 字符串类型 必选
Addr ="f01000"

# 账户名，可以随意设定
# 字符串类型 可选
Account = ""

# 基础参数，见上文
```

:::tip

基础参数在不配置时将会使用 `CommonProvider`, 如下:
```
[[Miners]]
  Addr = "f02472"
  Account = "litao"
```

基础参数一旦配置了一项,则所有项都必须自己配置,比如配置:
```
[[Miners]]
  Addr = "f02472"
  Account = "litao"
  TransferPath = "/mnt/transfer/2472"
```
这样的配置会导致基础参数中的其他配置项都去各自类型的零值,而不会用 `CommonProvider` 中的配置，
如 `f02472` 对应的 `ConsiderOnlineStorageDeals` 等于 `false`, 而并非是 `CommonProvider` 中的 true.

这一点需要特别注意,正确的配置方式:
```
[[Miners]]
  Addr = "f02472"
  Account = "litao"
  TransferPath = "/mnt/transfer/2472"
  ConsiderOnlineStorageDeals = true
  ConsiderOfflineStorageDeals = true
  ConsiderOnlineRetrievalDeals = true
  ConsiderOfflineRetrievalDeals = true
  ConsiderVerifiedStorageDeals = true
  ConsiderUnverifiedStorageDeals = true
  PieceCidBlocklist = []
  ExpectedSealDuration = "24h0m0s"
  MaxDealStartDelay = "336h0m0s"
  PublishMsgPeriod = "1m0s"
  MaxDealsPerPublishMsg = 8
  MaxProviderCollateralMultiplier = 2
  Filter = ""
  RetrievalFilter = ""
  MaxPublishDealsFee = "0 FIL"
  MaxMarketBalanceAddFee = "0 FIL"
  RetrievalPaymentAddress = ""
  [RetrievalPricing]
    Strategy = "default"
    [RetrievalPricing.Default]
      VerifiedDealsFreeTransfer = true
    [RetrievalPricing.External]
      Path = ""
```

这样不是很灵活,以后会考虑优化.

:::


### 数据库配置

droplet 运行过程中产生的数据的存储数据库的设置
目前支持BadgerDB和MySQLDB，默认使用BadgerDB

#### [Mysql]

MySQLDB的配置
```
[Mysql]

# 用于连接MySQL数据库的 connection string
# 字符串类型 如果要使用 MySQL 数据库的话，这是必选，否则使用默认的BadgerDB
ConnectionString = ""

# 打开连接的最大数量
# 整数类型 默认为100
MaxOpenConn = 100

# 空闲连接的最大数量
# 整数类型 默认为100
MaxIdleConn = 100

# 可复用连接的最大生命周期
# 时间字符串 默认为："1m"
# 时间字符串是由数字和时间单位组成的字符串，数字包括整数和小数，合法的单位包括 "ns", "us" (or "µs"), "ms", "s", "m", "h".
ConnMaxLifeTime = "1m"

# 是否输出数据库相关的调试信息
# 布尔值 默认false
Debug = false
```

###  扇区存储配置

配置 `droplet` 导入数据后生成的扇区的存储空间
支持使用两种类型的数据存储方式： 文件系统存储和对象存储

#### [[PieceStorage.Fs]]

配置本地文件系统作为扇区存储
对于大量数据的扇区，建议挂载和`Damocles`共用的文件系统进行配置 

```
[PieceStorage]
[[PieceStorage.Fs]]

# 存储空间的名称，它在 `droplet` 的所有的存储空间中，必须是唯一的
# 字符串类型 必选
Name = "local"

# 该存储空间是否可写（ read only false 即为可写）
# 布尔值 默认为 false
ReadOnly = false

# 该存储空间在本地文件系统中的路径
# 字符串类型 必选
Path = "/piecestorage/"

```

```
[PieceStorage]
[[PieceStorage.S3]]
# 存储空间的名称，它在 `droplet` 的所有的存储空间中，必须是唯一的
# 字符串类型 必选
Name = "s3"

# 该存储空间是否可写（ read only false 即为可写）
# 布尔值 默认为 false
ReadOnly = true

# 对象存储服务的入口
# 字符串类型 必选
# 支持单独的EndPoint（"oss-cn-shanghai.aliyuncs.com"）和完整的EndPoint Url（"http://oss-cn-shanghai.aliyuncs.com"）
EndPoint = "oss-cn-shanghai.aliyuncs.com"

# 对象存储服务的Bucket名称
# 字符串类型 必选
Bucket = "droplet"

# 指定在Bucket 中的子目录
# 字符串类型 可选
SubDir = "dir1/dir2"

# 访问对象存储服务的参数
# 字符串类型 其中AccessKey，SecretKey必选，token 可选
AccessKey = "LTAI5t6HiFgsqN6eVJ......"
SecretKey = "AlFNH9NakUsVjVRxMHaaYP7p......"
Token = ""

```


### 日志设置
配置 `droplet` 使用过程中，产生日志存储的位置

```
[Journal]

# 日志存储的位置
# 字符串类型 默认为："journal" (即`DROPLET_REPO`文件夹下面的journal文件夹)
Path = "journal"
```


### DAG存储设置

DAG 数据存储的配置

```
# 参考 github.com/filecoin-project/dagstore/dagstore.go
[DAGStore]

# DAG数据存储的根目录
# 字符串类型 默认为： "<DROPLET_REPO_PATH>/dagstore"
RootDir = "/root/.droplet/dagstore"

# 可以同时进行索引作业的最大数量
# 整数类型 默认为5 0表示不限制
MaxConcurrentIndex = 5

# 可以同时被抓取的最大未封装订单的数量
# 整数类型 默认为0 0表示不限制
MaxConcurrentReadyFetches = 0

# 可以被同时调用的存储API的最大数量
# 整数类型 默认为100
MaxConcurrencyStorageCalls = 100

# DAG 数据进行垃圾回收的时间间隔
# 时间字符串 默认为："1m0s"
# 时间字符串是由数字和时间单位组成的字符串，数字包括整数和小数，合法的单位包括 "ns", "us" (or "µs"), "ms", "s", "m", "h".
GCInterval = "1m0s"

# 临时文件的存储路径
# 字符串类型 可选 不设置则使用RooDir目录下的'transients'文件夹
Transient = ""

# 存储扇区索引数据的路径
# 字符串类型 可选 不设置则使用RooDir目录下的'index'文件夹
Index = ""

# 不使用本地缓存，直接读取数据源
# 布尔类型 默认为 false
UseTransient = false
```

### 数据检索

获取订单中存储的扇区数据时的相关配置

### [RetrievalPaymentAddress]
获取订单扇区数据时，使用的收款地址
```
RetrievalPaymentAddress = ""
```

### Metric 配置

配置 Metric 相关的参数


```toml
[Metrics]

# 是否启用 Metric
# 布尔值 默认为 false
Enabled = false

# Metric 导出设置
[Metrics.Exporter]

# Metric 导出的类型
# 字符串类型 可选值为 "prometheus" 和 "graphite" 默认为 "prometheus"
Type = "prometheus"

# Prometheus 导出设置
[Metrics.Exporter.Prometheus]

# 注册器的类型
# 字符串类型 可选值为 "define" 和 "default" 默认为 "define"
# define: 空白全新的注册器; default:Prometheus 提供的默认注册器
RegistryType = "define"

# 命名空间
# 字符串类型 默认为 ""
Namespace = ""

# 监听地址
# 字符串类型 默认为 "/ip4/0.0.0.0/tcp/4568"
EndPoint = "/ip4/0.0.0.0/tcp/4568"

# Metrics 指标的访问路径
# 字符串类型 默认为 "/debug/metrics"
Path = "/debug/metrics"

# Metric 指标聚合的周期
# 时间字符串 默认为 "10s"
ReportingPeriod = "10s"


# Graphite 导出设置
[Metrics.Exporter.Graphite]

# 命名空间
# 字符串类型 默认为 ""
Namespace = ""

# 监听地址
# 字符串类型 默认为 "127.0.0.1"
Host = "127.0.0.1"

# 监听端口
# 整数类型 默认为 4568
Port = 4568

# Metric 指标聚合的周期
# 时间字符串 默认为 "10s"
ReportingPeriod = "10s"
```
