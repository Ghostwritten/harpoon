# Skopeo: Pull vs Save

## Concepts

**Note:** Skopeo has no separate `pull` command; only `copy`. `skopeo copy` both fetches and writes the image to a destination.

## hpn pull vs hpn save with Skopeo

### hpn pull --runtime skopeo

**Command executed:**
```bash
skopeo copy --preserve-digests [--all] docker://image dir:/tmp/skopeo-pull-*
```

**Behavior:**
- **Destination format:** `dir:` (OCI layout)
- **Destination path:** `/tmp/skopeo-pull-*` (temporary directory)
- **Cleanup:** Directory is removed when the function returns
- **Purpose:** Verify the image can be pulled; no persistent storage
- **Result:** Image is written to a temp dir, then that dir is deleted; **you cannot inspect it later**

### hpn save --runtime skopeo

**Command executed:**
```bash
skopeo copy --preserve-digests [--all] docker://image docker-archive:./images/nginx.tar
```

**Behavior:**
- **Destination format:** `docker-archive:` (tar file)
- **Destination path:** User-specified (e.g. `./images/nginx.tar`)
- **Persistence:** Files are kept; **not deleted**
- **Purpose:** Download and save the image as a tar
- **Result:** Image is stored as a tar file; **you can inspect and reuse it**

## Comparison

| Operation   | Destination format   | Location      | Persistent | Inspectable |
|------------|----------------------|---------------|------------|-------------|
| `hpn pull` | `dir:`               | Temp directory| No         | No          |
| `hpn save` | `docker-archive:`    | User path     | Yes        | Yes         |

## Why does hpn pull not keep the image?

1. **Consistency with other runtimes:** Docker/Podman "pull" stores in the local daemon; Skopeo has no daemon.
2. **Pull = "can we fetch it?":** Pull is intended to verify the image is reachable, not to persist it.
3. **Use Save to persist:** To keep the image, use `hpn save`.

## Usage

### Only verify images are pullable

```bash
hpn pull --runtime skopeo -f images.txt
# Images are downloaded to a temp dir and then removed.
```

### Download and save as tar

```bash
hpn save --runtime skopeo -f images.txt --path ./images
# Inspect: skopeo inspect docker-archive:./images/nginx.tar
```

### Save as directory format (dir:)

**Option 1: Use skopeo directly**
```bash
skopeo copy --preserve-digests --all docker://nginx:latest dir:./images/nginx-latest
skopeo inspect dir:./images/nginx-latest
```

**Option 2: Save with hpn then convert**
```bash
hpn save --runtime skopeo -f images.txt --path ./images
skopeo copy docker-archive:./images/nginx.tar dir:./images/nginx-latest
```

## Summary

1. Skopeo has no standalone "pull"; only `copy`, which both fetches and writes.
2. **hpn pull:** Uses `dir:` to a temp directory, then deletes it (verification only).
3. **hpn save:** Uses `docker-archive:` to the path you specify (persistent).
4. To keep images when using Skopeo, use `hpn save`, not `hpn pull`.

## See Also

- [Skopeo: Images](/guide/advanced/skopeo-images) - Inspecting and cleaning Skopeo images
- [Examples](/guide/examples) - Real-world usage
