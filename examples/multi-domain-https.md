# 多域名HTTPS配置示例

## 概述

Kun Gateway支持在同一个HTTPS端口(443)上监听多个域名，每个域名使用不同的SSL证书。这是通过SNI(Server Name Indication)技术实现的。

## 配置步骤

### 1. 准备证书文件

为每个域名准备对应的证书和私钥文件：

```bash
# 域名1: example.com
example.com.crt  # 证书文件
example.com.key  # 私钥文件

# 域名2: api.example.com  
api.example.com.crt  # 证书文件
api.example.com.key  # 私钥文件

# 域名3: admin.example.com
admin.example.com.crt  # 证书文件
admin.example.com.key  # 私钥文件
```

### 2. 通过Web界面上传证书

1. 访问控制面Web界面
2. 点击"证书管理"菜单
3. 为每个域名添加证书：

   **添加 example.com 证书：**
   - 域名：example.com
   - 证书文件：选择 example.com.crt
   - 私钥文件：选择 example.com.key

   **添加 api.example.com 证书：**
   - 域名：api.example.com
   - 证书文件：选择 api.example.com.crt
   - 私钥文件：选择 api.example.com.key

   **添加 admin.example.com 证书：**
   - 域名：admin.example.com
   - 证书文件：选择 admin.example.com.crt
   - 私钥文件：选择 admin.example.com.key

### 3. 配置路由规则

为每个域名配置对应的路由规则：

```json
[
  {
    "domain": "example.com",
    "path": "/",
    "service": "default/webapp",
    "port": 8080,
    "weight": 100
  },
  {
    "domain": "api.example.com",
    "path": "/",
    "service": "default/api-server",
    "port": 8080,
    "weight": 100
  },
  {
    "domain": "admin.example.com",
    "path": "/",
    "service": "default/admin-panel",
    "port": 8080,
    "weight": 100
  }
]
```

### 4. 测试多域名HTTPS

```bash
# 测试 example.com
curl -H "Host: example.com" https://your-gateway-ip/ -k

# 测试 api.example.com
curl -H "Host: api.example.com" https://your-gateway-ip/ -k

# 测试 admin.example.com
curl -H "Host: admin.example.com" https://your-gateway-ip/ -k
```

## 工作原理

### SNI (Server Name Indication)

1. **客户端连接**：客户端在TLS握手时发送SNI扩展，包含目标域名
2. **证书选择**：网关根据SNI中的域名选择对应的证书
3. **TLS握手**：使用选中的证书完成TLS握手
4. **请求转发**：根据域名和路径将请求转发到对应的后端服务

### 证书管理流程

```
客户端请求 → SNI域名 → 证书管理器 → 选择证书 → TLS握手 → 路由匹配 → 转发到后端
```

## 高级配置

### 通配符证书

支持通配符证书，一个证书可以覆盖多个子域名：

```bash
# 上传通配符证书
*.example.com.crt  # 证书文件
*.example.com.key  # 私钥文件
```

### 默认证书

如果没有找到对应域名的证书，系统会使用第一个上传的证书作为默认证书。

### 证书轮换

支持证书的动态更新，无需重启服务：

1. 上传新证书
2. 系统自动加载新证书
3. 旧证书自动失效

## 监控和日志

### 查看证书状态

```bash
# 查看所有证书
curl http://localhost:9090/api/v1/certificates

# 查看数据面证书
curl http://localhost:8080/api/v1/certificates
```

### 查看证书使用日志

```bash
# 查看数据面日志
kubectl logs -n kube-system -l app=kun-gateway,component=dataplane | grep "证书"
```

## 故障排查

### 常见问题

1. **证书不匹配**
   ```
   错误：未找到域名 xxx.com 的证书
   解决：检查证书是否正确上传，域名是否匹配
   ```

2. **SNI不支持**
   ```
   错误：客户端不支持SNI
   解决：使用支持SNI的客户端，或配置默认证书
   ```

3. **证书过期**
   ```
   错误：证书已过期
   解决：更新证书文件
   ```

### 调试命令

```bash
# 检查证书有效性
openssl x509 -in example.com.crt -text -noout

# 测试SNI连接
openssl s_client -connect your-gateway-ip:443 -servername example.com

# 查看TLS握手过程
openssl s_client -connect your-gateway-ip:443 -servername example.com -msg
```

## 性能优化

### 证书缓存

- 证书加载后缓存在内存中
- 支持证书的热更新
- 避免频繁的文件I/O操作

### 连接复用

- 支持HTTP/2连接复用
- 减少TLS握手开销
- 提高并发性能

## 安全建议

1. **证书安全**
   - 使用强加密算法（RSA 2048位或ECC）
   - 定期更新证书
   - 保护私钥文件

2. **访问控制**
   - 限制证书上传权限
   - 监控证书使用情况
   - 记录证书操作日志

3. **网络安全**
   - 使用防火墙限制访问
   - 启用HSTS头
   - 配置安全的TLS版本 