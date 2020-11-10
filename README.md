# gracefulrpc
基于golang的rpc服务，使用tcp传输协议，支持gob, json, protobuf编解码器，同时支持零停机重启。

### 使用
go get -u github.com/phprao/gracefulrpc

### 零停机部署
信号：
  - SIGHUP  平滑重启
  - SIGTERM  平滑停止
  - SIGINT  立即停止
  
平滑停止和平滑重启默认延迟一分钟，可以设置。同时你还可以Writer对象来满足日志记录的需要。

### DEMO
rpc_gob  
rpc_json  
rpc_protobuf  