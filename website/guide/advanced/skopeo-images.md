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

## Limitation of `hpn pull --runtime skopeo`

Current implementation:

- Writes to a temp dir `/tmp/skopeo-pull-*`
- Removes that dir when the function returns

So you cannot inspect or reuse the image after pull. To persist images, use **`hpn save --runtime skopeo`** (or run `skopeo copy` yourself to a `dir:` or `docker-archive:` path).

## Best practices

1. **Use `hpn save` to persist:** `hpn save --runtime skopeo -f images.txt --path ./images`
2. **Inspect archives:** `skopeo inspect docker-archive:./images/nginx-latest.tar`
3. **Import to Docker/Podman when needed:** `skopeo copy dir:./images/nginx-latest docker-daemon:nginx:latest`
4. **Load with hpn:** `hpn load --path ./images`

## See Also

- [Skopeo: Pull vs Save](/guide/advanced/skopeo-pull-vs-save) - Pull vs save behavior
- [Examples](/guide/examples) - Real-world usage
