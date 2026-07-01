# SentinelMesh: Gossip-Propagated Collective Anomaly Detection

**SentinelMesh** is a decentralized anomaly correlation framework designed for distributed network intrusion sensing. Modern network defense typically relies on centralized Security Information and Event Management (SIEM) pipelines, which create latency bottlenecks, incur massive bandwidth costs, and introduce a single point of failure.

SentinelMesh replaces the centralized aggregator with a lightweight, decentralized **gossip-based correlation** mechanism. Independent IDS nodes exchange compact anomaly summaries via an epidemic protocol, utilizing a quorum consensus rule to collectively escalate "low-and-slow" attacks (e.g., distributed port scans, credential stuffing) that appear statistically normal to any single node.

---

## 🏛 Architecture

### System Topology
```mermaid
graph TD
    subgraph "SentinelMesh (Decentralized Gossip)"
    N1((Node 1)) <-->|Constant-Size Digest| N2((Node 2))
    N2 <-->|Constant-Size Digest| N3((Node 3))
    N3 <-->|Constant-Size Digest| N1
    N4((Node 4)) <-->|Constant-Size Digest| N2
    N1 <-->|Constant-Size Digest| N4
    end

    subgraph "Traditional SIEM (Centralized)"
    E1((Edge 1)) -->|Raw Logs/Flows| Agg[SIEM Aggregator]
    E2((Edge 2)) -->|Raw Logs/Flows| Agg
    E3((Edge 3)) -->|Raw Logs/Flows| Agg
    E4((Edge 4)) -->|Raw Logs/Flows| Agg
    end
```

### Node Workflow
```mermaid
sequenceDiagram
    participant Traffic as Network Traffic
    participant Local as Local Scorer (Node i)
    participant Cache as Digest Cache
    participant Peers as Network Peers

    Traffic->>Local: Ingest local flow partition
    Local->>Local: Compute O(1) EWMA z-score & tag category
    Local->>Cache: Store self digest 
    
    loop Every Gossip Round
        Local->>Peers: Push Digest to f random peers
        Peers-->>Cache: Receive incoming peer digests
        Cache->>Cache: Retain latest digest per peer in window W
        
        alt |{peers with score > \tau_{local}}| \ge q
            Cache->>Cache: Trigger Collective Alert (Quorum Escalation)!
        end
    end
```

---

## 📂 Project Structure (Multi-Track Monorepo)

This repository is organized into three parallel tracks to support simulation, machine-learning validation, and data visualization.

- **`simulator/` (Track 1 - Go)**: The core discrete-event simulation engine. Handles the parsing of the UNSW-NB15 dataset, pseudo-random node partitioning, $O(1)$ EWMA scoring, epidemic push-gossip exchange, and the quorum escalation rule.
- **`ml-crosscheck/` (Track 2 - Python)**: Independent scorer validation. Uses models like Isolation Forests and Autoencoders to cross-check the Go scorer's escalations and generate validation summaries.
- **`dashboard/` (Track 3 - Next.js)**: A frontend web application for interactive sweep result exploration. Visualizes metrics such as recall, bandwidth overhead, and convergence latency across variables like mesh size ($N$) and fanout ($f$).

*Supporting directories include `data/` (datasets and fetch scripts), `docs/` (architecture & progress tracking), `results/` (shared output contract), and `paper/` (LaTeX sources).*

---

## 🚀 Quickstart

### 1. Fetch the Dataset
The simulation utilizes the standard UNSW-NB15 dataset. Download it using the provided script:
```bash
./data/scripts/fetch_dataset.sh
```

### 2. Run the Simulator
Navigate to the simulator track and run a parameter sweep using the default configuration values. *(Execution CLI is under active development)*:
```bash
cd simulator
go run cmd/simulate/main.go --config configs/sweep_default.yaml
```

### 3. View Results
Results are written as CSV output to the `results/sweep/` directory. You can start the Next.js dashboard to interactively visualize the evaluation metrics:
```bash
cd dashboard
npm install
npm run dev
```

---

## 🧪 Testing

### Unit Tests
Each Go package has accompanying unit tests. Run all simulator tests from the `simulator/` directory:

```bash
cd simulator
go test ./... -v
```

### Integration Tests
Integration tests covering the full pipeline (data loading → fragmentation → gossip → quorum → metrics) live in `simulator/tests/`:

```bash
cd simulator
go test ./tests/... -v
```

### Test Data
A small synthetic CSV dataset is provided at `simulator/testdata/testdata.csv` with 15 flows across multiple attack categories. It is used by both unit and integration tests.

### Running All Tests
```bash
cd simulator
go test ./... -count=1
```

---

## 📊 Evaluation Goals

Based on discrete-event simulation using partitioned UNSW-NB15 traffic, this framework aims to measure:
- **Detection Recall**: The system's ability to recover detection capability for fragmented reconnaissance against baseline isolated edge nodes.
- **Bandwidth Overhead**: The reduction in peak single-point load compared to a centralized SIEM, tracking per-node payload costs.
- **Convergence Latency**: The scaling behavior of gossip propagation across varying mesh sizes, measured in discrete gossip rounds.