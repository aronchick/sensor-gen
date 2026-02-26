# sensor-gen

High-throughput IoT/OT sensor data generator simulating pipeline infrastructure (oil & gas, energy). Generates ~10,000+ JSON entries per second.

## Install

Installs both `sensor-gen` and `expanso-edge` (debug build):

```bash
curl -fsSL https://raw.githubusercontent.com/aronchick/sensor-gen/main/install.sh | sh
```

Install to a custom directory (no sudo needed):

```bash
curl -fsSL https://raw.githubusercontent.com/aronchick/sensor-gen/main/install.sh \
  | INSTALL_DIR=~/.local/bin sh
```

## Usage

```bash
# Default: output.jsonl at 10k entries/sec, runs until Ctrl+C
sensor-gen

# Run for 30 seconds with verbose stats
sensor-gen -d 30s -v

# Custom output file and rate
sensor-gen -o sensors.jsonl -rate 50000 -d 10s

# Append to existing file
sensor-gen -o sensors.jsonl --append -d 1m
```

## Sample Output

```json
{
  "sensor_id": "SNS-flo-2977",
  "timestamp": "2024-01-14T05:22:42.436377Z",
  "type": "flow_rate",
  "value": 12984.71,
  "unit": "bbl/hr",
  "location": { "lat": 37.82, "lon": -104.56, "mile_post": 203.7 },
  "pipeline_id": "PIPE-LA-001",
  "status": "normal",
  "quality_score": 0.87,
  "alert_level": "low"
}
```

## Sensor Types

| Type | Unit | Range |
|------|------|-------|
| pressure | psi | 200-1500 |
| temperature | fahrenheit | -20 to 180 |
| flow_rate | bbl/hr | 0-50000 |
| vibration | mm/s | 0-25 |
| corrosion | mpy | 0-50 |
| humidity | percent | 0-100 |
| gas_detector | ppm | 0-1000 |
| valve_position | percent | 0-100 |

## Building from Source

```bash
go build -o sensor-gen .
```

## License

MIT
