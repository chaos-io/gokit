# `middleware/session`

这个包只保留 session 相关 middleware，核心模型和校验能力来自 `github.com/chaos-io/gokit/session`。

## Transport middleware

- `HTTPMiddleware` 从 `Authorization: Bearer ...`、`X-Session-Token` 或 `session_token` cookie 读取 token。
- `UnaryServerInterceptor` / `StreamServerInterceptor` 从 gRPC metadata 的 `authorization` 或 `x-session-token` 读取 token。
- 读取 token 后调用 `session.Validator`，再用 `session.WithToken` / `session.WithSession` 写回 context。

## Endpoint middleware

- `ValidateMiddleware` 从 `session.TokenFromContext` 读取 token，调用 `session.Validator`，再写回校验后的 session。
- `AuthenticateMiddleware` 在校验 session 后调用 `UserResolver`，并把解析出的业务 user 写入 middleware 自己的 user context。

业务模型、签发、校验、store、token codec、`WithSession` 和 `WithToken` 都在 `github.com/chaos-io/gokit/session`。
