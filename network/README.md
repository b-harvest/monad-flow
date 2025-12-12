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

---

## 6. Implementation notes & lessons learned

This network sidecar was a useful place to experiment with lower-level techniques and protocol details.  
The following subsections are reserved for documenting key findings and design decisions:

### 6.1 Packet sniffing with eBPF

At a high level, the network sidecar uses eBPF as a low‑overhead “tap” on the kernel’s networking path so we can observe real Monad traffic (including Raptorcast chunks) without modifying the node binaries.

- **Why eBPF instead of pcap?**  
  Traditional pcap-based tools copy *every* packet into user space and then filter there, which is expensive when the link is busy. With eBPF we can filter directly in the kernel (e.g. by port, protocol, interface) and only ship the packets we actually care about to the application, keeping overhead and noise low.

- **Where we hook**  
  - `util/hook/packet_capture.c` defines two TC classifier programs: `tc_ingress` and `tc_egress`.  
  - These are attached to the interface’s **ingress/egress qdisc**, so every packet entering or leaving the NIC passes through `parse_packet`.
  - Inside `parse_packet`, we parse L2/L3/L4 headers and filter on TCP/UDP packets where either source or destination port is `8000` (the Monad traffic we care about).

- **What we capture**  
  - For matching packets, the eBPF program allocates a `struct pkt_sample` from a **ring buffer map** (`events`) and copies up to `MAX_PKT_SIZE` bytes of the raw frame into `sample->data`, storing the original length in `sample->len`.  
  - The sample is then submitted to user space via `bpf_ringbuf_submit`, making the packet payload available without blocking the kernel fast path.

- **How Go consumes it**  
  - In `network/main.go`, `util.NewBPFMonitor` sets up the eBPF program and exposes a `ringbuf.Reader` (`monitor.RingBufReader`).  
  - The main loop reads each record, decodes the first 4 bytes as the real packet length, and passes the remaining bytes into `parser.ParsePacket`, which then routes the result to the TCP/UDP managers for further decoding and performance analysis.

This design lets us treat the kernel as a high‑fidelity packet tap: we can reconstruct Monad / Raptorcast flows, measure timings, and correlate them with higher‑level metrics, all with minimal overhead on the data path.

### 6.2 ChaCha20 libraries in Rust vs Go

For consensus‑critical code, “same algorithm name” does not guarantee identical behavior across languages.

- **Different seed expansion**  
  - Rust’s `rand_chacha` does not treat a `u64` seed as a zero‑padded 32‑byte key.  
  - Instead, `SeedableRng::seed_from_u64` feeds the `u64` into a PCG32 PRNG and uses multiple PCG32 outputs to *expand* it into a full 32‑byte ChaCha20 key.  
  - A naïve Go implementation that simply writes `round` into the first 8 bytes of a 32‑byte key (and zeros the rest) will therefore produce a completely different key and different random stream.

- **Porting the key derivation**  
  - To make Rust and Go agree, we re‑implemented the exact PCG32‑based key expansion used by Rust in Go (same constants, rotations, and iteration count) and used that to generate the ChaCha20 key from the round value.  
  - After this change, both implementations produced identical 32‑byte keys and thus the same ChaCha20 stream.

- **Endianness pitfalls**  
  - Rust’s `U256` (e.g. `alloy_primitives::U256`) typically interprets bytes as little‑endian, whereas Go’s `big.Int.SetBytes` treats them as big‑endian.  
  - Without compensating for this (e.g. reversing byte order on the Go side), the same ChaCha20 output bytes would be interpreted as different integers, leading to divergent leader selection despite “matching” randomness.

The takeaway: for cross‑language determinism, you must fully specify and match **key derivation + endianness**, not just “use ChaCha20” on both sides.

### 6.3 Porting Raptorcast decoding logic to Go

The Raptorcast decoding logic used by `monad-bft` was originally implemented in Rust (in the `monad-raptorcast` / `monad-raptor` crates).  
For this sidecar, the relevant parts were **ported to Go**, including:

- the core decoding routines for Raptorcast symbols, and
- the shared static parameters and constants defined by **RFC 5053** (so that encoding/decoding behavior matches the Rust implementation exactly).

Instead of documenting every helper and type here, the important point is:

- the Go decoder should be considered a direct, spec‑aligned port of the Rust reference used in production `monad-bft`.

In the main packet‑processing flow, the decoder is invoked from the UDP managers after parsing and reassembling Raptorcast payloads.  
The long‑term goal of this section is to provide a **simple, reusable decoder entry point** that anyone can adopt when building a new Go implementation of `monad-bft` or related tools.

You can later add a short, canonical usage snippet here (e.g. a small `Decode(...)` helper) so future Go projects can plug in the same Raptorcast decoder without re‑reverse‑engineering the Rust implementation.

```ts
// network/udp/manager.go
func (m *Manager) processChunk(
	packet model.Packet,
	chunkData []byte,
	captureTime time.Time,
) (map[string]interface{}, error) {
	chunk, err := parser.ParseMonadChunkPacket(packet, chunkData)
	if err != nil {
		return nil, fmt.Errorf("chunk parsing failed: %w (data len: %d)", err, len(chunkData))
	}

	senderInfo, err := util.RecoverSenderHybrid(chunk, chunkData)
	if err != nil {
		log.Printf("[Recovery-Warn] Failed to recover sender: %v", err)
	}

	jsonData, err := json.Marshal(chunk)
	if err != nil {
		log.Printf("JSON marshaling failed: %s", err)
		return nil, err
	}

	payload := map[string]interface{}{
		"type":        util.MONAD_CHUNK_PACKET_EVENT,
		"data":        json.RawMessage(jsonData),
		"timestamp":   captureTime.UnixMicro(),
		"secp_pubkey": senderInfo.NodeID,
	}
  
    // this is entry point of raptorcast decoding
	decodedMsg, err := m.decoderCache.HandleChunk(chunk)
	if err != nil {
		if !errors.Is(err, decoder.ErrDuplicateSymbol) {
			return nil, fmt.Errorf("raptor processing error: %w", err)
		}
	}

	if decodedMsg != nil {
		appMessageHash := fmt.Sprintf("0x%x", decodedMsg.AppMessageHash)
		if err := parser.HandleDecodedMessage(decodedMsg.Data, appMessageHash); err != nil {
			log.Printf("[RLP-ERROR] Failed to decode message: %v", err)
		}
	}
	return payload, nil
}
```