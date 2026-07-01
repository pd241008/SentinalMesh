from pathlib import Path
from typing import Optional

import numpy as np
import pandas as pd
from sklearn.preprocessing import StandardScaler

FEATURE_COLUMNS = ["sbytes", "dbytes", "spkts", "dpkts", "rate"]

CATEGORY_COLUMN = "attack_cat"
LABEL_COLUMN = "Label"
ID_COLUMN = "id"

KNOWN_ATTACK_CATEGORIES = [
    "analysis", "backdoor", "dos", "exploits", "fuzzers",
    "generic", "reconnaissance", "shellcode", "worms",
]


class FlowDataset:
    def __init__(self, flows: pd.DataFrame):
        self.flows = flows.reset_index(drop=True)
        self._label_attacks()

    @classmethod
    def load_csv(cls, path: str | Path) -> "FlowDataset":
        path = Path(path)
        if not path.exists():
            raise FileNotFoundError(f"dataset not found: {path}")

        df = pd.read_csv(path, low_memory=False)

        df.columns = df.columns.str.strip().str.lower()

        if ID_COLUMN in df.columns:
            df[ID_COLUMN] = df[ID_COLUMN].astype(int)

        return cls(df)

    def _label_attacks(self):
        cat_col = CATEGORY_COLUMN if CATEGORY_COLUMN in self.flows.columns else None
        if cat_col:
            self.flows[cat_col] = self.flows[cat_col].astype(str).str.strip().str.lower()
            self.flows["is_attack"] = ~self.flows[cat_col].isin(["normal", "", "nan"])
        else:
            label_col = LABEL_COLUMN if LABEL_COLUMN in self.flows.columns else None
            if label_col:
                self.flows["is_attack"] = self.flows[label_col].astype(int) == 1
            else:
                self.flows["is_attack"] = False

    @property
    def attack_flows(self) -> pd.DataFrame:
        return self.flows[self.flows["is_attack"]].copy()

    @property
    def normal_flows(self) -> pd.DataFrame:
        return self.flows[~self.flows["is_attack"]].copy()

    def get_features(self, source: Optional[pd.DataFrame] = None) -> np.ndarray:
        src = source if source is not None else self.flows
        available = [c for c in FEATURE_COLUMNS if c in src.columns]
        if not available:
            raise ValueError(f"no feature columns found in {list(src.columns)}")
        return src[available].fillna(0).to_numpy(dtype=np.float64)

    def get_labels(self, source: Optional[pd.DataFrame] = None) -> np.ndarray:
        src = source if source is not None else self.flows
        return src["is_attack"].to_numpy(dtype=bool)

    def get_categories(self, source: Optional[pd.DataFrame] = None) -> np.ndarray:
        src = source if source is not None else self.flows
        if CATEGORY_COLUMN in src.columns:
            return src[CATEGORY_COLUMN].to_numpy()
        return np.array(["unknown"] * len(src))

    def partition_by_node(self, num_nodes: int) -> list["FlowDataset"]:
        if ID_COLUMN not in self.flows.columns:
            return [self]

        partitions = []
        for node_id in range(num_nodes):
            node_flows = self.flows[self.flows[ID_COLUMN] % num_nodes == node_id].copy()
            partitions.append(FlowDataset(node_flows))
        return partitions

    def train_test_split(
        self, train_ratio: float = 0.8, random_state: int = 42
    ) -> tuple["FlowDataset", "FlowDataset"]:
        normal = self.normal_flows
        shuffled = normal.sample(frac=1, random_state=random_state)
        split_idx = int(len(shuffled) * train_ratio)
        train = FlowDataset(shuffled.iloc[:split_idx])
        test = FlowDataset(pd.concat([shuffled.iloc[split_idx:], self.attack_flows]))
        return train, test

    def normalize(self, scaler: Optional[StandardScaler] = None) -> tuple["FlowDataset", StandardScaler]:
        features = self.get_features()
        if scaler is None:
            scaler = StandardScaler()
            features_scaled = scaler.fit_transform(features)
        else:
            features_scaled = scaler.transform(features)

        df = self.flows.copy()
        available = [c for c in FEATURE_COLUMNS if c in df.columns]
        for i, col in enumerate(available):
            df[col] = features_scaled[:, i]

        return FlowDataset(df), scaler

    def __len__(self) -> int:
        return len(self.flows)

    def __repr__(self) -> str:
        n_attack = self.flows["is_attack"].sum()
        return f"FlowDataset({len(self)} flows, {n_attack} attacks)"
