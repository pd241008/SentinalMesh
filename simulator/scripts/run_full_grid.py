#!/usr/bin/env python3
"""
Full Grid Sweep Orchestrator — Phase 1

Runs the Go sweep binary across 3 seeds, merges results, computes
mean±std per cell, and runs automated validation checks.

Outputs:
  results/full_grid/master_grid.csv — aggregated results
  results/full_grid/flags.md        — anomaly flags
  results/full_grid/seed_*/         — per-seed raw CSVs
"""
import os
import subprocess
import sys

import numpy as np
import pandas as pd

SIMULATOR_DIR = "/root/workspace/workspace/03-Code/Projects/Legacy/SentinalMesh/simulator"
OUTPUT_BASE = "/root/workspace/workspace/03-Code/Projects/Legacy/SentinalMesh/results/full_grid"
SEEDS = [42, 43, 44]

# ---------------------------------------------------------------------------
# Phase 1a: Run all seeds
# ---------------------------------------------------------------------------
def run_seeds():
    os.makedirs(OUTPUT_BASE, exist_ok=True)
    frames = []
    for seed in SEEDS:
        seed_dir = os.path.join(OUTPUT_BASE, f"seed_{seed}")
        print(f"\n{'='*60}")
        print(f"Running sweep with seed {seed}...")
        print(f"{'='*60}")
        cmd = [
            "./simulate",
            "--data", "../data/raw/UNSW_NB15_testing-set.csv",
            "--config", "configs/sweep_default.yaml",
            "--seed", str(seed),
            "--output", seed_dir,
        ]
        result = subprocess.run(cmd, cwd=SIMULATOR_DIR)
        if result.returncode != 0:
            print(f"FATAL: sweep failed for seed {seed}", file=sys.stderr)
            sys.exit(1)

        csv_path = os.path.join(seed_dir, "sweep_results.csv")
        df = pd.read_csv(csv_path)
        df["seed"] = seed
        frames.append(df)
        print(f"  → {len(df)} rows from seed {seed}")

    return pd.concat(frames, ignore_index=True)


# ---------------------------------------------------------------------------
# Phase 1b: Aggregate mean ± std per cell
# ---------------------------------------------------------------------------
def aggregate(combined: pd.DataFrame) -> pd.DataFrame:
    group_cols = ["N", "f", "k", "q", "W"]
    metric_cols = [c for c in combined.columns if c not in group_cols + ["seed"]]

    agg_dict = {m: ["mean", "std"] for m in metric_cols}
    stats = combined.groupby(group_cols).agg(agg_dict).reset_index()

    # Flatten multi-level columns
    stats.columns = [
        "_".join(col).rstrip("_") if col[1] else col[0]
        for col in stats.columns.values
    ]

    out_path = os.path.join(OUTPUT_BASE, "master_grid.csv")
    stats.to_csv(out_path, index=False)
    print(f"\nWrote aggregated results ({len(stats)} cells) to {out_path}")
    return stats


