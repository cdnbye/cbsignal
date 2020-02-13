### The signal server of [hlsjs-p2p-engine](https://github.com/cdnbye/hlsjs-p2p-engine), [ios-p2p-engine](https://github.com/cdnbye/ios-p2p-engine) and [android-p2p-engine](https://github.com/cdnbye/android-p2p-engine)

### build
```bash
git clone https://github.com/cdnbye/cbsignal.git
cd cbsignal
make
```

### deploy
Upload binary file, admin.sh and config.yaml to server, create `cert` directory with `signaler.pem` and `signaler.key`, then start service:
```bash
ulimit -n 1000000
chmod +x admin.sh
./admin.sh start
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

### go语言版的 CDNBye 信令服务器，可用于Web、安卓、iOS SDK等所有CDNBye产品
#### 编译二进制文件
```bash
git clone https://github.com/cdnbye/cbsignal.git
cd cbsignal
make
```

#### 部署
将编译生成的二进制文件、admin.sh和config.yaml上传至服务器，并在同级目录创建`cert`文件夹，将证书和秘钥文件分别改名为`signaler.pem`和`signaler.key`放入cert，之后启动服务：
```bash
ulimit -n 1000000
chmod +x admin.sh
./admin.sh start
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



