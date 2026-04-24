# Running youtube-randomizer with systemd

This is a minimal production-style setup on Linux: dedicated user, config file on disk, automatic restart on failure.

## 1. Build the binary

On your build machine (or the server, if it has Go):

```bash
make build
```

Copy the binary to the server, e.g.:

```bash
sudo cp youtube-randomizer /usr/local/bin/
sudo chmod 755 /usr/local/bin/youtube-randomizer
```

## 2. Service user

Use a non-login system account so the process does not run as root:

```bash
sudo useradd --system --no-create-home --shell /usr/sbin/nologin youtube-randomizer
```

## 3. Environment file

systemd loads variables from a single file (same `KEY=value` style as `.env` in this repo, but without `export`):

```bash
sudo install -m 600 /dev/null /etc/youtube-randomizer.env
sudo chown root:root /etc/youtube-randomizer.env
sudo nano /etc/youtube-randomizer.env
```

Minimum content:

```ini
YT_API_KEY=your-real-key-here
```

Optional (same names as in [`.env.example`](../.env.example)):

```ini
HOST=127.0.0.1
PORT=8080
CACHE_TTL=1h
FETCH_TIMEOUT=3m
```

**Security:** mode `600`, owned by root. The service user does not need read access to this file; systemd injects the variables into the process.

To listen on all interfaces (e.g. behind a reverse proxy on the same host, you may still prefer loopback):

```ini
HOST=0.0.0.0
PORT=8080
```

The unit sets **`WorkingDirectory=/`**, so the process never loads a repo **`.env`** (that only applies when you run the binary from a directory that contains **`.env`**). Under systemd, config comes from **`EnvironmentFile=`** only. The binary still reads **`HOST`**, **`PORT`**, etc. from the injected environment—the same variable names as [`.env.example`](../.env.example).

## 4. Install the unit

```bash
sudo cp systemd/youtube-randomizer.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable youtube-randomizer
sudo systemctl start youtube-randomizer
```

Check status:

```bash
sudo systemctl status youtube-randomizer
curl -sS http://127.0.0.1:8080/healthz
```

Logs:

```bash
sudo journalctl -u youtube-randomizer -f
```

## 5. Hardening and networking

The unit file adds common **systemd** sandboxing (no ambient caps, IPv4/IPv6-only sockets, `MemoryDenyWriteExecute`, etc.). If the binary fails to start on an unusual host, try commenting out **`MemoryDenyWriteExecute=`** first, then ask in your distro’s forums.

- **Firewall:** allow only what you need (e.g. proxy → `PORT`, or nothing public if you only use SSH + local curl).
- **TLS:** terminate HTTPS in **nginx**, **Caddy**, or similar, and proxy to `127.0.0.1:PORT` with **`HOST=127.0.0.1`**.
- **Unit tweaks:** adjust `ProtectSystem=`, `ReadWritePaths=`, or add capabilities only if you know you need them. Prefer a reverse proxy instead of binding directly to a privileged port.

## 6. Updating the binary

```bash
sudo systemctl stop youtube-randomizer
sudo cp /path/to/new/youtube-randomizer /usr/local/bin/
sudo systemctl start youtube-randomizer
```

After changing `/etc/youtube-randomizer.env`:

```bash
sudo systemctl restart youtube-randomizer
```
