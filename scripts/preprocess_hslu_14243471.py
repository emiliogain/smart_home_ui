#!/usr/bin/env python3
"""
Preprocess HSLU-style raw exports (event_data + periodic_data_monthly_csv) for replay.

Reads:
  - event_data.csv  (long format: datetime_utc, id, country, room, sensor, value)
  - periodic_data_monthly_csv/periodic_data_*.csv
        (long format with min_value, average_value, max_value)

Writes under OUTPUT_ROOT one folder per user:
  user_<id>/events.csv
      movement only; rooms unsupported by the app (e.g. bathroom, unknown) dropped;
      movement rows restricted to value on/off (matches replay apply rules).
  user_<id>/periodic_data_merged.csv
      temperature, humidity, ambient_light only; unsupported rooms dropped;
      rows with empty average_value dropped.
  user_<id>/sensors_manifest.json
      Sensor definitions derived from that user's rows (names/types align with the
      backend + hsludata.Apply*). Includes counts_by_type (distinct channels per type).
  user_<id>/merged_timeline.csv
      That user's events and periodic rows in one file, sorted by datetime_utc.
      stream=event uses column value; stream=periodic uses average_value only (no min/max).

Room mapping matches backend/internal/hsludata/rooms.go (NormalizeDatasetRoom).
Supported event sensor: movement. Supported periodic: temperature, humidity, ambient_light.
(CO2, VOC, sound_*, door, etc. are not written.)

Monthly periodic files are concatenated in filename order (periodic_data_YYYY_MM.csv).

Usage:
  python3 scripts/preprocess_hslu_14243471.py \\
    --events ~/Downloads/event_data.csv \\
    --periodic-dir ~/Downloads/periodic_data_monthly_csv \\
    --output-dir ./datasets/hslu_processed

Replay one user (from repo root, backend/; requires merged_timeline.csv in that folder):
  go run ./cmd/hslu14243471-replay -data-dir ../datasets/hslu_processed/user_7 -user-id 7
"""

from __future__ import annotations

import argparse
import csv
import json
import os
import sys
from collections import defaultdict
from dataclasses import dataclass, field
from datetime import datetime, timezone
from typing import DefaultDict, Dict, Iterable, List, Set, TextIO, Tuple

# Match internal/hsludata/apply.go (event + periodic filters).
EVENT_SENSORS = frozenset({"movement"})
PERIODIC_SENSORS = frozenset({"temperature", "humidity", "ambient_light"})


def normalize_dataset_room(room: str) -> Tuple[str | None, bool]:
    """Match hsludata.NormalizeDatasetRoom: (location, skip)."""
    r = (room or "").strip().lower()
    if r == "livingroom":
        return "living_room", False
    if r == "kitchen":
        return "kitchen", False
    if r == "bedroom":
        return "bedroom", False
    if r == "general":
        return "living_room", False
    if r == "bathroom":
        return None, True
    return None, True


def periodic_to_registration(loc: str, sensor: str) -> Tuple[str, str, str]:
    """CSV periodic sensor -> (api_name, api_type, location). Matches simulator.DefaultSensors."""
    s = sensor.strip().lower()
    if s == "temperature":
        return f"temp_{loc}", "temperature", loc
    if s == "humidity":
        return f"humidity_{loc}", "humidity", loc
    if s == "ambient_light":
        return f"light_{loc}", "light", loc
    raise ValueError(f"unsupported periodic sensor {sensor!r}")


@dataclass
class UserInventory:
    event_rows: int = 0
    periodic_rows: int = 0
    motion_rooms: Set[str] = field(default_factory=set)
    # (location, periodic sensor key) e.g. ("living_room", "temperature")
    periodic_channels: Set[Tuple[str, str]] = field(default_factory=set)


def build_manifest_sensors(inv: UserInventory) -> List[Dict[str, str]]:
    by_name: Dict[str, Dict[str, str]] = {}
    for loc, s in sorted(inv.periodic_channels):
        name, typ, location = periodic_to_registration(loc, s)
        by_name[name] = {"name": name, "type": typ, "location": location}
    for loc in sorted(inv.motion_rooms):
        name = f"motion_{loc}"
        if name not in by_name:
            by_name[name] = {"name": name, "type": "motion", "location": loc}
    return [by_name[k] for k in sorted(by_name.keys())]


