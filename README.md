# gcurl: Enterprise-Grade, High-Performance cURL Replica in Go

`gcurl` is an open-source, client-side networking utility and reusable library written entirely in Go. Engineered with a focus on low-level system efficiency, explicit concurrency boundaries, and advanced distributed telemetry, `gcurl` acts as a drop-in cURL replica while expanding into a high-throughput load and validation engine.

Unlike standard scripts that wrap high-level HTTP abstractions, `gcurl` exposes and controls the full lifecycle of the transport layer—from explicit Layer 4 TCP socket dial deadlines and custom DNS resolution down to Layer 7 payload streaming interfaces (`io.Reader`/`io.Writer`) and synchronous OpenTelemetry trace propagation.

---

## 🏗️ Architectural Core & Design Philosophy

`gcurl` is built to enforce a clean separation between its CLI ingress interface and its core execution engine. The codebase architecture effectively delivers two components inside a single repository: `gcurl` (the POSIX-compliant binary tool) and `libgcurl` (the decoupled, thread-safe underlying library).

```text
 Terminal / Ingress Inputs (gcurl -X POST -H "..." http://api)
          │
          ▼
 ┌────────────────────────────────────────────────────────┐
 │ COBRA & PFLAG INGRESS LAYER (internal/cli)             │
 │  └─► POSIX Compliance, Input Sanitization Gates        │
 └────────┬───────────────────────────────────────────────┘
          │
          ▼ (Serializes & Maps to Immutable Boundary Contract)
 ┌────────────────────────────────────────────────────────┐
 │ BOUNDARY SPECIFICATION CORE (pkg/config)               │
 │  └─► struct config.RequestConfiguration                │
 └────────┬───────────────────────────────────────────────┘
          │
          ▼ (Injected dynamically down the pipeline)
 ┌────────────────────────────────────────────────────────┐
 │ TRANSPORT & TELEMETRY ENGINE (pkg/transport)           │
 │  ├─► Custom Layer 4 TCP Dialer (DNS Overrides)         │
 │  ├─► Low-Overhead zero-copy streaming buffers          │
 │  └─► Synchronous Ephemeral OTel Trace Exporter         │
 └────────────────────────────────────────────────────────┘
```
