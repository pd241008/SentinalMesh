import pandas as pd

df = pd.read_csv("/root/workspace/workspace/03-Code/Projects/Legacy/SentinalMesh/results/full_grid/master_grid.csv")

# Extract the N=64, q=4 gap table for k=4, k=8, k=16 (f=3 maybe?)
print("=== Asymmetry Gap Table (N=64, q=4, f=3) ===")
gap_df = df[(df['N']==64) & (df['q']==4) & (df['f']==3)]
for _, row in gap_df.iterrows():
    gap = row['gossip_flow_recon_recall_mean'] - row['gossip_flow_dos_recall_mean']
    print(f"k={int(row['k'])}: Recon={row['gossip_flow_recon_recall_mean']:.4f}, DoS={row['gossip_flow_dos_recall_mean']:.4f}, Gap={gap:.4f}")

# Verify Bandwidth Scaling with N
print("\n=== Bandwidth Scaling (k=8, q=4, f=3) ===")
bw_df = df[(df['k']==8) & (df['q']==4) & (df['f']==3)].sort_values('N')
for _, row in bw_df.iterrows():
    print(f"N={int(row['N'])}: Gossip_BW={row['gossip_bandwidth_mean']:.4f} KB/s, Cent_BW={row['cent_bandwidth_mean']:.4f} KB/s")

# Verify Latency Scaling with N
print("\n=== Latency Scaling (k=8, q=4, f=3) ===")
for _, row in bw_df.iterrows():
    print(f"N={int(row['N'])}: Gossip_Latency={row['gossip_latency_mean']:.4f}, Cent_Latency={row['cent_latency_mean']:.4f}")

