# Paper Mapping

This document maps the concepts and equations described in the IEEE GSCon 2027 draft to the specific code implementations in the `simulator/` track.

## Section III: SentinelMesh Design

### Local Anomaly Scoring (Sec III.A)
- **Concept:** $O(1)$ exponentially-weighted moving baseline (EWMA) z-score.
- **Code:** `simulator/internal/scorer/scorer.go` (`FeatureEWMA.update()`)

### Gossip Digest Exchange (Sec III.B)
- **Concept:** Constant-size digest $D_i(t) = \langle \text{node\_id}_i, s_i(t), c_i(t), t \rangle$.
- **Code:** `simulator/internal/node/node.go` (`Digest` struct). The timestamp $t$ is implemented as discrete gossip rounds.
- **Mechanism:** Epidemic push exchange to $f$ random peers.
- **Code:** `simulator/internal/gossip/gossip.go` (Pending Implementation).

### Quorum-Based Escalation (Sec III.C)
- **Concept:** Equation 2 - $\text{Alert}(c, t) = \mathbb{1}[ |\{i : s_i(t') > \tau_{\text{local}}, c_i(t')=c, t-t'\le W\}| \ge q ]$
- **Code:** `simulator/internal/quorum/quorum.go` (Pending Implementation). It implements a strict indicator function over discrete rounds $W$, without arbitrary decay or smoothing.

## Section IV: Simulation Methodology
### Traffic Partitioning and Attack Fragmentation (Sec IV.B)
- **Concept:** Partitioning across $N$ nodes and k-way fragmentation of logical campaigns.
- **Code:** `simulator/internal/fragment/fragment.go`. Due to the nature of the dataset lacking diverse IP spaces, partitioning uses pseudo-random hashing on the `Flow.ID`. Attack fragmentation uses a round-robin distribution across $k$ nodes.
