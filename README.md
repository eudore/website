# Overview

website目标是一个自用的多功能平台，作为[eudore框架](https://github.com/eudore/eudore)的damo项目，应用和框架相互促进，共同发展。

[在线demo](https://www.eudore.cn),用户密码均为guest.

# 功能

待实现：

- 功能 - seo优化
- 功能 - 角色管理
- 功能 - 堡垒机录像回放

2020年9月20日更新：
- **全面重构**
- 功能 - 添加堡垒机功能，支持终端和阅览器双协议登录
- 功能 - 堡垒机ssh协议登录允许使用证书
- 功能 - 堡垒机录像存储
- 功能 - 堡垒机使用sshd或agent进行控制
- 管理 - 限流
- 管理 - SingleFlight
- 管理 - 黑名单
- 功能 - gravatar支持

2019年11月10日完成:
- 功能 - 第三方Oauth2对接
- 功能 - 简单聊天功能
- 功能 - 简单文档功能
- 功能 - 优化输出访问日志格式
- 鉴权 - PBAC策略支持browser
- 开发 - 启用配置化

2019年9月29日完成：
- 部署 - docker部署
- 部署 - Dockerfile
- 开发 - 统一消息弹出
- 开发 - DB处理封装
- 开发 - 编译自动重启
- 开发 - 启动命名支持
- 管理 - 熔断器及后台
- 体验 - Gzip启用
- 体验 - 静态资源合并
- 体验 - web前端I18n实现
- 功能 - 静态文件服务
- 功能 - 服务状态显示
- 功能 - 登录验证码
- 功能 - 用户登录
- 功能 - 用户权限管理
- 功能 - 权限管理
- 功能 - 策略管理
- 安全 - 防止sql注入
- 安全 - CSP启用
- 安全 - SRI自动计算
- 安全 - 禁用Cookie防止csrf
- 鉴权 - ACl
- 鉴权 - Rbac
- 鉴权 - Pbac
- 认证 - ak认证
- 认证 - Token认证
- 认证 - Bearer认证

# docker部署

eudore/website使用docker部署，分为使用git或者go拉包两种方式，主要区别在于命令中文件位置不同。

使用git获取包：

```bash
# git获取website包
git clone https://github.com/eudore/website.git
cd website
# 创建初始化sql
bash generatesql.sh > init.sql
# 创建website镜像
docker build -t eudore/website .
# 运行容器
docker run -d --name websitedb -e POSTGRES_USER=website -e POSTGRES_PASSWORD=website -e POSTGRES_DB=website -v $(pwd)/init.sql:/docker-entrypoint-initdb.d/init.sql library/postgres
docker run -d -p 8080:80 --link websitedb eudore/website
```

使用goget获取包：

```bash
# go get获取website包
GO111MODULE=off go get github.com/eudore/website
# 创建初始化sql
bash $GOPATH/src/github.com/eudore/website/generatesql.sh > init.sql
# 创建website镜像
docker build -t eudore/website -f $GOPATH/src/github.com/eudore/website/Dockerfile .
# 运行容器
docker run -d --name websitedb -e POSTGRES_USER=website -e POSTGRES_PASSWORD=website -e POSTGRES_DB=website -v $(pwd)/init.sql:/docker-entrypoint-initdb.d/init.sql library/postgres
docker run -d -p 8080:80 --link websitedb eudore/website
```

然后访问[http://localhost:8080/auth/](http://localhost:8080/auth/)