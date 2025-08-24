# GlitchMesh 🧪

**GlitchMesh** is a lightweight, developer-focused proxy tool designed for **testing microservice resilience**. It allows you to simulate network faults like latency, errors, and dropped requests, making it easier to validate the robustness of distributed systems, all without touching production code.

---
> ⚠️ **Note:** This project is **under active development**. Many features are still being implemented, and things may change frequently. Feedback and contributions are highly appreciated!
---

## 🚀 Features

- 🧩 **Single-port intelligent proxy** – Routes requests to appropriate services based on rules
- 🐢 **Latency injection** – Add artificial delay to simulate slow downstream services
- 🔥 **Fault simulation** – Simulate service crashes, errors, and timeouts
- ⚙️ **Config-driven** – Define fault behavior in a simple YAML file
- 📜 **Easy logging** – Transparent logging of all injected faults and proxy activity

## Upcoming Features

**Network Faults**
  - [ ] Connection Timeouts - Simulate services that accept connections but never respond
  - [ ] Connection Drop - Randomly Dropping connections mid request
  - [ ] Bandwidth Throttling - Limiting throughput to simulate network congestion

**Response Manupulation**
  - [ ] Partial response corruption - Corrupt random bytes in response bodies

**Advanced Simulations**
  - [ ] Circuit breaker simulation - Fail fast after consecutive failures

**Other Features**
  - [ ] Hot reloading of configs - reloading configs without restarts
  - [ ] Sidecar Mode for transparent proxy (mainly for K8s)
  - [ ] API Based control - to enable/disable faults using API endpoints
  - [ ] Percentage Based Faults - Apply faults to only X% of the requests