def counts_by_type(sensors: List[Dict[str, str]]) -> Dict[str, int]:
    out: Dict[str, int] = defaultdict(int)
    for s in sensors:
        out[s["type"]] += 1
    return dict(sorted(out.items()))


def write_sensors_manifest(user_dir: str, inv: UserInventory) -> None:
    sensors = build_manifest_sensors(inv)
    payload = {
        "sensors": sensors,
        "counts_by_type": counts_by_type(sensors),
    }
    path = os.path.join(user_dir, "sensors_manifest.json")
    with open(path, "w", encoding="utf-8") as f:
        json.dump(payload, f, indent=2)


def partition_events(
    events_path: str,
    out_root: str,
    inventory: DefaultDict[str, UserInventory],
) -> Dict[str, int]:
    counts: Dict[str, int] = {}
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
            room_raw = row.get("room") or ""
            loc, skip = normalize_dataset_room(room_raw)
            if skip or not loc:
                continue
            val = (row.get("value") or "").strip().lower()
            if val not in ("on", "off"):
                continue
            get_writer(uid).writerow(row)
            inv = inventory[uid]
            inv.event_rows += 1
            inv.motion_rooms.add(loc)
            counts[uid] = counts.get(uid, 0) + 1

    for fp, _ in writers.values():
        fp.close()
    return counts


def parse_datetime_utc_sort_key(s: str) -> datetime:
    """Parse dataset timestamps for ordering (sub-microsecond fractions truncated to 6 digits)."""
    raw = (s or "").strip()
    if not raw:
        return datetime.min.replace(tzinfo=timezone.utc)
    if raw.endswith("Z"):
        raw = raw[:-1] + "+00:00"
    if "." in raw and "T" in raw:
        prefix, suffix = raw.split(".", 1)
        digits = []
        tz = ""
        for i, ch in enumerate(suffix):
            if ch.isdigit():
                digits.append(ch)
            else:
                tz = suffix[i:]
                break
        frac = ("".join(digits) + "000000")[:6]
        raw = f"{prefix}.{frac}{tz}"
    dt = datetime.fromisoformat(raw)
    if dt.tzinfo is None:
        dt = dt.replace(tzinfo=timezone.utc)
    return dt


MERGED_FIELDNAMES = (
    "datetime_utc",
    "stream",
    "id",
    "country",
    "room",
    "sensor",
    "value",
    "average_value",
)


def write_user_merged_timeline(user_dir: str) -> Tuple[int, int, int]:
    """
    Combine user_dir/events.csv and user_dir/periodic_data_merged.csv for the same
    participant into user_dir/merged_timeline.csv sorted by datetime_utc.
    Periodic rows carry average_value only; event rows use value.
    Missing events or periodic file is treated as empty. Returns (ev_n, per_n, total_rows).
    """
    ev_path = os.path.join(user_dir, "events.csv")
    per_path = os.path.join(user_dir, "periodic_data_merged.csv")
    rows: List[Tuple[datetime, int, str, Dict[str, str]]] = []
    ev_n = 0

    if os.path.isfile(ev_path):
        with open(ev_path, "r", newline="", encoding="utf-8-sig") as inf:
            r = csv.DictReader(inf)
            if not r.fieldnames or "datetime_utc" not in r.fieldnames:
                raise SystemExit(f"{ev_path}: missing datetime_utc header")
            for row in r:
                dt_s = (row.get("datetime_utc") or "").strip()
                try:
                    dt = parse_datetime_utc_sort_key(dt_s)
                except ValueError:
                    continue
                out = {
                    "datetime_utc": dt_s,
                    "stream": "event",
                    "id": (row.get("id") or "").strip(),
                    "country": (row.get("country") or "").strip(),
                    "room": (row.get("room") or "").strip(),
                    "sensor": (row.get("sensor") or "").strip(),
                    "value": (row.get("value") or "").strip(),
                    "average_value": "",
                }
                rows.append((dt, 0, dt_s, out))
                ev_n += 1

    per_n = 0
    if os.path.isfile(per_path):
        with open(per_path, "r", newline="", encoding="utf-8-sig") as inf:
            r = csv.DictReader(inf)
            if not r.fieldnames or "datetime_utc" not in r.fieldnames:
                raise SystemExit(f"{per_path}: missing datetime_utc header")
            if "average_value" not in r.fieldnames:
                raise SystemExit(f"{per_path}: missing average_value column")
            for row in r:
                dt_s = (row.get("datetime_utc") or "").strip()
                try:
                    dt = parse_datetime_utc_sort_key(dt_s)
                except ValueError:
                    continue
                avg = (row.get("average_value") or "").strip()
                if not avg:
                    continue
                out = {
                    "datetime_utc": dt_s,
                    "stream": "periodic",
                    "id": (row.get("id") or "").strip(),
                    "country": (row.get("country") or "").strip(),
                    "room": (row.get("room") or "").strip(),
                    "sensor": (row.get("sensor") or "").strip(),
                    "value": "",
                    "average_value": avg,
                }
                rows.append((dt, 1, dt_s, out))
                per_n += 1

    rows.sort(key=lambda t: (t[0], t[1], t[2], t[3]["id"], t[3]["room"], t[3]["sensor"]))
    out_path = os.path.join(user_dir, "merged_timeline.csv")
    with open(out_path, "w", newline="", encoding="utf-8") as outf:
        w = csv.DictWriter(outf, fieldnames=list(MERGED_FIELDNAMES), extrasaction="ignore")
        w.writeheader()
        for _, __, ___, rec in rows:
            w.writerow(rec)

    return ev_n, per_n, len(rows)


