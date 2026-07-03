import pandas as pd
import numpy as np

# 1. Variance across seeds for N=16 dip
print("=== Latency Variance Across Seeds (k=8, q=4, f=3) ===")
dfs = []
for seed in [42, 43, 44]:
    df = pd.read_csv(f"/root/workspace/workspace/03-Code/Projects/Legacy/SentinalMesh/results/full_grid/seed_{seed}/sweep_results.csv")
    df['seed'] = seed
    dfs.append(df)

all_df = pd.concat(dfs)
target_df = all_df[(all_df['k']==8) & (all_df['q']==4) & (all_df['f']==3)].copy()

for n in [8, 16, 32, 64]:
    sub = target_df[target_df['N']==n]
    lats = sub['gossip_latency'].values
    print(f"N={n}: Mean={np.mean(lats):.4f}, Std={np.std(lats):.4f}, Values={lats}")

