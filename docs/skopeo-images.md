# Inspecting images with Skopeo

## Current behavior

**Note:** With `hpn pull --runtime skopeo`, images are written to a temporary directory that is removed when the command finishes, so you cannot inspect them afterward. Use `hpn save --runtime skopeo` to persist images.

## How Skopeo stores images

Skopeo is **daemonless**. It does not keep an image store like Docker or Podman. The `copy` command fetches an image and writes it to a chosen destination.

### Supported destination types

1. **`dir:`** — Directory (OCI layout)
2. **`docker-archive:`** — Docker archive (tar file)
3. **`docker-daemon:`** — Docker daemon
4. **`containers-storage:`** — Podman/CRI-O storage
5. **`oci:`** — OCI image layout
6. **`ostree:`** — OSTree storage

## Inspecting images

### Method 1: `skopeo inspect` (metadata)

#### Remote (registry)

```bash
skopeo inspect docker://docker.io/library/nginx:latest
skopeo inspect --override-arch amd64 docker://docker.io/library/nginx:latest
```

#### Local directory (`dir:`)

```bash
skopeo inspect dir:/path/to/image-dir
skopeo inspect dir:/tmp/skopeo-pull-nginx_latest
```

#### Docker daemon

```bash
skopeo inspect docker-daemon:nginx:latest
docker images | grep nginx
```

#### Containers storage (e.g. Podman)

```bash
skopeo inspect containers-storage:nginx:latest
podman images | grep nginx
```

#### Archive (tar file)

```bash
skopeo inspect docker-archive:/path/to/image.tar
skopeo inspect docker-archive:./images/nginx_latest.tar
```

### Method 2: `skopeo list-tags`

```bash
skopeo list-tags docker://docker.io/library/nginx
skopeo list-tags docker://registry.example.com/myproject/myapp
```

### Method 3: Copy into Docker/Podman to inspect

```bash
# From dir to Docker
skopeo copy dir:/path/to/image-dir docker-daemon:nginx:latest
docker images | grep nginx
docker inspect nginx:latest

# From dir to Podman
skopeo copy dir:/path/to/image-dir containers-storage:localhost/nginx:latest
podman images | grep nginx
podman inspect nginx:latest
```

### Method 4: Persist with `skopeo copy` then inspect

```bash
# Save as directory (easy to inspect)
skopeo copy docker://docker.io/library/nginx:latest dir:./images/nginx-latest
skopeo inspect dir:./images/nginx-latest

# Save as tar
skopeo copy docker://docker.io/library/nginx:latest docker-archive:./images/nginx-latest.tar
skopeo inspect docker-archive:./images/nginx-latest.tar
```

## Limitation of `hpn pull --runtime skopeo`

Current implementation:

- Writes to a temp dir `/tmp/skopeo-pull-*`
- Removes that dir when the function returns (`defer os.RemoveAll(tempDir)`)

So you cannot inspect or reuse the image after pull. To persist images, use **`hpn save --runtime skopeo`** (or run `skopeo copy` yourself to a `dir:` or `docker-archive:` path).

## Command reference

### Inspect

```bash
skopeo inspect docker://docker.io/library/nginx:latest
skopeo inspect dir:./images/nginx-latest
skopeo inspect docker-daemon:nginx:latest
skopeo inspect docker-archive:./images/nginx-latest.tar
```

### List tags

```bash
skopeo list-tags docker://docker.io/library/nginx
skopeo list-tags docker://registry.example.com/myproject/myapp
```

### Copy between storages

```bash
skopeo copy docker://nginx:latest dir:./nginx-latest
skopeo copy dir:./nginx-latest docker-daemon:nginx:latest
skopeo copy dir:./nginx-latest containers-storage:localhost/nginx:latest
skopeo copy docker://nginx:latest docker-archive:./nginx-latest.tar
```

### Remove images

```bash
# Dir layout
rm -rf ./images/nginx-latest

# Docker
docker rmi nginx:latest

# Podman
podman rmi nginx:latest
```

## Best practices

1. **Use `hpn save` to persist:** `hpn save --runtime skopeo -f images.txt --path ./images`
2. **Inspect archives:** `skopeo inspect docker-archive:./images/nginx-latest.tar`
3. **Import to Docker/Podman when needed:** `skopeo copy dir:./images/nginx-latest docker-daemon:nginx:latest`
4. **Load with hpn:** `hpn load --path ./images`

## Cleanup

### `hpn pull` (Skopeo)

Images are written to `/tmp/skopeo-pull-*` and removed automatically. No manual cleanup.

### Manually saved images

If you used `skopeo copy` or `hpn save`, clean up as needed:

- **Dir:** `rm -rf ./images/nginx-latest` or `rm -rf ./images/*/`
- **Tar:** `rm -f ./images/*.tar ./images/*.tar.sha256`
- **Docker:** `docker rmi <image>` or `docker image prune -a`
- **Podman:** `podman rmi <image>` or `podman image prune -a`

### Temp dirs

```bash
ls -la /tmp/skopeo-pull-*
rm -rf /tmp/skopeo-pull-*
```

### Disk usage

```bash
du -sh ./images/
docker system df
podman system df
```

## Summary

- **`hpn pull --runtime skopeo`:** Temp dir only; auto-cleaned; not inspectable afterward.
- **Persist and inspect:** Use `hpn save` or `skopeo copy` to `dir:` or `docker-archive:`.
- **Cleanup:** Dir/tar → `rm`; Docker/Podman → `rmi` / `image prune`.
