# BRPOOL (brpool.org) 🇧🇷

An ultra-low latency, high-performance Bitcoin mining pool tailored specifically for Brazilian home miners. Built from scratch in **Golang** for maximum efficiency and competitive speed.

---

## 🚀 The Vision

The Bitcoin network was born to be decentralized, and true decentralization starts at home. **BRPOOL** is a specialized mining pool dedicated exclusively to Brazilian home miners operating low-hashrate micro-hardware. 

Whether you are running a lottery miner on your desk or a small home setup, network latency shouldn't be the reason you miss a block. By hosting our infrastructure closer to home and leveraging Go's highly concurrent architecture, we deliver the lowest possible latency for the Brazilian community.

We proudly support and empower **Solo Mining**. Your hardware, your house, your lottery ticket to financial sovereignty.

---

## 🛠️ Supported Hardware

BRPOOL is optimized to handle the connection dynamics and share submissions of low-power, localized devices, including but not limited to:
*   **NerdMiner** (All ESP32-based variations)
*   **NerdAxe**
*   **Hydro 6.1**
*   Other micro-hardware, DIY setups, and legacy ASIC miners.

---

## ✨ Features

*   **Written in Go:** Built for speed, low memory footprint, and high concurrency to process shares instantly.
*   **Localized Routing:** Designed specifically for the Brazilian network infrastructure to minimize network hops and ping times.
*   **Solo Mining Optimization:** Tailored difficulty adjustment (VarDiff) perfectly calibrated for micro-hashrate devices so they never drop connection or flood the stratum with stale shares.
*   **Lightweight Stratum Protocol:** Clean implementation of the Stratum mining protocol to ensure stable uptime.

---

## ⚙️ Architecture Overview

The pool is structured to keep resource usage minimal while ensuring that when a home miner hits a valid block, it is broadcasted to the Bitcoin network instantly.

*   **Stratum Server:** Handles TCP connections from miners with highly optimized Go goroutines.
*   **Share Processor:** Validates incoming shares concurrently without blocking the main network thread.
*   **Template Provider:** Syncs with local Bitcoin full nodes/Bitcoin Core to fetch the latest block templates immediately.

---

## 🛠️ Getting Started (Development)

### Prerequisites
*   Go (version 1.22 or higher)
*   Access to a Bitcoin Full Node (`bitcoind`) with RPC enabled.

### Installation
1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/brpool.git
   cd brpool
   ```

2. Configure your environment variables or config.json file (see config.example.json for reference).

3. Build the project:
   ```bash
   go build -o brpool ./cmd/poolserver/main.go
   ```

4. Run the pool server:
   ```bash
   ./brpool
   ```
