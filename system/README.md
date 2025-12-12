# System Module

The system module collects low-level runtime metrics from the host where Monad nodes are running (e.g. off-CPU waits, scheduler behavior, perf counters, turbostat, etc.) and forwards them to the Monad Flow backend.

---

## 1. Environment configuration

Create a `.env` file in the `system` directory:

```bash
TARGET_SERVICES=monad-bft.service,monad-execution.service
TARGET_TRACE_PID=1382013
BACKEND_URL=<backendURL>
```

- `TARGET_SERVICES`: comma-separated list of systemd units you want to monitor (e.g. `monad-bft.service`, `monad-execution.service`).
- `TARGET_TRACE_PID`: PID of the primary process you want to trace (typically the main `monad-execution` process).  
  - In production you should replace this with the actual PID or implement PID discovery before starting the system module.
- `BACKEND_URL`: URL of the Monad Flow backend API where collected metrics will be sent.

---

## 2. Required tools (before build)

Some parts of the system module rely on external profiling/tracing tools.  
You must install them yourself according to your distribution and environment.

At minimum:

- `offcputime-bpfcc` (from BCC / bpfcc-tools)

Example installation hints (Ubuntu-like systems):

```bash
sudo apt update

# BCC / bpfcc-tools (includes offcputime-bpfcc on many distros)
sudo apt install -y bpfcc-tools linux-headers-$(uname -r)
```

On other distributions, consult your package manager or the official BCC documentation and ensure:

- the `offcputime-bpfcc` binary is installed and
- it is in `PATH` (or you know its full path and configure the system module accordingly).

Depending on which submodules you enable, you may also need:

- `perf` (usually from `linux-perf` / `linux-tools` packages)
- `turbostat` (often from `linux-tools-common` / `linux-tools-$(uname -r)`)

---

## 3. Build

From the repository root:

```bash
cd system
go mod tidy
go build -o go-system .
```

This produces a binary named `go-system` in the `system` directory.

---

## 4. Run

Run the system module with sufficient privileges to use eBPF, perf, turbostat, and to inspect PIDs / systemd units:

```bash
cd system
sudo ./go-system
```

Make sure:

- the `.env` file is present in `system/`, and
- `TARGET_SERVICES`, `TARGET_TRACE_PID`, and `BACKEND_URL` are set appropriately for the host.

---

## 5. Run under PM2

To keep the system module running as a managed daemon, you can use `pm2`.

First ensure `pm2` is installed:

```bash
sudo npm install -g pm2
```

Then start the system module:

```bash
cd /path/to/monad-flow/system

pm2 start ./go-system --name go-system
```

If you need elevated privileges (e.g. for eBPF / perf), run `pm2` under `sudo` or configure your environment so that the process has the required capabilities:

```bash
sudo pm2 start ./go-system --name go-system
```

---

## 6. Function hooking & IDA workflow (DBI)

Monad Flow can hook specific functions (e.g. `TrieDb::commit`, `BlockExecutor::execute`) at runtime via Frida-based dynamic binary instrumentation.  
To do that safely and reproducibly, we recommend the following workflow.

### 6.1 Identifying function symbols in IDA

- Use IDA (or your preferred disassembler) on the `monad-execution` binary to:
  - locate candidate functions to hook (e.g. TrieDB-related code, block executor paths),
  - confirm their mangled names,
- Once you have symbol names, document them here and in the code comments, so the DBI configuration can be kept in sync with new releases of `monad-execution`.

In practice you will often see **only mangled symbols** and very few (or no) “stripped” names.  
We assume this is largely due to aggressive compiler optimizations and the fact that Monad, as a high‑performance chain, leans heavily on low‑level kernel features and toolchain optimizations.  
That means:

- you cannot always rely on symbol tables alone to identify the exact function you want to time, and
- the binary itself is not being “re‑optimized” later (which is important, because low‑level kernel interactions would easily break under post‑compile rewriting).

Instead, when reversing, you will typically:

- use **unique strings or data patterns** used inside the target function (e.g. log messages, error strings, metric labels) as anchors, and
- work backwards from those references in IDA to identify the true function boundaries you want to hook and measure.

This way, even when symbols are mangled or partially stripped, you can still robustly map “the function that uses this string” to a specific hook target.

![Execute Block Function handle](/docs/assets/execute_block_function_handle.png)
![Execute Block Function handle](/docs/assets/commit_function_handle.png)

### 6.2 Wiring function names into the hook configuration

- The DBI / Frida layer expects a list of target symbols (function names) to attach to.
- In the TypeScript configuration file (for example `system/main.ts` or an equivalent hook config module), you typically:
  - define an array of target symbols, and
  - pass it to the Frida script that actually installs `onEnter` / `onLeave` hooks.

Example:
```ts
// system/main.ts
var hookTargets = []string{
    ...
}
```

- When you add or change symbols in IDA, update this list and redeploy the system module so the new hooks take effect.

### 6.3 Warning: MPT lookup hooks and zombie processes

Some MPT / Trie-related lookups are extremely short-lived and called very frequently.  
Hooking them naïvely (especially with heavy logging or expensive timing logic) can:

- significantly pause the traced process, and
- in extreme cases, leave the process in a broken or zombie-like state if hooks block on I/O or crash mid-execution.

When instrumenting MPT lookups:

- keep hooks as lightweight as possible (minimal work in `onEnter` / `onLeave`),
- prefer sampling (hooking only a subset of calls) over tracing every single invocation,
- test on non-critical nodes first, and
- be prepared to immediately detach hooks or restart the process if instability is observed.
