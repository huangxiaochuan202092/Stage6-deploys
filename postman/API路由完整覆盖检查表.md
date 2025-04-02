# API路由完整覆盖检查表

本文档用于检查所有API路由是否都有对应的Postman测试请求。

## 1. 页面路由

| 路由 | 请求方法 | 测试请求名称 | 状态 |
|------|---------|------------|------|
| `/` | GET | 获取主页 | ✅ |
| `/admin` | GET | 获取管理页面 | ✅ |
| `/user_manager` | GET | 获取用户管理页面 | ✅ |
| `/task_manager` | GET | 获取任务管理页面 | ✅ |
| `/blog_manager` | GET | 获取博客管理页面 | ✅ |
| `/wenjuan_manager` | GET | 获取问卷管理页面 | ✅ |

## 2. 用户认证接口

| 路由 | 请求方法 | 测试请求名称 | 状态 |
|------|---------|------------|------|
| `/user/send-code` | POST | 发送验证码 | ✅ |
| `/user/login-or-register` | POST | 登录或注册 | ✅ |

## 3. 用户管理接口

| 路由 | 请求方法 | 测试请求名称 | 状态 |
|------|---------|------------|------|
| `/user/` | GET | 获取所有用户 | ✅ |
| `/user/profile` | GET | 获取当前用户 | ✅ |
| `/user/:id` | GET | 获取指定用户 | ✅ |
| `/user/:id` | PUT | 管理员-更新用户 | ✅ |
| `/user/:id` | PUT | 普通用户-更新自己 | ✅ |
| `/user/:id` | DELETE | 管理员-删除用户 | ✅ |
| `/user/:id` | DELETE | 普通用户-删除自己 | ✅ |

## 4. 博客管理接口

| 路由 | 请求方法 | 测试请求名称 | 状态 |
|------|---------|------------|------|
| `/user/blog` | GET | 获取所有博客 | ✅ |
| `/user/blog/:id` | GET | 获取指定博客 | ✅ |
| `/user/blog` | POST | 创建博客 | ✅ |
| `/user/blog/:id` | PUT | 更新博客 | ✅ |
| `/user/blog/:id` | PUT | 更新自己的博客(普通用户) | ✅ |
| `/user/blog/:id` | DELETE | 删除博客 | ✅ |
| `/user/blog/:id` | DELETE | 删除自己的博客(普通用户) | ✅ |

## 5. 任务管理接口

| 路由 | 请求方法 | 测试请求名称 | 状态 |
|------|---------|------------|------|
| `/user/tasks` | GET | 获取所有任务 | ✅ |
| `/user/tasks/:id` | GET | 获取指定任务 | ✅ |
| `/user/tasks` | POST | 创建任务 | ✅ |
| `/user/tasks/:id` | PUT | 更新任务 | ✅ |
| `/user/tasks/:id` | PUT | 更新自己的任务(普通用户) | ✅ |
| `/user/tasks/:id` | DELETE | 删除任务 | ✅ |
| `/user/tasks/:id` | DELETE | 删除自己的任务(普通用户) | ✅ |

## 6. 问卷管理接口

### 6.1 问卷基本操作

| 路由 | 请求方法 | 测试请求名称 | 状态 |
|------|---------|------------|------|
| `/user/wenjuans` | GET | 获取所有问卷 | ✅ |
| `/user/wenjuans/search` | GET | 搜索问卷 | ✅ |
| `/user/wenjuans/:id` | GET | 获取指定问卷 | ✅ |
| `/user/wenjuans` | POST | 创建问卷 | ✅ |
| `/user/wenjuans` | POST | 创建问卷(普通用户) | ✅ |
| `/user/wenjuans/:id` | PUT | 更新问卷 | ✅ |
| `/user/wenjuans/:id` | PUT | 更新自己的问卷(普通用户) | ✅ |
| `/user/wenjuans/:id` | DELETE | 删除问卷 | ✅ |
| `/user/wenjuans/:id` | DELETE | 删除自己的问卷(普通用户) | ✅ |
| `/user/wenjuans/:id/pin` | POST | 置顶问卷 | ✅ |
| `/user/wenjuans/:id/unpin` | POST | 取消置顶问卷 | ✅ |

### 6.2 问卷分类管理

| 路由 | 请求方法 | 测试请求名称 | 状态 |
|------|---------|------------|------|
| `/user/wenjuans/categories` | GET | 获取所有分类 | ✅ |
| `/user/wenjuans/categories` | POST | 创建分类 | ✅ |
| `/user/wenjuans/categories/:id` | PUT | 更新分类 | ✅ |
| `/user/wenjuans/categories/:id` | DELETE | 删除分类 | ✅ |

### 6.3 问卷答案管理

| 路由 | 请求方法 | 测试请求名称 | 状态 |
|------|---------|------------|------|
| `/user/wenjuans/:id/answers` | GET | 获取问卷的所有答案 | ✅ |
| `/user/wenjuans/:id/answers/:answerId` | GET | 获取指定问卷答案 | ✅ |
| `/user/wenjuans/:id/answers` | POST | 提交问卷答案 | ✅ |
| `/user/wenjuans/:id/answers/:answerId` | PUT | 更新问卷答案 | ✅ |
| `/user/wenjuans/:id/answers/:answerId` | DELETE | 删除问卷答案 | ✅ |
| `/user/wenjuans/:id/answers/stats` | GET | 获取问卷答案统计 | ✅ |

## 使用指南

1. 在进行API测试时，使用此表格检查是否已经测试了所有路由
2. "状态"列使用✅表示已有测试请求，❌表示尚未测试
3. 如有新增路由，请及时更新此检查表和测试集合
