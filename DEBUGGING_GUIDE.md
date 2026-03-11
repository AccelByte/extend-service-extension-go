# Debugging Guide — Extend Service Extension (Go)

> **Audience:** Junior developers and game developers who are new to backend development.
> This guide focuses on Visual Studio Code but most concepts apply to any IDE or editor.

---

## Table of Contents

1. [Overview](#1-overview)
2. [Understanding the Service Architecture](#2-understanding-the-service-architecture)
3. [Prerequisites](#3-prerequisites)
4. [Environment Setup for Local Development](#4-environment-setup-for-local-development)
5. [Running the Service Locally](#5-running-the-service-locally)
6. [Attaching the Debugger (VS Code)](#6-attaching-the-debugger-vs-code)
7. [Setting Breakpoints and Inspecting State](#7-setting-breakpoints-and-inspecting-state)
8. [Reading and Understanding Logs](#8-reading-and-understanding-logs)
9. [Common Issues and How to Diagnose Them](#9-common-issues-and-how-to-diagnose-them)
10. [Testing Endpoints Manually](#10-testing-endpoints-manually)
11. [Debugging with AI Assistance](#11-debugging-with-ai-assistance) *(optional — requires an AI assistant)*
12. [Tips and Best Practices](#12-tips-and-best-practices)

---

## 1. Overview

This document guides you through debugging an **Extend Service Extension** app written in Go.
An Extend Service Extension is a custom backend service hosted by AccelByte Gaming Services (AGS).
It exposes RESTful HTTP endpoints (powered internally by a gRPC server and gRPC-Gateway) that your
game client or other services can call.

Think of it this way:

```
Game Client / AGS
       │  HTTP (REST)
       ▼
 gRPC-Gateway  (port 8000)   ← translates HTTP ↔ gRPC
       │  gRPC
       ▼
 gRPC Server   (port 6565)   ← your business logic lives here
       │
       ▼
 AccelByte CloudSave / other AGS services
```

When something goes wrong, you need to trace the problem through one or more of these layers.
The sections below show you how to do that step by step.

---

## 2. Understanding the Service Architecture

Before you can debug effectively, it helps to know what each piece of code does.

| File / Package | What it does |
|---|---|
| `main.go` | Entry point. Wires together gRPC server, gRPC-Gateway, metrics, tracing, and auth. |
| `pkg/service/myService.go` | **Your business logic.** Implements the gRPC service methods (e.g., `CreateOrUpdateGuildProgress`, `GetGuildProgress`). |
| `pkg/storage/storage.go` | Talks to AccelByte CloudSave to persist and retrieve data. |
| `pkg/common/authServerInterceptor.go` | Validates every incoming request's IAM token and permission before it reaches your service logic. |
| `pkg/common/logging.go` | Bridges the gRPC middleware logger to Go's `slog`. |
| `pkg/common/tracerProvider.go` | Sets up OpenTelemetry distributed tracing (Zipkin exporter). |
| `pkg/proto/service.proto` | Defines the gRPC API contract — endpoints, request/response shapes, required permissions. |
| `pkg/pb/` | Auto-generated Go code from the `.proto` files. Do not edit directly. |

**Key port numbers** (defined as constants in `main.go`):

| Port | Purpose |
|---|---|
| `6565` | gRPC server (internal, used by gRPC-Gateway) |
| `8000` | gRPC-Gateway HTTP/REST (the port you call from a browser or Postman) |
| `8080` | Prometheus metrics endpoint (`/metrics`) |

---

## 3. Prerequisites

Make sure you have the following installed and configured before you start debugging.

- **Go 1.24+** — `go version` should show `go1.24.x` or later.
- **VS Code** with the **[Go extension](https://marketplace.visualstudio.com/items?itemName=golang.go)** (`golang.go`) installed.
  The repository includes a recommended extensions list in `.vscode/extensions.json` — VS Code will prompt you to install them when you open the folder.
- **Delve** — the Go debugger. The Go VS Code extension installs it for you automatically. You can verify with `dlv version`.
- **AccelByte credentials** — `AB_BASE_URL`, `AB_CLIENT_ID`, `AB_CLIENT_SECRET`, and `BASE_PATH` must be set. See section 4 for details.

---

## 4. Environment Setup for Local Development

The service reads its configuration from environment variables. A template file is provided.

### 4.1 Create your `.env` file

```
# In the repository root
cp .env.template .env
```

Then open `.env` and fill in the values. The critical ones are:

```dotenv
# AccelByte environment
AB_BASE_URL=https://<your-ags-environment>.accelbyte.io
AB_CLIENT_ID=<your-client-id>
AB_CLIENT_SECRET=<your-client-secret>

# The URL base path that this service is served under.
# For local development this is typically just a slash-prefixed name.
BASE_PATH=/guild

# Set to "false" to disable token validation during local development.
# NEVER disable this in production.
PLUGIN_GRPC_SERVER_AUTH_ENABLED=false

# Log verbosity: debug | info | warn | error
LOG_LEVEL=debug
```

> **Why `PLUGIN_GRPC_SERVER_AUTH_ENABLED=false`?**
> During local debugging, you probably don't have an AGS token handy for every test call.
> Disabling auth lets you call endpoints freely. Remember to turn it back on before deploying.

### 4.2 VS Code task shortcut

A ready-made VS Code task called **"Create .env File"** (see `.vscode/tasks.json`) can generate
the `.env` file for you if you have `fillenv` installed.

---

## 5. Running the Service Locally

### 5.1 From the terminal

```bash
# Make sure your .env file is sourced or variables are exported
export $(grep -v '^#' .env | xargs)

go run main.go
```

### 5.2 From VS Code

Use **Terminal → Run Task → "Run: Service"**.
This task is pre-configured in `.vscode/tasks.json` and automatically picks up `BASE_PATH`
from the VS Code input prompt.

### 5.3 Confirming the service is up

Once started you should see log lines similar to:

```json
{"time":"...","level":"INFO","msg":"app server started","service":"extend-app-service-extension"}
{"time":"...","level":"INFO","msg":"starting gRPC-Gateway HTTP server","port":8000}
{"time":"...","level":"INFO","msg":"serving prometheus metrics","port":8080,"endpoint":"/metrics"}
```

Open `http://localhost:8000<BASE_PATH>/apidocs/` in your browser to see the Swagger UI and
verify that the REST endpoints are live.

---

## 6. Attaching the Debugger (VS Code)

The repository ships with a ready-to-use launch configuration in `.vscode/launch.json`:

```jsonc
{
  "name": "Debug: Service",
  "type": "go",
  "request": "launch",
  "mode": "auto",
  "program": "${workspaceFolder}",
  "envFile": "${workspaceFolder}/.env",   // <-- loads your .env automatically
  "cwd": "${workspaceFolder}",
  "console": "integratedTerminal"
}
```

### Steps

1. Make sure your `.env` file is filled in (section 4).
2. Open the **Run and Debug** panel (`Ctrl+Shift+D` / `Cmd+Shift+D`).
3. Select **"Debug: Service"** from the dropdown at the top.
4. Press **F5** (or click the green ▶ button).

VS Code will compile the service with debug symbols and start it under **Delve**,
the official Go debugger. The service behaves identically to `go run main.go`, but now
you can pause execution, inspect variables, and step through code.

> **Other IDEs / editors:** If you're not using VS Code, you can start Delve manually:
> ```bash
> dlv debug . -- # Delve starts and listens; attach your IDE to it
> ```
> Or run in headless mode so your IDE can connect to it:
> ```bash
> dlv debug --headless --listen=:2345 --api-version=2 .
> ```

---

## 7. Setting Breakpoints and Inspecting State

### 7.1 Where to put breakpoints

| What you want to investigate | Suggested file and location |
|---|---|
| A specific REST endpoint being called | `pkg/service/myService.go` — top of the relevant method |
| Auth/token validation failure | `pkg/common/authServerInterceptor.go` — `NewUnaryAuthServerIntercept` |
| Data not saving/loading correctly | `pkg/storage/storage.go` — `SaveGuildProgress` / `GetGuildProgress` |
| Service not starting at all | `main.go` — where `os.Exit(1)` is called |

### 7.2 Setting a breakpoint

Click in the **gutter** (the area left of the line numbers) next to the line you want to pause on.
A red dot appears. When execution reaches that line, VS Code will pause and show you:

- **Variables** panel — all local variables and their current values.
- **Watch** panel — expressions you want to monitor continuously.
- **Call Stack** panel — how you got to this point (which function called which).
- **Debug Console** — evaluate arbitrary Go expressions live.

### 7.3 Stepping through code

| Action | Keyboard shortcut | What it does |
|---|---|---|
| Continue | `F5` | Run until the next breakpoint |
| Step Over | `F10` | Execute the current line; stay at the same level |
| Step Into | `F11` | Enter the function called on the current line |
| Step Out | `Shift+F11` | Finish the current function and return to the caller |
| Restart | `Ctrl+Shift+F5` | Restart the debug session |
| Stop | `Shift+F5` | Stop the debugger |

### 7.4 Conditional breakpoints

Right-click a breakpoint dot → **Edit Breakpoint** → type a condition (e.g., `req.GuildId == "guild_001"`).
The debugger will only pause when the condition is true — useful when you only want to inspect a specific request.

---

## 8. Reading and Understanding Logs

The service uses Go's structured `slog` package and emits **JSON log lines** to stdout.
Every log line looks like this:

```json
{"time":"2026-03-10T12:00:00Z","level":"INFO","msg":"HTTP request","method":"POST","path":"/guild/v1/admin/namespace/mygame/progress","duration":"1.234ms"}
```

### 8.1 Log levels

| Level | When it appears | Use for |
|---|---|---|
| `DEBUG` | Only when `LOG_LEVEL=debug` | Fine-grained tracing of logic branches, variable values |
| `INFO` | Default | Normal operational events (server started, request received) |
| `WARN` | Something unexpected but recoverable | Worth investigating |
| `ERROR` | Something failed | Always investigate |

Set `LOG_LEVEL=debug` in your `.env` during development so you see everything.

### 8.2 gRPC request/response logging

The service uses `go-grpc-middleware` to automatically log every gRPC call start, finish,
and payload (at `DEBUG` level). When you send an HTTP request to port 8000, look for log pairs like:

```json
{"msg":"started call","grpc.method":"CreateOrUpdateGuildProgress", ...}
{"msg":"finished call","grpc.code":"OK","grpc.duration":"2ms", ...}
```

If `grpc.code` is anything other than `OK` (e.g., `Unauthenticated`, `Internal`, `NotFound`),
scroll up to find the `ERROR` log that explains why.

### 8.3 Pretty-printing logs

Raw JSON is compact but hard to read. Pipe output through `jq` for prettier output during local runs:

```bash
go run main.go 2>&1 | jq '.'
```

---

## 9. Common Issues and How to Diagnose Them

### 9.1 Service fails to start — `BASE_PATH` error

**Symptom:**
```
{"level":"ERROR","msg":"BASE_PATH envar is not set or empty"}
```

**Cause:** The `BASE_PATH` environment variable is missing or doesn't start with `/`.

**Fix:** Set `BASE_PATH=/guild` (or any path starting with `/`) in your `.env`.

---

### 9.2 Service exits with "unable to login using clientId and clientSecret"

**Symptom:**
```json
{"level":"ERROR","msg":"error unable to login using clientId and clientSecret","error":"..."}
```

**Cause:** `AB_CLIENT_ID` / `AB_CLIENT_SECRET` are wrong, or `AB_BASE_URL` points to the wrong environment.

**Fix:** Double-check your credentials in `.env`. Verify the base URL is reachable: `curl $AB_BASE_URL/iam/v3/public/config`.

---

### 9.3 All requests return `401 Unauthenticated`

**Symptom:** Every HTTP call to port 8000 returns HTTP 401.

**Cause:** Token validation is enabled (`PLUGIN_GRPC_SERVER_AUTH_ENABLED=true`) and the bearer token
in your request is missing, expired, or lacks the required permission defined in the `.proto` file.

**Fix (for local debugging):** Set `PLUGIN_GRPC_SERVER_AUTH_ENABLED=false` in `.env`.
**Fix (for real tokens):** Use a valid AGS token with the required permissions. Check `pkg/proto/service.proto`
for the exact permission resource and action required by each endpoint.

---

### 9.4 Endpoint returns `500 Internal Server Error`

**Symptom:** HTTP 500 or gRPC `codes.Internal` response.

**How to diagnose:**
1. Check the terminal/log output for an `ERROR` level log around the time of the request.
2. Put a breakpoint in `pkg/service/myService.go` inside the failing method.
3. Step into the `g.storage.*` calls to see if CloudSave is returning an error.
4. Check that the namespace in your request matches a real namespace in your AGS environment.

---

### 9.5 Breakpoints are not hit

**Symptoms:** You set a breakpoint but execution never pauses there.

**Possible causes:**
- The code path is not reached (the request goes to a different endpoint or fails earlier, e.g., at auth).
- The build is optimized (rare in debug mode, but verify you used the VS Code launch config, not `go build -o`).
- There is a copy of the service already running on port 6565/8000, and your request is hitting that one instead.

**Fix:** Check for port conflicts with `ss -tlnp | grep -E '6565|8000'`. Kill stale processes if needed.

---

### 9.6 Changes to `.proto` files are not reflected

**Symptom:** You edited `pkg/proto/service.proto` but the change has no effect.

**Fix:** Regenerate the protobuf Go bindings by running the **"Proto: Generate"** VS Code task,
or directly:
```bash
./proto.sh
```
Then restart the service.

---

## 10. Testing Endpoints Manually

### 10.1 Swagger UI

Navigate to `http://localhost:8000<BASE_PATH>/apidocs/` in your browser.
The built-in Swagger UI lets you explore and call each endpoint directly from the browser.

### 10.2 curl

```bash
# Example: create or update guild progress
curl -s -X POST \
  "http://localhost:8000/guild/v1/admin/namespace/mygame/progress" \
  -H "Content-Type: application/json" \
  -d '{"guildProgress": {"guildId": "guild_001", "namespace": "mygame"}}' | jq .
```

Add `-H "Authorization: Bearer <your-token>"` when auth is enabled.

### 10.3 Postman

The `demo/` directory contains Postman collection files (`*.postman_collection.json`) with
pre-built requests. Import them into Postman and update the environment variables
(`baseUrl`, `namespace`, `token`) to match your local setup.

### 10.4 grpcurl (gRPC directly)

You can also call the gRPC server directly on port 6565, bypassing the HTTP gateway:

```bash
# List available services (reflection is enabled)
grpcurl -plaintext localhost:6565 list

# Call a method
grpcurl -plaintext -d '{"namespace":"mygame","guildProgress":{"guildId":"guild_001"}}' \
  localhost:6565 service.Service/CreateOrUpdateGuildProgress
```

---

## 11. Debugging with AI Assistance

> **This section is optional.** If your team does not use AI tooling, or if you are discouraged
> from using AI at work, skip this section entirely. Every other section in this guide is
> self-contained and does not require an AI assistant.

If you have access to an AI coding assistant (such as **Claude Code**, GitHub Copilot, or similar),
you can use it as a powerful debugging companion. The workflow below uses Claude Code as an example,
but the same principles apply to any AI tool.

### 11.1 Debugging skill for this repository

This repository ships with a **Claude skill** at [`.claude/skills/debugging-guide/`](.claude/skills/debugging-guide/SKILL.md).
A Claude skill is a reusable set of instructions that tells the AI *exactly* how to help with a
specific domain — in this case, debugging and documenting Extend Service Extension apps.

> **Requirement:** Claude skills are supported by **Claude Code** and several other AI provider or
> IDE extension that implements the [Agent Skills open standard](https://agentskills.io).
> If your tool does not support agent skills, you can still paste the contents of
> [`.claude/skills/debugging-guide/SKILL.md`](.claude/skills/debugging-guide/SKILL.md) directly
> into your AI chat window as a system prompt or first message.

Once the skill is active, you can invoke it in two ways:

| Intent | How to invoke |
|---|---|
| Debug a live issue | `/debugging-guide Go — getting 500 on GetGuildProgress` |
| Write or update the debugging guide | `/debugging-guide write Go` |
| Let the AI decide | Describe a debugging problem naturally — the AI loads the skill automatically |

### 11.2 What AI assistants are good at (and not so good at)

| Good at | Not so good at |
|---|---|
| Explaining unfamiliar code and libraries | Knowing the current state of your running process |
| Suggesting fixes for error messages | Accessing live environment credentials |
| Writing test cases for a specific function | Understanding your exact AGS namespace/tenant config |
| Identifying patterns in log output | Replacing thorough end-to-end testing |

Use AI as a smart colleague who can read the code with you — not as an oracle that knows everything.

### 11.3 MCP servers in this repository

This repository ships with two **Model Context Protocol (MCP)** server configurations in `.vscode/mcp.json`:

| Server | What it does |
|---|---|
| `extend-sdk` | Provides AI assistants with knowledge of the AccelByte Extend SDK symbols, types, and usage patterns |
| `ags-api` | Exposes live AGS REST APIs so an AI can query them on your behalf (requires `AB_BASE_URL`, `OAUTH_CLIENT_ID`, `OAUTH_CLIENT_SECRET` to be set in the environment) |

These servers extend the AI's awareness of AccelByte-specific code and services, making its answers
more accurate and grounded in the actual API surface.

### 11.4 Effective prompting for debugging

**Paste the error, not just the symptom.** Instead of asking "why does my service crash?",
copy the full error log or stack trace and paste it:

> *"My Extend Service Extension exits at startup with this log output. What is the most likely cause
> and how do I fix it?*
> ```json
> {"level":"ERROR","msg":"error unable to login using clientId and clientSecret","error":"401 Unauthorized"}
> ```"

**Include the relevant file.** AI assistants produce much better answers when they can see the code:

> *"Here is my `pkg/service/myService.go`. The `GetGuildProgress` method returns `codes.Internal`
> whenever the guild doesn't exist yet. Should I handle `NotFound` separately?"*

**Ask for explanations first, fixes second.** Understanding why something is wrong is more
valuable than just copying a fix you don't understand:

> *"Explain what `PLUGIN_GRPC_SERVER_AUTH_ENABLED` does in this codebase and why turning it off
> during development is safe."*

### 11.5 Asking AI to help read logs

Structured JSON logs are easy for AI to parse. Paste a block of logs and ask:

> *"Here are 20 lines of structured JSON logs from my Extend Service. Can you identify the request
> that failed and explain the sequence of events that led to the error?"*

### 11.6 Asking AI to help write targeted tests

Once you've identified a bug, ask the AI to write a unit test that would catch it:

> *"Write a Go unit test for `MyServiceServerImpl.GetGuildProgress` that covers the case where
> CloudSave returns a not-found error. Use the mock at `pkg/service/mocks/`."*

The repository already contains mocks in `pkg/service/mocks/` generated with `go.uber.org/mock`.

---

## 12. Tips and Best Practices

- **Always debug with `LOG_LEVEL=debug`** during local development. The extra gRPC payload logs
  often reveal the issue without needing a breakpoint.

- **Disable auth locally, enable it before committing.** `PLUGIN_GRPC_SERVER_AUTH_ENABLED=false`
  is a development shortcut, not a permanent setting.

- **Use conditional breakpoints** to avoid stopping on every iteration of a loop or every
  incoming request. Narrow the condition to the specific input you're investigating.

- **Check port availability** before starting the service if you see "address already in use":
  ```bash
  ss -tlnp | grep -E '6565|8000|8080'
  ```

- **Regenerate proto bindings after every `.proto` change.** Forgetting to run `./proto.sh`
  after editing `service.proto` is a common source of confusing compile or runtime errors.

- **Commit working `.env.template` changes** but never commit your actual `.env` file.
  It contains secrets. Confirm it is in `.gitignore`.

- **Read the Call Stack.** When the debugger pauses, look at the Call Stack panel in VS Code.
  Starting from the top (current frame) and reading down tells you exactly how execution
  arrived at the current line — invaluable for understanding where a request enters your code.

- **Use the Watch panel.** Add expressions like `req.Namespace`, `err`, or `len(guildProgress.Objectives)`
  to the Watch panel so you can monitor them across multiple steps without re-expanding variables.

---

## References

- [AccelByte Extend — Introduction](https://docs.accelbyte.io/gaming-services/modules/foundations/extend/)
  Overview of the Extend add-on: what it is, its three app types (Override, Service Extension, Event Handler),
  and how it fits into AGS.

- [AccelByte Extend Service Extension — Introduction](https://docs.accelbyte.io/gaming-services/modules/foundations/extend/service-extension/)
  Deep dive into the Service Extension type: gRPC + gRPC-Gateway architecture, protobuf contract, lifecycle management, and getting started guide.

- [Debugging Skill — `.claude/skills/debugging-guide/SKILL.md`](.claude/skills/debugging-guide/SKILL.md)
  AI agent skill bundled with this repository. Gives your AI assistant a structured playbook for
  diagnosing issues and writing or updating this guide. Requires Claude Code or any tool that
  supports the [Agent Skills open standard](https://agentskills.io).
  *Developers not using AI can ignore this file — this document is fully standalone.*
