# UNSW-NB15 Dataset Processing

## Provenance
The simulation utilizes the standard **UNSW-NB15** test and train splits, sourced from the `CyberSecurityUP` GitHub mirror. 
- **Train Set:** `UNSW_NB15_training-set.csv` (175,341 rows)
- **Test Set:** `UNSW_NB15_testing-set.csv` (82,332 rows)

The automated script `data/scripts/fetch_dataset.sh` handles the acquisition of these files.

## Partitioning Rationale
Traditional network intrusion detection evaluates datasets holistically at a single observation point. SentinelMesh requires evaluating distributed signal across $N$ nodes.

### Deterministic Pseudo-Random Assignment
Because the original raw captures of the UNSW-NB15 dataset only utilized approximately 45 synthetic testbed IPs, strict subnet-based grouping would fail to provide meaningful diversity across large mesh sizes (e.g., $N=64$). 

To resolve this, normal background flows are partitioned across the simulated nodes using **deterministic pseudo-random assignment** based on a hash of the integer Flow ID. This ensures an even, reproducible distribution of benign traffic across the mesh.

### K-Way Campaign Fragmentation
To simulate "low-and-slow" distributed attacks (such as coordinated reconnaissance or credential stuffing), malicious flows of a specific category are isolated and fragmented across $k$ nodes using a **round-robin** distribution. 

This guarantees that each node observes only $1/k$ of the attack volume, forcing the individual per-node IDS threshold ($\tau_{\text{local}}$) to fail, and requiring the gossip-based correlation mechanism to recover the detection signal.
