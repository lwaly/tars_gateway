# tars_gateway
优先想法做一个与具体业务无关的腾讯 tars 框架网关，后面在考虑实现与具体框架无关的一个网关服务，组件化嵌入协议解析，服务调用等。

现支持客户端tcp转内部tars调用，http转内部非tars（实际http）调用。

tars 网关功能：  
rsa 加密：支持 2 中填充方式 RSA_PKCS1_PADDING，RSA_PKCS1_OAEP_PADDING  
异地登录下线处理：通过消息队列（使用的是 NATS Streaming）实现分布式  
鉴权：使用 jwt token 鉴权方式进行客户端鉴权，身份识别  
单机限速：包括限制连接数，总流量，单连接流量(流量可限制到server一级，连接数tcp下只能限制到app一级，http限制到server一级)  
负载均衡：支持轮询与hash  
支持服务开启关闭：tcp支持到app一级，http限制到server一级  
支持黑白名单：tcp支持到server一级，http限制到server一级  
支持配置热更新：程序监听SIGHUP信号，重读配置，可更新秘钥，黑白名单，服务开关闭，tcp的app，server的 id与name对应关系  

http转发说明：分割url path，第一部分为应用名，第二部分为服务名称与obj对象名。（例：https://github.com/lwaly/tars_gateway/user,lwaly为tars应用名，tars_gateway为tars服务名）

qq群：148083988（交流tars网关实现，提出意见，bug，实现）  
已上传客户端测试程序，地址：https://github.com/lwaly/tars_gateway_test_client