# MySQL 配置文件说明

## 配置文件说明

### 1. my.cnf (默认配置)
- **用途**: 通用配置，适用于大多数环境
- **特点**: 平衡了性能和稳定性
- **推荐**: 开发和测试环境使用

### 2. my.cnf.development (开发环境)
- **用途**: 开发环境专用配置
- **特点**: 
  - 较小的内存占用
  - 开启更多日志用于调试
  - 较宽松的安全设置
  - 更敏感的慢查询检测

### 3. my.cnf.production (生产环境)
- **用途**: 生产环境专用配置
- **特点**:
  - 优化的性能设置
  - 严格的安全配置
  - 关闭调试日志
  - 更大的内存和连接数限制

## 使用方法

### 切换配置文件

1. **使用开发环境配置**:
   ```bash
   cp mysql/my.cnf.development mysql/my.cnf
   docker-compose restart mysql
   ```

2. **使用生产环境配置**:
   ```bash
   cp mysql/my.cnf.production mysql/my.cnf
   docker-compose restart mysql
   ```

3. **使用默认配置**:
   ```bash
   # 当前已经是默认配置
   docker-compose restart mysql
   ```

### 配置参数说明

#### 重要参数调整指南

1. **innodb_buffer_pool_size**
   - 开发环境: 128M
   - 生产环境: 1G (建议为系统内存的 70-80%)

2. **max_connections**
   - 开发环境: 100
   - 生产环境: 500

3. **innodb_flush_log_at_trx_commit**
   - 开发环境: 2 (性能优先)
   - 生产环境: 1 (安全优先)

4. **slow_query_log**
   - 开发环境: 开启，long_query_time = 1
   - 生产环境: 开启，long_query_time = 2

## 监控和调优

### 查看当前配置
```sql
SHOW VARIABLES LIKE 'innodb_buffer_pool_size';
SHOW VARIABLES LIKE 'max_connections';
SHOW VARIABLES LIKE 'slow_query_log';
```

### 查看慢查询
```sql
SHOW VARIABLES LIKE 'slow_query_log_file';
-- 然后查看文件内容
```

### 性能监控
```sql
-- 查看连接数
SHOW STATUS LIKE 'Threads_connected';

-- 查看查询缓存命中率
SHOW STATUS LIKE 'Qcache%';

-- 查看 InnoDB 状态
SHOW ENGINE INNODB STATUS;
```

## 故障排除

### 常见问题

1. **MySQL 启动失败**
   - 检查配置文件语法
   - 检查日志文件权限
   - 检查端口是否被占用

2. **性能问题**
   - 调整 innodb_buffer_pool_size
   - 检查慢查询日志
   - 优化查询和索引

3. **连接问题**
   - 检查 max_connections 设置
   - 检查网络配置
   - 检查防火墙设置

### 日志文件位置
- 错误日志: `/var/log/mysql/error.log`
- 慢查询日志: `/var/log/mysql/slow.log`
- 一般查询日志: `/var/log/mysql/general.log`
- 二进制日志: `/var/log/mysql/mysql-bin.log.*`

## 安全建议

1. **生产环境必须**:
   - 设置强密码
   - 限制 root 用户远程访问
   - 定期备份数据
   - 监控日志文件

2. **网络安全**:
   - 使用防火墙限制访问
   - 考虑使用 SSL 连接
   - 定期更新 MySQL 版本

3. **数据安全**:
   - 定期备份
   - 测试恢复流程
   - 监控异常访问
