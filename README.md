# Overview

[![Go Report Card](https://goreportcard.com/badge/github.com/eudore/website)](https://goreportcard.com/report/github.com/eudore/website)
[![GoDoc](https://godoc.org/github.com/eudore/website?status.svg)](https://godoc.org/github.com/eudore/website)

website目标是一个自用的多功能平台，作为[eudore框架](https://github.com/eudore/eudore)的damo项目，应用和框架相互促进，共同发展。

[在线demo](https://www.eudore.cn/auth/),用户密码均为guest.

# 功能

待实现：

- 体验 - seo优化
- 管理 - 限流
- 管理 - SingleFlight
- 管理 - 黑名单
- 开发 - 自动测试
- 功能 - 第三方Oauth2对接
- 功能 - gravatar支持
- 开发 - 启用配置化
- 功能 - 角色管理

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
- 体验 - 静态资源自动push
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