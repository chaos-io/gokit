# Tracing 使用说明

`tracing` 为基于 Go kit 的服务提供 OpenTelemetry 链路追踪辅助能力。它的默认策略是“追踪失败不影响业务链路”：未启用 tracing 或初始化失败时，会返回 noop tracer 和可安全调用的 shutdown 函数。

## 配置

`New` 会通过 `chaos/config` 读取 `tracing` 配置段：

```yaml
tracing:
  enable: false
  endpoint: localhost:4318
  sampleRatio: 1
  environment: dev
  serviceVersion: ''
  serviceInstanceID: ''
```

字段说明：

- `enable`：为 `true` 时启用 OTLP 导出；为 `false` 时使用 noop tracer。
- `endpoint`：OTLP HTTP endpoint，可写成 `host:port` 或完整 URL。空值会回退到 `localhost:4318`。
- `sampleRatio`：root trace 采样比例。`1` 表示全采样，`0.01` 表示约 1% 采样，`<= 0` 表示主动丢弃新 root trace。旧配置文件缺失该字段或值为 `0` 时，默认按 `1` 处理，保持兼容。
- `environment`：部署环境，会导出为 `deployment.environment`。
- `serviceVersion`：服务版本，会导出为 `service.version`。
- `serviceInstanceID`：实例 ID、Pod 名或主机 ID，会导出为 `service.instance.id`。

## 初始化

普通服务优先使用 `New`：

```go
tracer, shutdown, err := tracing.New("example.user.UserService")
if err != nil {
	logs.Warnw("failed to set tracer", "error", err)
}
defer func() {
	_ = shutdown(context.Background())
}()
```

测试或嵌入式场景需要显式传配置时，使用 `NewWith`：

```go
tracer, shutdown, err := tracing.NewWith(ctx, "example.user.UserService", &tracing.Config{
	Enable:            true,
	Endpoint:          "otel-collector:4318",
	SampleRatio:       0.05,
	Environment:       "prod",
	ServiceVersion:    "v1.2.3",
	ServiceInstanceID: "user-server-7d8f",
})
```

`NewTracer` 和 `NewTraceProvider` 是更底层的辅助函数，适合只需要指定 OTLP endpoint 的调用方。应用代码通常优先使用 `New` 或 `NewWith`。

## 上下文传播

HTTP 入站请求应在 endpoint 执行前提取 W3C TraceContext 和 Baggage：

```go
serverOptions := []httptransport.ServerOption{
	httptransport.ServerBefore(tracing.HTTPToContext),
}
```

HTTP 出站请求可使用 `InjectHTTPHeader` 注入传播头：

```go
tracing.InjectHTTPHeader(ctx, req.Header)
```

gRPC 入站请求应在 endpoint 执行前从 metadata 提取 trace 上下文：

```go
serverOptions := []grpctransport.ServerOption{
	grpctransport.ServerBefore(tracing.GRPCToContext),
}
```

当前 propagator 使用 W3C TraceContext 和 Baggage。如果运行环境仍依赖 B3 或 Jaeger header，需要先补充兼容 propagator，再期望跨技术栈 trace 连续。

## Endpoint Span

服务端和客户端 endpoint 可分别使用 `TraceServer` 和 `TraceClient` 包装：

```go
endpoint := tracing.TraceServer(tracer, "GetUser")(endpoint)
clientEndpoint := tracing.TraceClient(tracer, "GetUser")(clientEndpoint)
```

`TraceServer` 创建 server span，`TraceClient` 创建 client span。两者都会记录返回的 error 和 `endpoint.Failer` 暴露的业务错误。若业务错误只希望作为属性记录、不希望把 span 标记为失败，可使用 `WithIgnoreBusinessError`。

operation name 应保持低基数，例如 RPC 方法名。不要把用户 ID、资源 ID、原始 path、query string 或请求体字段放进 span name 或高频属性。

## 生产建议

- span 优先发送到本地 sidecar、daemonset 或集群内 OpenTelemetry Collector，不建议业务服务直接连接远端 tracing backend。
- 高流量服务不要长期全采样。可从 `sampleRatio: 0.01` 或 `0.05` 开始，事故排查或灰度验证时再临时调高。
- `environment`、`serviceVersion`、`serviceInstanceID` 建议由部署系统注入，方便按环境、版本、实例过滤 trace。
- tracing 应按 best-effort 处理。导出失败、collector 不可用或 shutdown flush 失败应记录日志并监控，但不能导致业务请求失败。
- 不要在 span 属性中写入敏感信息或高基数字段。优先记录稳定的方法名、状态分类、重试次数和依赖名称。
