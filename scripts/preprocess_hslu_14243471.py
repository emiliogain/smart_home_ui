#!/usr/bin/env python3
"""
Preprocess HSLU-style raw exports (event_data + periodic_data_monthly_csv) for replay.

Reads:
  - event_data.csv  (long format: datetime_utc, id, country, room, sensor, value)
  - periodic_data_monthly_csv/periodic_data_*.csv
        (long format with min_value, average_value, max_value)

Writes under OUTPUT_ROOT one folder per user:
  user_<id>/events.csv
      Same schema as the source; only rows with sensor == "movement" (door etc. dropped).
  user_<id>/periodic_data_merged.csv
      Same schema as periodic sources; only sensors used by the app:
      temperature, humidity, ambient_light (CO2, VOC, sound_* dropped).

Monthly periodic files are concatenated in filename order (periodic_data_YYYY_MM.csv).

Usage:
  python3 scripts/preprocess_hslu_14243471.py \\
    --events ~/Downloads/event_data.csv \\
    --periodic-dir ~/Downloads/periodic_data_monthly_csv \\
    --output-dir ./datasets/hslu_processed

Replay one user (from repo root, backend/):
  go run ./cmd/hslu14243471-replay -data-dir ../datasets/hslu_processed/user_7 -user-id 7 -playback 1
  # Real-time vs dataset timestamps (default). Fixed UI cadence every 10s: add -every 10
"""

from __future__ import annotations

import argparse
import csv
import json
import os
import sys
from collections import defaultdict
from typing import Dict, Iterable, List, TextIO, Tuple

# Match internal/hsludata/apply.go (event + periodic filters).
EVENT_SENSORS = frozenset({"movement"})
PERIODIC_SENSORS = frozenset({"temperature", "humidity", "ambient_light"})


def partition_events(events_path: str, out_root: str) -> Dict[str, int]:
    counts: Dict[str, int] = defaultdict(int)
    writers: Dict[str, Tuple[TextIO, csv.DictWriter]] = {}
    fieldnames: List[str] | None = None

    def get_writer(uid: str) -> csv.DictWriter:
        if uid not in writers:
            d = os.path.join(out_root, f"user_{uid}")
            os.makedirs(d, exist_ok=True)
            fp = open(os.path.join(d, "events.csv"), "w", newline="", encoding="utf-8")
            assert fieldnames is not None
            w = csv.DictWriter(fp, fieldnames=fieldnames, extrasaction="ignore")
            w.writeheader()
            writers[uid] = (fp, w)
        return writers[uid][1]

    with open(events_path, "r", newline="", encoding="utf-8-sig") as inf:
        r = csv.DictReader(inf)
        if not r.fieldnames:
            raise SystemExit("events: missing header")
        fieldnames = list(r.fieldnames)
        for row in r:
            sensor = (row.get("sensor") or "").strip().lower()
            if sensor not in EVENT_SENSORS:
                continue
            uid = (row.get("id") or "").strip()
            if not uid:
                continue
            get_writer(uid).writerow(row)
            counts[uid] += 1

    for fp, _ in writers.values():
        fp.close()
    return dict(counts)


def merge_periodic(periodic_dir: str, out_root: str) -> Dict[str, int]:
    counts: Dict[str, int] = defaultdict(int)
    writers: Dict[str, Tuple[TextIO, csv.DictWriter]] = {}
    fieldnames: List[str] | None = None

    def get_writer(uid: str) -> csv.DictWriter:
        if uid not in writers:
            d = os.path.join(out_root, f"user_{uid}")
            os.makedirs(d, exist_ok=True)
            fp = open(os.path.join(d, "periodic_data_merged.csv"), "w", newline="", encoding="utf-8")
            assert fieldnames is not None
            w = csv.DictWriter(fp, fieldnames=fieldnames, extrasaction="ignore")
            w.writeheader()
            writers[uid] = (fp, w)
        return writers[uid][1]

    names = sorted(
        f
        for f in os.listdir(periodic_dir)
        if f.startswith("periodic_data_")
        and f.endswith(".csv")
        and f != "periodic_data_merged.csv"
    )
    if not names:
        raise SystemExit(f"no periodic_data_*.csv in {periodic_dir!r}")

    for name in names:
        path = os.path.join(periodic_dir, name)
        with open(path, "r", newline="", encoding="utf-8-sig") as inf:
            r = csv.DictReader(inf)
            if not r.fieldnames:
                continue
            if fieldnames is None:
                fieldnames = list(r.fieldnames)
            for row in r:
                sensor = (row.get("sensor") or "").strip().lower()
                if sensor not in PERIODIC_SENSORS:
                    continue
                uid = (row.get("id") or "").strip()
                if not uid:
                    continue
                get_writer(uid).writerow(row)
                counts[uid] += 1

    for fp, _ in writers.values():
        fp.close()
    return dict(counts)


def main(argv: Iterable[str]) -> int:
    p = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    p.add_argument("--events", required=True, help="Path to event_data.csv")
    p.add_argument("--periodic-dir", required=True, help="Directory containing periodic_data_*.csv")
    p.add_argument(
        "--output-dir",
        default="datasets/hslu_processed",
        help="Root folder for user_<id>/ outputs (default: ./datasets/hslu_processed)",
    )
    args = p.parse_args(list(argv))

    events_path = os.path.expanduser(args.events)
    periodic_dir = os.path.expanduser(args.periodic_dir)
    out_root = os.path.abspath(os.path.expanduser(args.output_dir))

    if not os.path.isfile(events_path):
        print(f"not found: {events_path}", file=sys.stderr)
        return 1
    if not os.path.isdir(periodic_dir):
        print(f"not a directory: {periodic_dir}", file=sys.stderr)
        return 1

    os.makedirs(out_root, exist_ok=True)
    if os.path.abspath(periodic_dir) == os.path.abspath(out_root):
        print("--output-dir must not be the same as --periodic-dir", file=sys.stderr)
        return 1
    print(f"Writing processed data to {out_root}")

    print("Partitioning events (movement only)…")
    ev_counts = partition_events(events_path, out_root)
    print(f"  users with events: {len(ev_counts)}  total rows: {sum(ev_counts.values())}")

    print("Merging periodic streams (temperature, humidity, ambient_light only)…")
    per_counts = merge_periodic(periodic_dir, out_root)
    print(f"  users with periodic: {len(per_counts)}  total rows: {sum(per_counts.values())}")

    all_users = sorted(set(ev_counts) | set(per_counts), key=lambda x: int(x) if x.isdigit() else x)
    summary = {
        "output_dir": out_root,
        "event_sensors_kept": sorted(EVENT_SENSORS),
        "periodic_sensors_kept": sorted(PERIODIC_SENSORS),
        "users": {
            uid: {
                "events_rows": ev_counts.get(uid, 0),
                "periodic_rows": per_counts.get(uid, 0),
                "replay_hint": f"go run ./cmd/hslu14243471-replay -data-dir {os.path.join(out_root, 'user_' + uid)} -user-id {uid} -playback 1",
            }
            for uid in all_users
        },
    }
    summary_path = os.path.join(out_root, "summary.json")
    with open(summary_path, "w", encoding="utf-8") as sf:
        json.dump(summary, sf, indent=2)
    print(f"Wrote {summary_path}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv[1:]))
