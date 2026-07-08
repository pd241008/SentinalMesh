# SentinelMesh Results Summary & Root Cause Analysis

## Executive Summary
The discrepancy between the canonical table (~99% recall), the newer table (~50% recall), and the current pipeline (~3% recall) has been fully traced to the evolution of the **Matched Counterfactual Control (MCC)** alignment logic and its interaction with the subtraction mechanism in `metrics.go`.

---

## 1. The Canonical Config on the Current Pipeline
Using the absolute latest pipeline (post-alignment-fix commit `9a1468d`), the corrected recall drops even further. 

Current Pipeline Corrected Recall for `N=32`, `f=3`, `k=8`, `W=5`:
- **q=2:** Recon Corrected Recall = **3.15%**, DoS = **57.36%**
- **q=8:** Recon Corrected Recall = **11.66%**, DoS = **19.00%**

---

## 2. Root Cause of the Massive Drop
The massive drop across the three tables (99% -> 50% -> 3%) is caused by changes to `assignedNode` routing for control flows in `fragment.go`.

Here is the exact chronology:

1. **Canonical Table (~99%):** Generated on Jul 4 before MCC was working correctly or control alerts were populated. The `cAlerts` logic was ignored, meaning `correctedDetectedReconFlows++` triggered for almost every detection. Thus, Corrected Recall roughly equaled Uncorrected Recall (99%).
2. **"NEW" Table (~50.31%):** Generated today via `run_full_grid.py`, which executed `./simulate`—a **stale binary compiled on Jul 4**. That binary contained the *first* iteration of MCC, where control replacement flows were assigned using `assignedNode = int(hashID(flow.ID) % uint32(numNodes))` instead of `rrCount % actualK`. By scattering the normal replacement flows randomly, the control group artificially starved the specific attack nodes of traffic. Because background traffic was unnaturally low on those nodes in the control group, it generated far fewer noise alerts in the `escRound`. The MCC only partially subtracted noise, resulting in 50.31%. 
3. **Current Pipeline (~3.15%):** Today's commit (`9a1468d`) perfectly aligned the control flows by using `assignedNode = rrCount % actualK` for the replacement flows. This restored the exact baseline traffic density on the targeted attack nodes in the control group. Because the control group now perfectly mirrors the background traffic, it correctly triggers spurious noise alerts at the exact same rate as the treatment group. Since `q=2` has a massive 68.55% false positive rate (i.e., background noise naturally triggers a quorum), the `metrics.go` logic treats almost every true positive as a "NoiseCoincidence" and discards it, plummeting the True Corrected Recall to 3.15%.

---

## 3. Config Differences
The configurations in `test_stagger.yaml` and `sweep_default.yaml` are structurally identical for the parameters in question (W=5, thresholds, dataset splits, seeds 42/43/44). The difference you observed was purely due to executing a stale compiled binary (`./simulate` from Jul 4) versus running the current source code (`go run`).

---

## 4. Did the alignment fix touch recall calculation?
**Yes, mathematically.** The alignment fix in `fragment.go` changed how replacement normal flows are routed in the control group. By perfectly aligning the node targets, it correctly recreated the background noise on those specific nodes. Because `metrics.go` specifically relies on `cAlerts` to subtract `NoiseCoincidences`, accurately restoring the control group traffic fundamentally increased the number of subtracted true positives, plunging the corrected recall metric.

---

## 5. Monotonicity Check
Running the full 168-cell monotonicity script on the newly generated grid for the current pipeline reveals **54 cells that severely break monotonic decline in `q`**. 
For example, for `N=32`, `f=3`, `k=16`, Recon Recall drops to 26.28% at `q=2`, but then **rises** to 34.34% at `q=8`. The subtraction logic in `metrics.go` interacting with the perfectly aligned noise creates severe, non-monotonic inversion artifacts at lower values of `q`.
