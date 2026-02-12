# Proxy environment variable behavior

## Summary

### 1. Default behavior (proxy not enabled in config)

**Code**: `cmd/hpn/root.go` (proxy only applied when `cfg.Proxy.Enabled` is true)

**Runtime implementation**: `internal/runtime/docker.go` (and similarly for other runtimes)

- When `options.Proxy == nil` or `options.Proxy.Enabled == false`, the code does **not** set `cmd.Env`.
- Go’s `exec.Command` then **inherits the parent process’s environment**.
- So if `http_proxy`, `https_proxy`, `HTTP_PROXY`, or `HTTPS_PROXY` are set in the environment, the child process (docker/podman/nerdctl/skopeo) **will use that proxy**.

**Conclusion:** By default, if proxy-related environment variables are set, image pull (and other operations) will use the proxy.

### 2. When proxy is enabled in config

If `proxy.enabled` is true in the config, the runtime is started with proxy env vars taken from the config. Those override any proxy variables from the environment.

### 3. Skopeo Save

Skopeo’s Save path sets `cmd.Env = os.Environ()`, so it still inherits the full environment; behavior is consistent.

## Inheritance rules

- If `cmd.Env == nil`: child gets the full parent environment.
- If `cmd.Env != nil`: child gets only the variables in `cmd.Env`.

## Behavior matrix

| Scenario | Proxy in config | Proxy in env | Result |
|----------|-----------------|--------------|--------|
| 1        | Disabled / unset | Set         | Uses env proxy |
| 2        | Disabled / unset | Not set     | No proxy |
| 3        | Enabled         | Set         | Uses config proxy (overrides env) |
| 4        | Enabled         | Not set     | Uses config proxy |

## Verification

```bash
export http_proxy=http://proxy.example.com:8080
export https_proxy=http://proxy.example.com:8080
./hpn pull -f images.txt
# Image operations should use the proxy (check proxy logs or network if needed).
```
