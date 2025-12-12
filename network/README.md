# Network Module

This module captures packets on a given network interface (including Raptorcast chunks) and forwards structured data to the Monad Flow backend.

![Monad Data](/docs/assets/monad-data.png)

---

## 1. Environment configuration

Create a `.env` file in the `network` directory:

```bash
MTU=1480
# BACKEND_URL=no
BACKEND_URL=<backendURL>
```

- `MTU`: MTU used when parsing/capturing packets (default: `1480`).
- `BACKEND_URL`: URL of the Monad Flow backend API.  
  - Set to `no` or comment out to disable backend forwarding if you only want local debugging.

---

## 2. Prerequisites (before build)

From the repository root:

```bash
cd network

# Update package index
sudo apt update

# Install Clang for eBPF / packet capture code
sudo apt install -y clang

# Install kernel headers and libraries matching the current kernel
sudo apt install -y \
  linux-headers-$(uname -r) \
  linux-libc-dev \
  gcc-multilib \
  libbpf-dev
```

Compile the eBPF hook used for packet capture:

```bash
clang -O2 -g -target bpf \
  -I/usr/include/$(gcc -dumpmachine) \
  -c ./util/hook/packet_capture.c \
  -o ./util/hook/packet_capture.o
```

Fetch and tidy Go module dependencies:

```bash
go mod tidy
```

---

## 3. Build

From `network/`:

```bash
go build -o go-network .
```

This produces a binary named `go-network` in the `network` directory.

---

## 4. Run

You need to run the network module with sufficient privileges to attach eBPF hooks and capture packets on the target interface.

Run directly with Go:

```bash
sudo go run main.go <interface-name>
```

or run the built binary:

```bash
sudo ./go-network <interface-name>
```

Replace `<interface-name>` with the actual network interface, e.g. `eth0`, `ens3`, etc.

---

## 5. Run under PM2

To keep the network module running as a managed process, you can use `pm2`.

First, ensure `pm2` is installed:

```bash
sudo npm install -g pm2
```

Then start the built binary with `pm2` (recommended):

```bash
cd network

sudo pm2 start ./go-network --name go-network -- <interface-name>
```

If you prefer to run via `go run` through `pm2` (mainly for development):

```bash
sudo pm2 start go --name go-network -- run main.go <interface-name>
```
