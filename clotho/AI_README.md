AI Prompt – Clotho

You are assisting with **Clotho**, the API orchestration layer in the Fly monorepo.  
Clotho exposes **HTTP/REST APIs** externally and calls **internal domain services via gRPC**.  
It does not implement business logic – it only orchestrates Custos, Orders, Billing, etc.

---

## 📂 Directory Responsibilities

### Root
- `README.md` → Human-facing doc (overview, usage, principles).  
- `CLOTHO_AI_PROMPT.md` → AI context for generation rules.  
- `LICENSE`, `Makefile`, etc. for project setup.  

### `cmd/`
- `clotho/` → CLI entrypoint using cobra (subcommands: `serve`, `version`).  

### `configs/`
- YAML/ENV configuration files (loaded via Mora config loader).  

### `internal/application/usecase/`
- Contains orchestration logic.  
- Example: `user_proxy.go` calls Custos gRPC for user data.  
- Example: `order_proxy.go` calls Orders gRPC.  

### `internal/infrastructure/client/`
- gRPC clients to call Custos/Orders/Payment.  
- Each client file wraps proto-generated stubs.  

### `internal/infrastructure/http/`
- **handler/**: HTTP handlers for external routes.  
- `router.go`: Sets up routes, integrates middleware.  

### `internal/middleware/`
- `auth.go`: Middleware using Mora Auth to validate Access Tokens.  

### `docs/`
- Additional architecture docs, API examples, OpenAPI specs.  

---

## ⚠️ AI Instructions
- Do not implement business/domain logic in Clotho. That belongs to Custos/Orders/Payment.  
- Always call domain services via gRPC clients (`internal/infrastructure/client/`).  
- All external APIs must use Mora Auth middleware for Access Token validation.  
- Expose HTTP endpoints only – no GraphQL, no Gateway logic.  
- Use `cobra` in `cmd/clotho` to implement CLI (serve, version).  
- Provide minimal but clear starter implementations (healthcheck, example user proxy).  

---

## 🚀 Next Steps for AI
1. Generate `cmd/clotho` with cobra-based CLI.  
2. Implement `internal/infrastructure/http/router.go` with Gin (or go-zero) as HTTP server.  
3. Create `internal/middleware/auth.go` using Mora Auth.  
4. Scaffold `internal/infrastructure/client/custos_grpc.go` with sample gRPC call.  
5. Add `internal/application/usecase/user_proxy.go` to orchestrate Custos.  
6. Add `GET /health` endpoint as a first handler.  