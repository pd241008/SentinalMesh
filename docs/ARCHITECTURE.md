# SentinelMesh Architecture

## Multi-Track Monorepo
The SentinelMesh project is divided into three parallel tracks to separate the core simulation from machine learning validation and front-end visualization.

### Track 1: Simulator (Go)
**Location:** `simulator/`
**Role:** The core discrete-event simulation engine.
- Simulates independent IDS nodes over discrete gossip rounds.
- Handles dataset loading, flow ingestion, deterministic node partitioning, and k-way campaign splitting.
- Implements the local EWMA scorer and gossip-based quorum consensus.

### Track 2: ML Crosscheck (Python)
**Location:** `ml-crosscheck/`
**Role:** Independent scorer validation.
- Validates the Go-based $O(1)$ EWMA scorer against traditional ML models (e.g., Isolation Forest, Autoencoders).
- Uses the identical parsed datasets to ensure fair comparison.

### Track 3: Dashboard (Next.js)
**Location:** `dashboard/`
**Role:** Interactive frontend for visualizing simulation outputs.
- Reads result contracts to display metrics like recall, bandwidth overhead, and convergence latency.

## Data Contracts
The three tracks communicate via shared file contracts located in the `results/` directory.

- **Sweep Results (`results/sweep/`)**: The Go simulator outputs `.csv` files containing the evaluation metrics across varying parameters ($N$, $f$, $q$, $k$). Track 3 (Dashboard) consumes these CSVs for visualization.
- **Crosscheck Validation (`results/crosscheck/`)**: Track 2 (Python) outputs validation summaries (JSON/CSV) to compare the efficacy of the ML baselines against the gossip consensus.
