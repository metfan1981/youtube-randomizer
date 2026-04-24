# youtube-randomizer

Serve **up to three random videos** from a YouTube channel’s uploads. Open `/@handle` in the browser, use the embedded player, reload for new picks.

Upload lists are cached per handle (`CACHE_TTL`, default **1h**) so you don’t hammer the YouTube Data API on every refresh.

---

## Requirements

- **Go** 1.22 or newer
- A [YouTube Data API v3](https://developers.google.com/youtube/v3) key (Google Cloud project with the API enabled)

---

## Makefile

| Target | What it does |
|--------|----------------|
| **`make build`** | Produces `./youtube-randomizer` |
| **`make help`** | Prints CLI/env usage (`go run . -h`) |
| **`make clean`** | Removes `./youtube-randomizer` |

---

## Quick start

Tracked template: **[`.env.example`](.env.example)** (safe to commit; replace the placeholder key).

```bash
cp .env.example .env
# Edit .env: set YT_API_KEY and optional HOST, PORT, CACHE_TTL, FETCH_TIMEOUT
make build
./youtube-randomizer
```

If **`.env`** exists in the **current working directory**, it is loaded on startup (skipped when you pass **`-h`**, **`-help`**, or **`--help`** so help still works if **`.env`** is broken). Values **already set in your shell** are **not** overwritten. **systemd** does not read repo **`.env`**—use **`EnvironmentFile=`** (see **[`systemd/README.md`](systemd/README.md)**). **`.env`** is gitignored for local secrets.

Then open **`http://<HOST>:<PORT>/@mkbhd`** (defaults: **127.0.0.1**, **8080**).

- Listen address: **`HOST`** env or **`-host`** (flag wins); default **loopback**. Use **`HOST=0.0.0.0`** or **`-host 0.0.0.0`** for all interfaces.
- Port: **`PORT`** or **`-port`** (flag wins); default **8080**.

You can still run **`go run .`** for quick iteration; flags work the same.

---

## HTTP routes

| URL | Purpose |
|-----|---------|
| **`/`** | Short help page |
| **`/@handle`** | Random embeds (handle is lowercased; bare part: `^[a-z0-9._-]{3,30}$`) |
| **`/healthz`** | Liveness (`ok`) |

---

## Environment

| Variable | Required | Default | Purpose |
|----------|----------|---------|---------|
| **`YT_API_KEY`** | yes | — | YouTube Data API key |
| **`HOST`** | no | `127.0.0.1` | Listen IP/hostname when **`-host`** is not set |
| **`PORT`** | no | `8080` | Port when **`-port`** is not set |
| **`CACHE_TTL`** | no | `1h` | How long to keep each channel’s upload ID list (`time.ParseDuration`) |
| **`FETCH_TIMEOUT`** | no | `3m` | Max time for a **full** upload-list fetch (cache miss); raise for huge channels |

## Flags

| Flag | Default | Purpose |
|------|---------|---------|
| **`-host`** | `$HOST` or `127.0.0.1` | Listen IP/hostname (**`-host`** overrides **`$HOST`**); **`0.0.0.0`** = all interfaces |
| **`-port`** | `$PORT` or `8080` | Listen port (**`-port`** overrides **`$PORT`**) |
| **`-help`**, **`-h`** | — | Print usage and exit **0** (no key required) |

---

## Quota and first load

**On each cache miss** (first request for a handle, or after **`CACHE_TTL`**):

- Cost ≈ **`1 + ceil(N / 50)`** quota units, where **`N`** = uploads in the playlist (one `channels.list` + one `playlistItems.list` per 50 items).
- Example: **`N = 2000`** → **41** units for that refresh.  
  **Cached** requests only pick random IDs locally.

**Large channels:** building the full list paginates through every page in **one** request. That can exceed a short deadline. Default **`FETCH_TIMEOUT=3m`**; increase (e.g. **`10m`**) if you see timeouts. After the list is cached, responses stay fast.

---

## Playlist note

The app uses the channel’s default **uploads** playlist (`UU…`). Some channels also have related lists (e.g. long-form **UULF…**, shorts **UUSH…**, live **UULV…**). If the mix feels wrong, a future version could switch playlists; v1 stays on **uploads**.

---

## systemd (Linux)

Full walkthrough (user, **`EnvironmentFile`**, firewall, logs): **[`systemd/README.md`](systemd/README.md)**.

Short version:

1. **`make build`** and **`sudo cp youtube-randomizer /usr/local/bin/`**
2. Create **`/etc/youtube-randomizer.env`** with at least **`YT_API_KEY`**; optional **`HOST`**, **`PORT`**, **`CACHE_TTL`**, **`FETCH_TIMEOUT`** (same names as [`.env.example`](.env.example)).
3. **`sudo cp systemd/youtube-randomizer.service /etc/systemd/system/`**, then **`daemon-reload`**, **`enable`**, **`start`**.

To bind all interfaces, set **`HOST=0.0.0.0`** in **`/etc/youtube-randomizer.env`** (no need to change **`ExecStart`**).
