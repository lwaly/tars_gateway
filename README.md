# tars_gateway
优先想法做一个与具体业务无关的腾讯 tars 框架网关，后面在考虑实现与具体框架无关的一个网关服务，组件化嵌入协议解析，服务调用等。

tars 网关功能：
rsa 加密：支持 2 中填充方式 RSA_PKCS1_PADDING，RSA_PKCS1_OAEP_PADDING
异地登录下线处理：通过消息队列（使用的是 NATS Streaming）实现分布式
鉴权：使用 jwt token 鉴权方式进行客户端鉴权，身份识别
单机限速：包括限制连接数，总流量，单连接流量

qq群：148083988（交流tars网关实现，提出意见，bug，实现）
已上传客户端测试程序，地址：https://github.com/lwaly/tars_gateway_test_client