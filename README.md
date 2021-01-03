### The signal server of [hlsjs-p2p-engine](https://github.com/cdnbye/hlsjs-p2p-engine), [ios-p2p-engine](https://github.com/cdnbye/ios-p2p-engine) and [android-p2p-engine](https://github.com/cdnbye/android-p2p-engine)

### build
Make sure that the golang development environment is installed
```bash
git clone https://github.com/cdnbye/cbsignal.git
cd cbsignal
make
```
or directly use compiled linux file [cbsignal](https://github.com/cdnbye/cbsignal/releases) .

### deploy
Upload binary file, admin.sh and config.yaml to server, create `cert` directory with `signaler.pem` and `signaler.key`, then start service:
```bash
chmod +x admin.sh
chmod +x cbsignal
./admin.sh start
```

### Set up Allow List
Domian Allow List allows you to limit the use of signaling service to your website and your streams, thus preventing unwanted use of your service on a third-party site. Set up your domain names in the config.yaml:
```yaml
allow_list:
  - "localhost"
  - "YOUE_DOMAIN1"
  - "YOUE_DOMAIN2"
```
If the accessing domain name doesn't match your whitelisted domain names, clients will not be able to connect to the server and will not receive or generate any peer traffic.

### test
```
import Hls from 'cdnbye';
var hlsjsConfig = {
    p2pConfig: {
        wsSignalerAddr: 'ws://YOUR_SIGNAL',
        // Other p2pConfig options provided by hlsjs-p2p-engine
    }
};
// Hls constructor is overriden by included bundle
var hls = new Hls(hlsjsConfig);
// Use `hls` just like the usual hls.js ...
```

### Get real-time information of signal service
```
GET /info
```
Response:
```
Status: 200

{
  "ret": 0,
  "data": {
      "version"
      "current_connections"
      "capacity"
      "utilization_rate"
      "compression_enabled"
  }
}
```

### Cluster Mode
RPC is used to communicate between all nodes. Specify master IP in `config_cluster.yaml`, then  start service:
```bash
./admin.sh start cluster config_cluster.yaml
``` 

### go语言版的 CDNBye 信令服务器，可用于Web、安卓、iOS SDK等所有CDNBye产品
#### 编译二进制文件
请先确保已安装golang开发环境
```bash
git clone https://github.com/cdnbye/cbsignal.git
cd cbsignal
make
```
或者直接使用已经编译好的linux可执行文件 [cbsignal](https://github.com/cdnbye/cbsignal/releases)

#### 部署
将编译生成的二进制文件、admin.sh和config.yaml上传至服务器，并在同级目录创建`cert`文件夹，将证书和秘钥文件分别改名为`signaler.pem`和`signaler.key`放入cert，之后启动服务：
```bash
chmod +x admin.sh
chmod +x cbsignal
echo -17 > /proc/$(pidof cbsignal)/oom_adj     # 防止进程被OOM killer杀死
./admin.sh start
```

### 设置域名白名单
域名白名单可以防止未经授权的网站使用你的信令服务，可以在 config.yaml 文件中进行设置：
```yaml
allow_list:
  - "localhost"
  - "YOUE_DOMAIN1"
  - "YOUE_DOMAIN2"
```

### 测试
```
import Hls from 'cdnbye';
var hlsjsConfig = {
    p2pConfig: {
        wsSignalerAddr: 'ws://YOUR_SIGNAL',
        // Other p2pConfig options provided by hlsjs-p2p-engine
    }
};
// Hls constructor is overriden by included bundle
var hls = new Hls(hlsjsConfig);
// Use `hls` just like the usual hls.js ...
```

### 通过API获取信令服务的实时信息
```
GET /info
```
响应:
```
Status: 200

{
  "ret": 0,
  "data": {
      "version"
      "current_connections"
      "capacity"
      "utilization_rate"
      "compression_enabled"
  }
}
```

### 集群模式
节点之间采用RPC进行通信，首先在 `config_cluster.yaml` 中指定master节点的内网IP, 然后启动服务：
```bash
./admin.sh start cluster config_cluster.yaml
``` 




