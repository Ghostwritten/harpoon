# Concurrency

## Overview

Harpoon (hpn) supports concurrent execution for pull, save, load, and push, which can significantly speed up batch operations.

## Configuration

### 1. Command line (recommended)

Each subcommand supports `--workers`:

```bash
hpn pull -f images.txt --workers 10
hpn save -f images.txt --path ./images --workers 8
hpn load --path ./images --workers 5
hpn push -f images.txt --registry new-registry.com --workers 6
```

### 2. Config file

In `~/.hpn/config.yaml`:

```yaml
parallel:
  max_workers: 5     # default concurrency
  auto_adjust: true  # future feature
```

### 3. Default

If `--workers` is not set and there is no config, the default is **5** workers.

### Priority

1. **`--workers`** (highest)
2. **`parallel.max_workers`** in config
3. **5** (lowest)

## Resource usage

### Summary table

| Operation | CPU | Memory | Network | Disk I/O | Suggested workers |
|-----------|-----|--------|---------|----------|--------------------|
| Pull      | M   | M      | H       | M        | 5–10               |
| Save      | H   | M      | —       | H        | 3–8                |
| Load      | M   | M      | —       | H        | 5–10               |
| Push      | L–M | L      | H       | L        | 5–15               |

### Choosing `--workers`

- **Low-end (2–4 cores, 4–8 GB RAM):** `--workers 2` or `--workers 3`
- **Mid (4–8 cores, 8–16 GB RAM):** `--workers 5` (default) or `--workers 8`
- **High (8+ cores, 16+ GB RAM, SSD):** `--workers 10` or `--workers 15`
- **Limited network:** `--workers 3`
- **HDD / slow disk:** `--workers 3`

## Execution time

All operations report total execution time:

```bash
$ hpn pull -f images.txt --workers 5
...
Summary: 10 successful, 0 failed
Total time: 2m15s
```

## How it works

Operations use a **worker pool**:

1. A fixed number of workers (goroutines) are started.
2. Tasks are queued and workers take tasks from the queue.
3. Results are collected; one failure does not stop others.
4. Progress is shown per task.

## Troubleshooting

- **System slow or OOM:** Lower `--workers` (e.g. `--workers 2`).
- **Network timeouts:** Lower concurrency, check connectivity, or increase timeout in config.
- **Tuning:** Try different `--workers` values and compare total time.
