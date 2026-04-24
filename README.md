# youtube-randomizer

Open **`/@handle`** in your browser, get **up to three random videos** from that channel embedded. Refresh for new picks.

Uploads are cached per channel (**`CACHE_TTL`**, default **1h**) so refreshes don’t burn YouTube API quota.

## Bootstrap (local)

```bash
cp .env.example .env      # edit: at least YT_API_KEY=...
make build
./youtube-randomizer
```

Open **`http://127.0.0.1:8080/@mkbhd`** (replace with a real handle).

> Get a key: [YouTube Data API v3](https://developers.google.com/youtube/v3) → Google Cloud Console → create project → enable API → **Credentials → API key**.

## Server (systemd on Linux)

See **[`systemd/README.md`](systemd/README.md)** — bootstrap, run, manage.

## Config

`.env` (local) or **`EnvironmentFile=`** (systemd). Shell-exported vars win over `.env`.

| Variable | Default | Purpose |
|---|---|---|
| `YT_API_KEY` | — | **Required.** YouTube Data API v3 key |
| `HOST` | `127.0.0.1` | Listen IP (`0.0.0.0` for all interfaces) |
| `PORT` | `8080` | Listen port |
| `CACHE_TTL` | `1h` | How long each channel’s upload list is cached |
| `FETCH_TIMEOUT` | `3m` | Max time to paginate all uploads on a cache miss |

Flags override env: **`-host`**, **`-port`**, **`-h`/`-help`** (exit 0). Run **`make help`** to see them.

## Routes

| URL | Purpose |
|---|---|
| `/` | Short usage page |
| `/@handle` | Three random embeds (handle must match `^[a-z0-9._-]{3,30}$`) |
| `/healthz` | Liveness check |

## Make targets

| Target | Purpose |
|---|---|
| `make build`        | compile `./youtube-randomizer` |
| `make help`         | print CLI usage |
| `make clean`        | remove the local binary |
| `sudo make install` | install prebuilt binary, user, env file, and systemd unit (run `make build` first — see [`systemd/README.md`](systemd/README.md)) |
| `sudo make uninstall` | remove binary and unit (keeps env file and user) |

## Quota note

Each **cache miss** costs **`1 + ceil(N/50)`** quota units (one `channels.list` + one `playlistItems.list` page per 50 uploads). Very large channels can exceed a short deadline on first load — bump **`FETCH_TIMEOUT`** (e.g. **`10m`**) if needed. Cached refreshes are local-only.

The app uses the channel’s default **uploads** playlist (`UU…`); related playlists (`UULF…` long-form, `UUSH…` shorts, `UULV…` live) are not swapped in v1.

## Troubleshooting

- **`invalid handle`** → URL must be **`/@name`** where `name` matches `^[a-z0-9._-]{3,30}$`.
- **`could not load channel`** → check **`YT_API_KEY`**, the handle, and `journalctl` / stderr for the real error (sanitized from the browser).
- **Slow first load** → large channel, raise **`FETCH_TIMEOUT`**.
- **Port in use** → change **`PORT`** or **`-port`**.
