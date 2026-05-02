# `middleware/session`

这个包已经按不兼容旧版的方式重建为一套 **stateful session** 组件。

核心原则只有三条：

1. `token` 只负责承载最小 claims，并通过 HMAC 签名保证不可伪造。
2. `session` 的真实有效性以 `Store` 中的状态为准，而不是只信 token。
3. 鉴权中间件只做两层事：校验 session、解析并注入 user。

## 当前实现

### 1. 核心模型

- `Subject`
  - 登录主体，当前包含 `UserID` 和 `AppID`
- `Session`
  - 服务端会话状态，包含 `ID`、`Subject`、`IssuedAt`、`ExpiresAt`、`RevokedAt`
- `IssuedSession`
  - 签发结果，包含 `Token` 和最终 `Session`
- `TokenClaims`
  - token 中承载的最小字段，只包含校验所需的 session 标识和时间信息

### 2. 接口边界

- `Issuer`
  - `Issue(ctx, subject)`：签发 session
- `Validator`
  - `Validate(ctx, token)`：校验 token 并返回有效 session
- `Revoker`
  - `Revoke(ctx, sessionID)`：吊销 session
- `Store`
  - `Save(ctx, session)`：持久化 session
  - `Find(ctx, sessionID)`：按 ID 读取 session
- `RevocationStore`
  - `Revoke(ctx, sessionID, revokedAt)`：可选的吊销能力
- `TokenCodec`
  - `Encode/Decode`：负责 token 编解码，不关心 store

这几个接口是有意拆开的：

- token 签名和 session 状态存储解耦
- 签发、校验、吊销的职责分离
- Redis / MySQL / PostgreSQL / 内存实现都可以直接接到 `Store`

### 3. 默认实现

#### `Manager`

`Manager` 组合了 `Store` 和 `TokenCodec`，同时实现 `Issuer`、`Validator`、`Revoker`。

签发流程：

1. 生成随机 session ID
2. 构造 `Session`
3. 将 `Session` 映射为最小 `TokenClaims`
4. 用 `TokenCodec` 编码 token
5. 将 session 落到 `Store`

校验流程：

1. `TokenCodec.Decode` 校验 token 结构和签名
2. 从 `Store` 按 `SessionID` 读取 session
3. 检查是否存在、是否过期、是否已吊销
4. 比对 token claims 和 store 中的 session 状态是否一致

这意味着：

- 只拿到旧 token 但 store 中 session 已被吊销时，校验会失败
- store 中 session 状态被改写后，旧 token 也会失效
- 单端登录可以通过 store 语义实现，而不是塞进 token 逻辑里

#### `HMACCodec`

默认 token 编码器是 `HMACCodec`：

- 格式是三段式：`base64url(header).base64url(claims).base64url(signature)`
- header 内含 `alg`、`typ`、`kid`
- 签名算法固定为 `HS256`

#### `StaticKeyring`

默认密钥管理是 `StaticKeyring`：

- 必须显式传入 `Key{ID, Secret}`
- secret 最少 32 字节
- 支持 `active + fallback keys`
- token 校验时按 `kid` 找 key，因此支持平滑轮换

#### `MemoryStore`

包内提供了一个 `MemoryStore` 用于测试和本地场景：

- 实现了 `Store`
- 实现了 `RevocationStore`
- 支持 `WithSingleSessionPerUser()`

启用 `WithSingleSessionPerUser()` 后：

- 新 session 保存时，会把同一 `UserID + AppID` 的旧活跃 session 标记为 `RevokedAt`
- 这样旧 token 再校验时会直接失败
- 历史 session 仍然保留在 `sessions` 存档中，但不会继续留在活跃索引里

## 中间件能力

### `ValidateMiddleware`

作用：

1. 从 context 中读取 token
2. 调用 `Validator.Validate`
3. 把校验后的 `Session` 注入 context

### `AuthenticateMiddleware`

作用：

1. 先执行 `ValidateMiddleware`
2. 调用 `UserResolver` 用 session 解析真实用户
3. 校验 `resolved user ID` 与 `session.Subject.UserID` 一致
4. 把 user 注入 context

这样服务侧不需要重复写：

- 读 token
- 校验 session
- 查用户
- 校验用户和 session 是否匹配
- 往 context 回填 user

## Context 工具

- `WithToken / TokenFromContext`
- `WithSession / SessionFromContext`
- `WithUser / AnyUserFromContext / UserFromContext[T]`

推荐做法是：

- transport 层把 header/cookie 里的 token 提前写入 context
- endpoint middleware 统一做 session 和 user 注入
- endpoint / service 层只从 context 取 `Session` 或业务用户

## 使用示例

```go
keyring, err := session.NewStaticKeyring(
	session.Key{
		ID:     "2026-05",
		Secret: []byte("0123456789abcdef0123456789abcdef"),
	},
	session.Key{
		ID:     "2026-04",
		Secret: []byte("fedcba9876543210fedcba9876543210"),
	},
)
if err != nil {
	return err
}

codec, err := session.NewHMACCodec(keyring)
if err != nil {
	return err
}

store := session.NewMemoryStore(session.WithSingleSessionPerUser())

manager, err := session.NewManager(store, codec)
if err != nil {
	return err
}

issued, err := manager.Issue(ctx, session.Subject{
	UserID: "user-1",
	AppID:  1001,
})
if err != nil {
	return err
}

validated, err := manager.Validate(ctx, issued.Token)
if err != nil {
	return err
}

_ = validated
```

## 生产接入建议

- 不要把 secret 写死在代码里。应从 KMS、配置中心或环境变量注入 `Keyring`。
- 生产环境应实现自己的 `Store`，通常落到 Redis 或数据库。
- 如果要做单端登录，推荐在 `Store.Save` 这一层定义“同主体仅保留一个有效 session”的语义。
- 如果要做主动吊销，`Store` 还应实现 `RevocationStore`。
- transport 层只负责把 token 提前放进 context，不要在每个 endpoint 里重复写 header/cookie 解析。

## 这次重构解决了什么

- 去掉了库内置默认 secret 的高风险设计
- 把 token 签名和 session 状态校验拆开
- 把签发、校验、吊销能力拆成清晰接口
- 直接提供了通用的 session 校验和 user 注入中间件
- 默认实现已经支持密钥轮换、吊销、单端登录语义
