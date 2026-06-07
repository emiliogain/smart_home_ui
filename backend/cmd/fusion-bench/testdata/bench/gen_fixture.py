#!/usr/bin/env python3
"""Generate a compact, realistic merged_timeline.csv + sensors_manifest.json fixture
for validating cmd/fusion-bench. Synthetic (user 99), NOT real HSLU data.

Phases exercise every context the models can emit, including the 90-lux boundary
case where the rule-based and fuzzy models are expected to disagree. Motion is
emitted as repeated PIR "on" events during active phases (real PIR fires
repeatedly while occupied) plus explicit clears, so presence decays realistically.
"""
import csv
import json
import os
from datetime import datetime, timedelta, timezone

OUT = os.path.join(os.path.dirname(__file__), "user_99")
os.makedirs(OUT, exist_ok=True)

START = datetime(2023, 6, 1, 8, 0, 0, tzinfo=timezone.utc)
rows = []  # (dt, stream, room, sensor, value, average_value)


def ts(minutes):
    return START + timedelta(minutes=minutes)


def periodic(minute, room, sensor, value):
    rows.append((ts(minute), "periodic", room, sensor, "", f"{value}"))


def motion(minute, room, on=True):
    rows.append((ts(minute), "event", room, "movement", "on" if on else "off", ""))


def env(minute, room, temp=None, hum=None, light=None):
    if temp is not None:
        periodic(minute, room, "temperature", temp)
    if hum is not None:
        periodic(minute, room, "humidity", hum)
    if light is not None:
        periodic(minute, room, "ambient_light", light)


def occupy(room, start, end, step=1):
    """Emit repeated PIR 'on' events while a room is occupied, then a clear."""
    m = start
    while m <= end:
        motion(m, room, True)
        m += step
    motion(end + 1, room, False)


# Phase 1 (08:00-08:30) comfortable: bright living room, person present.
env(0, "livingroom", temp=22.0, hum=45.0, light=500.0)
env(0, "kitchen", light=100.0)
env(0, "bedroom", light=40.0)
occupy("livingroom", 0, 29)

# Phase 2 (08:30-09:00) reading: moderate living light.
env(30, "livingroom", light=320.0)
occupy("livingroom", 30, 59)

# Phase 3 (09:00-09:30) watching tv: dim living light.
env(60, "livingroom", light=50.0)
occupy("livingroom", 60, 89)

# Phase 4 (09:30-10:00) the 90-lux boundary case (rule=TV, fuzzy=comfortable).
env(90, "livingroom", light=90.0)
occupy("livingroom", 90, 119)

# Phase 5 (10:00-11:00) cooking: kitchen presence, warm + humid + bright kitchen.
env(120, "livingroom", temp=24.0, hum=58.0, light=200.0)
env(120, "kitchen", light=450.0)
occupy("kitchen", 120, 179)

# Phase 6 (11:00-12:00) away: no motion anywhere, lights low.
env(180, "livingroom", light=5.0)
env(180, "kitchen", light=5.0)
env(180, "bedroom", light=5.0)
# (no motion events -> presence decays -> NO_ONE_HOME)

# Phase 7 (12:00-12:30) overheating alert.
env(240, "livingroom", temp=29.5, light=400.0)
occupy("livingroom", 240, 269)

# Phase 8 (12:30-13:30) comfortable again.
env(270, "livingroom", temp=22.0, hum=45.0, light=500.0)
occupy("livingroom", 270, 329)

# Phase 9 (13:30-15:00) sleeping: last motion bedroom, then dark + cool + still.
# Periodic sensors keep reporting every 10 min during the still period (as real
# periodic streams do), so ticks continue and presence decays -> SLEEPING / away.
env(330, "livingroom", temp=18.0, light=2.0)
env(330, "kitchen", light=0.0)
env(330, "bedroom", light=2.0)
occupy("bedroom", 330, 334)  # brief bedroom motion, then stillness until 15:00
for m in range(340, 421, 10):
    env(m, "livingroom", temp=18.0, light=2.0)
    env(m, "bedroom", light=2.0)

rows.sort(key=lambda r: r[0])

with open(os.path.join(OUT, "merged_timeline.csv"), "w", newline="") as f:
    w = csv.writer(f)
    w.writerow(["datetime_utc", "stream", "id", "country", "room", "sensor", "value", "average_value"])
    for dt, stream, room, sensor, value, avg in rows:
        w.writerow([dt.strftime("%Y-%m-%dT%H:%M:%S.000000000"), stream, "99", "XX", room, sensor, value, avg])

manifest = {
    "sensors": [
        {"name": "temp_living_room", "type": "temperature", "location": "living_room"},
        {"name": "humidity_living_room", "type": "humidity", "location": "living_room"},
        {"name": "light_living_room", "type": "light", "location": "living_room"},
        {"name": "motion_living_room", "type": "motion", "location": "living_room"},
        {"name": "light_kitchen", "type": "light", "location": "kitchen"},
        {"name": "motion_kitchen", "type": "motion", "location": "kitchen"},
        {"name": "light_bedroom", "type": "light", "location": "bedroom"},
        {"name": "motion_bedroom", "type": "motion", "location": "bedroom"},
    ],
    "counts_by_type": {"temperature": 1, "humidity": 1, "light": 3, "motion": 3},
}
with open(os.path.join(OUT, "sensors_manifest.json"), "w") as f:
    json.dump(manifest, f, indent=2)

print(f"wrote {len(rows)} rows to {OUT}/merged_timeline.csv")