# ---------------------------------------------------------------------------
# Phase 1c: Automated Validation Checks
# ---------------------------------------------------------------------------
def validate(stats: pd.DataFrame, combined: pd.DataFrame) -> str:
    flags = []

    # --- Check 1: Independent recall flatness across q per (N, f, k) ---
    for (n, f, k), grp in stats.groupby(["N", "f", "k"]):
        for cat, col in [("Recon", "indep_flow_recon_recall_mean"),
                         ("DoS", "indep_flow_dos_recall_mean")]:
            if col not in grp.columns:
                continue
            vals = grp[col].dropna()
            if len(vals) > 1 and vals.std() > 0.01:
                flags.append(
                    f"**CHECK 1 FAIL** — Independent {cat} recall varies across q "
                    f"for N={n}, f={f}, k={k}: std={vals.std():.4f}, "
                    f"values={vals.tolist()}"
                )

    # --- Check 2: Centralized recall >= Gossip recall at every q ---
    for _, row in stats.iterrows():
        cell_id = f"N={int(row['N'])}, f={int(row['f'])}, k={int(row['k'])}, q={int(row['q'])}"
        for cat, gcol, ccol in [
            ("Recon", "gossip_flow_recon_recall_mean", "cent_flow_recon_recall_mean"),
            ("DoS", "gossip_flow_dos_recall_mean", "cent_flow_dos_recall_mean"),
        ]:
            if gcol in row and ccol in row:
                g_val = row[gcol]
                c_val = row[ccol]
                if pd.notna(g_val) and pd.notna(c_val) and g_val > c_val + 0.001:
                    flags.append(
                        f"**CHECK 2 FAIL** — Gossip {cat} recall ({g_val:.4f}) > "
                        f"Centralized ({c_val:.4f}) at {cell_id}"
                    )

    # --- Check 3: Proportional retention band (DoS vs Recon within ~10pp) ---
    for _, row in stats.iterrows():
        cell_id = f"N={int(row['N'])}, f={int(row['f'])}, k={int(row['k'])}, q={int(row['q'])}"
        g_recon = row.get("gossip_flow_recon_recall_mean", None)
        g_dos = row.get("gossip_flow_dos_recall_mean", None)
        i_recon = row.get("indep_flow_recon_recall_mean", None)
        i_dos = row.get("indep_flow_dos_recall_mean", None)

        if all(pd.notna(v) and v > 0 for v in [g_recon, g_dos, i_recon, i_dos]):
            ratio_recon = g_recon / i_recon if i_recon > 0 else 0
            ratio_dos = g_dos / i_dos if i_dos > 0 else 0
            gap = abs(ratio_recon - ratio_dos)
            if gap > 0.10:
                flags.append(
                    f"**CHECK 3 FLAG** — Proportional retention gap={gap:.4f} "
                    f"(Recon={ratio_recon:.4f}, DoS={ratio_dos:.4f}) at {cell_id}"
                )

    # --- Check 4: Monotonicity of recall and FPR as q increases ---
    mono_flags = []
    for (n, f, k), grp in stats.groupby(["N", "f", "k"]):
        grp_sorted = grp.sort_values("q")
        cell_prefix = f"N={n}, f={f}, k={k}"

        for label, col in [
            ("Gossip Recon Recall", "gossip_flow_recon_recall_mean"),
            ("Gossip DoS Recall", "gossip_flow_dos_recall_mean"),
            ("Gossip Recon FPR", "gossip_recon_fpr_mean"),
            ("Gossip DoS FPR", "gossip_dos_fpr_mean"),
        ]:
            if col not in grp_sorted.columns:
                continue
            vals = grp_sorted[col].values
            qs = grp_sorted["q"].values
            for i in range(1, len(vals)):
                if vals[i] > vals[i-1] + 0.005:  # small tolerance for noise
                    mono_flags.append(
                        f"**MONOTONICITY FLAG** — {label} increases from "
                        f"q={int(qs[i-1])}→q={int(qs[i])} "
                        f"({vals[i-1]:.4f}→{vals[i]:.4f}) at {cell_prefix}"
                    )

    flags.extend(mono_flags)

    return flags


def write_flags(flags: list, stats: pd.DataFrame):
    out_path = os.path.join(OUTPUT_BASE, "flags.md")
    total_cells = len(stats)

    with open(out_path, "w") as f:
        f.write("# Full Grid Validation Flags\n\n")
        f.write(f"**Total cells evaluated**: {total_cells}\n")
        f.write(f"**Seeds**: {SEEDS}\n")
        f.write(f"**Total flags**: {len(flags)}\n\n")

        if not flags:
            f.write("> [!TIP]\n")
            f.write("> All validation checks passed. No anomalies detected.\n")
        else:
            f.write("## Flagged Cells\n\n")
            for i, flag in enumerate(flags, 1):
                f.write(f"{i}. {flag}\n")

    print(f"\nWrote {len(flags)} flags to {out_path}")


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------
if __name__ == "__main__":
    combined = run_seeds()
    stats = aggregate(combined)
    flags = validate(stats, combined)
    write_flags(flags, stats)

    print(f"\n{'='*60}")
    print(f"DONE — {len(flags)} flags raised across {len(stats)} cells")
    print(f"{'='*60}")
    if flags:
        print("\nFlagged issues:")
        for fl in flags[:10]:
            print(f"  • {fl}")
        if len(flags) > 10:
            print(f"  ... and {len(flags) - 10} more (see flags.md)")
