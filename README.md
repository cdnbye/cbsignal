
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
sudo ./admin.sh start
```

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
      "cluster_mode"
      "num_goroutine"
      "num_per_map"
  }
}
```

### Cluster Mode
RPC is used to communicate between all nodes. Specify master IP in `config_cluster.yaml`, then  start service:
```bash
sudo ./admin.sh start cluster config_cluster.yaml
``` 

## Related projects
* [cbsignal_node](https://github.com/cdnbye/cbsignal_node) - High performance CDNBye signaling service written in node.js

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
echo -17 > /proc/$(pidof cbsignal)/oom_adj     # 防止进程被OOM killer杀死
sudo ./admin.sh start
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
      "cluster_mode"
      "num_goroutine"
      "num_per_map"
  }
}
```

### 集群模式
节点之间采用RPC进行通信，首先在 `config_cluster.yaml` 中指定master节点的内网IP, 然后启动服务：
```bash
sudo ./admin.sh start cluster config_cluster.yaml
``` 

## 相关项目
* [cbsignal_node](https://github.com/cdnbye/cbsignal_node) - 基于node.js开发的高性能CDNBye信令服务