def merge_periodic(
    periodic_dir: str,
    out_root: str,
    inventory: DefaultDict[str, UserInventory],
) -> Dict[str, int]:
    counts: Dict[str, int] = {}
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
                room_raw = row.get("room") or ""
                loc, skip = normalize_dataset_room(room_raw)
                if skip or not loc:
                    continue
                avg = (row.get("average_value") or "").strip()
                if not avg:
                    continue
                get_writer(uid).writerow(row)
                inv = inventory[uid]
                inv.periodic_rows += 1
                inv.periodic_channels.add((loc, sensor))
                counts[uid] = counts.get(uid, 0) + 1

    for fp, _ in writers.values():
        fp.close()
    return counts


def main(argv: Iterable[str]) -> int:
    p = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    p.add_argument("--events", required=True, help="Path to raw event_data.csv")
    p.add_argument("--periodic-dir", required=True, help="Directory with raw periodic_data_*.csv")
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

    inventory: DefaultDict[str, UserInventory] = defaultdict(UserInventory)

    print("Partitioning events (movement, supported rooms, on/off only)…")
    ev_counts = partition_events(events_path, out_root, inventory)
    print(f"  users with events: {len(ev_counts)}  total rows: {sum(ev_counts.values())}")

    print("Merging periodic streams (supported sensors + rooms only)…")
    per_counts = merge_periodic(periodic_dir, out_root, inventory)
    print(f"  users with periodic: {len(per_counts)}  total rows: {sum(per_counts.values())}")

    all_users = sorted(set(ev_counts) | set(per_counts), key=lambda x: int(x) if x.isdigit() else x)
    merged_timeline_total: Dict[str, int] = {}
    for uid in all_users:
        user_dir = os.path.join(out_root, f"user_{uid}")
        if os.path.isdir(user_dir):
            write_sensors_manifest(user_dir, inventory[uid])
            me, mp, mt = write_user_merged_timeline(user_dir)
            merged_timeline_total[uid] = mt
            print(f"  user_{uid}: merged_timeline.csv ({me} event + {mp} periodic = {mt} rows)")

    summary = {
        "output_dir": out_root,
        "event_sensors_kept": sorted(EVENT_SENSORS),
        "periodic_sensors_kept": sorted(PERIODIC_SENSORS),
        "rooms_note": "bathroom and unknown rooms dropped; general -> living_room (see hsludata/rooms.go)",
        "users": {
            uid: {
                "events_rows": ev_counts.get(uid, 0),
                "periodic_rows": per_counts.get(uid, 0),
                "merged_timeline_csv": os.path.join(out_root, f"user_{uid}", "merged_timeline.csv"),
                "merged_timeline_rows": merged_timeline_total.get(uid, 0),
                "counts_by_type": counts_by_type(build_manifest_sensors(inventory[uid])),
                "sensor_channels": len(build_manifest_sensors(inventory[uid])),
                "replay_hint": (
                    f"go run ./cmd/hslu14243471-replay -data-dir "
                    f"{os.path.join(out_root, 'user_' + uid)} -user-id {uid}"
                ),
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
