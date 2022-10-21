# qqbot

这是一个自用的 QQ 机器人，目前主要实现的功能就是对 B 站小程序和 B 站链接进行解析。  
也可以添加其它功能，要添加新功能应该还是比较方便的（大概）。

需要使用[go-cqhttp](https://github.com/Mrs4s/go-cqhttp)进行 QQ 登录以及消息收发。

## 使用

- 该程序只支持通过反向 websocket（ws-reverse）与 go-cqhttp 通信，暂不支持也不打算支持其它协议。

## TODO

- 把端口、日志等级之类的设置项抽离到配置文件中
- 鬼知道要加什么功能
