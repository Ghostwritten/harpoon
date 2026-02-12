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

### CPU

- **Pull:** Medium (layer decompression, verification)
- **Save:** High (compression, checksums)
- **Load:** Medium (tar decompression, import)
- **Push:** Low–medium (compression for upload)

**Tip:** For CPU-heavy operations (Save), use workers ≈ CPU cores. For I/O-heavy (Pull/Load/Push), you can use more.

### Memory

- Each worker uses memory for image data.
- Larger images use more memory (roughly 100–500 MB per task).

**Tip:** Increase workers when memory is plentiful; use 2–3 when it is limited.

### Network

- **Pull / Push:** High usage.
- **Save / Load:** No network (local only).

**Tip:** Use higher concurrency (e.g. 10–20) on fast networks; 3–5 on slow or limited networks.

### Disk I/O

- **Save:** High (writing tar).
- **Load:** High (reading tar).
- **Pull:** Medium (writing layers).
- **Push:** Low (reading layers).

**Tip:** SSDs can handle more workers (e.g. 10–15); HDDs work better with fewer (e.g. 3–5).

### Runtime limits

Docker/Podman daemons have their own limits. Typically 5–10 workers is safe.

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

All operations report total execution time, for example:

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

## Examples

### Batch pull (high concurrency)

```bash
hpn pull -f images.txt --workers 10
```

### Batch save (moderate concurrency)

```bash
hpn save -f images.txt --path ./backup --workers 5
```

### Batch push

```bash
hpn push -f images.txt --registry new-registry.com --workers 8
```

## Notes

1. Avoid very high `--workers`; it can exhaust CPU, memory, or disk I/O.
2. Monitor with tools like `htop` and `iostat`.
3. Pull/Push are limited by network bandwidth.
4. Some runtimes have their own concurrency limits.
5. Save needs enough free disk space.

## Troubleshooting

- **System slow or OOM:** Lower `--workers` (e.g. `--workers 2`).
- **Network timeouts:** Lower concurrency, check connectivity, or increase timeout in config.
- **Tuning:** Try different `--workers` values and compare total time; set a default in config if desired.

## Optimization

- **Pull/Push (network-bound):** Try higher concurrency (e.g. 10–15).
- **Save (CPU/disk-bound):** Use moderate concurrency (e.g. 5–8).
- **Load (disk-bound):** Try 8–10 workers.
- Adjust for CPU count, memory, and storage (SSD vs HDD).
