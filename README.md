### 简要介绍
im是一个即时通讯服务器，代码全部使用golang完成。主要功能  
1.支持tcp，websocket接入  
2.离线消息同步    
3.单用户多设备同时在线    
4.单聊，群聊，以及房间聊天场景  
5.支持服务水平扩展
数据库：MySQL+Redis  
通讯框架：GRPC  
长连接通讯协议：Protocol Buffers  
日志框架：Zap  

本文 摘自 gim im conn 部分，单独出来更容易理解