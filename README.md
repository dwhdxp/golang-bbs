# 介绍
Bluebell是基于Go语言开发的以帖子为核心功能的社区论坛系统

## 主要功能
- 用户系统
  - 用户注册：支持用户名和邮箱注册，注册成功后异步发送欢迎邮件
  - 用户登录：支持JWT身份认证系统
  - Token刷新：支持访问令牌的刷新机制 
- 社区管理：支持多分类社区的创建和管理
- 帖子功能：
  - 发布帖子
  - 帖子详情查看：支持分页查看
  - 帖子搜索
- 互动功能：
  - 帖子投票
  - 评论系统      
## 技术特点
- 后端框架：基于Gin框架开发
- 数据存储：
  - MySQL：存储用户、帖子、评论等基础数据
  - Redis：处理点赞计数、排行榜等高并发场景
- 消息队列：
  - RabbitMQ：处理异步任务（如发送注册欢迎邮件）
- 其他特性：
  - 雪花算法：生成分布式ID
  - JWT：处理用户认证
  - Swagger：API文档自动生成
  - Zap：日志管理
  - Validator：请求参数验证
  - Rate Limit：接口限流保护
## 项目结构
```
bluebell_backend/
├── controller/ # 处理HTTP请求，参数校验
├── logic/      # 业务逻辑层
├── dao/        # 数据访问层（MySQL/Redis）
├── models/     # 数据模型
├── pkg/        # 第三方包
├── routers/    # 路由配置
├── settings/   # 配置管理
└── middlewares/# 中间件
```
