# SentinelMesh Project Progress

This document tracks the completion status of the SentinelMesh project across its three distinct tracks: Simulator (Go), ML Validation (Python), and Dashboard (Next.js).

## Track 1: Core Simulator (Go)
**Status: Phase 1 Complete**

### Phase 1: Core Simulator Architecture
- [x] **Sub-phase 1.1: Project Setup & Data Modeling**
  - [x] Scaffolded multi-track mono-repo directory structure.
  - [x] Initialized Go module (`github.com/pd241008/sentinelmesh/simulator`).
  - [x] Created automated dataset fetch script (`data/scripts/fetch_dataset.sh`).
  - [x] Implemented `dataset.go`: Parsing UNSW-NB15 flows with strict labeling.
  - [x] Implemented `fragment.go`: Pseudo-random deterministic node partitioning (by Flow ID) and k-way targeted campaign splitting (round-robin).
- [x] **Sub-phase 1.2: Local Detection & Node Foundation**
  - [x] Configured sweep parameters (`configs/sweep_default.yaml`).
  - [x] Implemented `scorer.go`: $O(1)$ multi-feature EWMA z-score calculator (strict evaluation, no arbitrary decay).
  - [x] Implemented `node.go`: Independent node logic, digest caching by discrete round, and per-round flow ingestion.
- [x] **Sub-phase 1.3: Distributed Mechanism (Gossip & Quorum)**
  - [x] Implement `quorum.go`: Strict Equation 2 escalation rule evaluation.
  - [x] Implement `gossip.go`: Epidemic push-based exchange logic and random peer selection.
- [x] **Sub-phase 1.4: Orchestration, Baselines & Metrics**
  - [x] Implement `baseline/`: Independent (isolated) and Centralized aggregator simulation baselines.
  - [x] Implement `metrics/`: Logic for recall, bandwidth, and convergence latency tracking.
  - [x] Implement `sweep/`: Automated execution loop for N/f/q/k parameters.
  - [x] Implement `cmd/simulate/main.go`: CLI entry point.

## Track 2: ML Crosscheck (Python)
**Status: Pending**
- [ ] Setup `pyproject.toml` and testing frameworks.
- [ ] Data loader reusing partitioned UNSW-NB15 data.
- [ ] Scorer models (Isolation Forest / Autoencoder).
- [ ] Automated validation and comparison reports.

## Track 3: Dashboard (Next.js)
**Status: Pending**
- [ ] Scaffolding and build config (Next.js, TS).
- [ ] Data loader for `results/sweep` and `results/crosscheck`.
- [ ] Interactive charting components.
- [ ] Sweep visualization and ML cross-check views.
