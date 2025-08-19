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

- [ ] Sidecar Mode for transparent proxy (mainly for K8s)
