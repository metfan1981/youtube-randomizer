# Run on a server (systemd)

Clone the repo on the server, install, configure, start. Requires Go 1.22+ on the server.

## 1. Clone and install

```bash
git clone <this-repo-url> youtube-randomizer
cd youtube-randomizer
sudo make install
```

`sudo make install` (idempotent) does:

- `go build` the binary
- `install` it to `/usr/local/bin/youtube-randomizer`
- create system user `youtube-randomizer` (if missing)
- copy `.env.example` to `/etc/youtube-randomizer.env` with mode 600 (if missing)
- copy the systemd unit and run `daemon-reload`

## 2. Set the API key

```bash
sudoedit /etc/youtube-randomizer.env
```

At minimum set **`YT_API_KEY`**. Optional: `HOST`, `PORT`, `CACHE_TTL`, `FETCH_TIMEOUT` — same names as [`.env.example`](../.env.example). To bind all interfaces use **`HOST=0.0.0.0`**; otherwise keep `127.0.0.1` and put a reverse proxy in front.

## 3. Start

```bash
sudo systemctl enable --now youtube-randomizer
curl -sS http://127.0.0.1:8080/healthz         # -> "ok"
```

## Manage

```bash
sudo systemctl status   youtube-randomizer
sudo systemctl restart  youtube-randomizer     # after editing the env file
sudo systemctl stop     youtube-randomizer
sudo systemctl disable  youtube-randomizer     # don't start at boot
sudo journalctl -u youtube-randomizer -f       # follow logs
sudo journalctl -u youtube-randomizer -n 200   # recent logs
```

## Update

Pull and reinstall:

```bash
git pull
sudo make install
sudo systemctl restart youtube-randomizer
```

Edit config only:

```bash
sudoedit /etc/youtube-randomizer.env
sudo systemctl restart youtube-randomizer
```

## Uninstall

```bash
sudo make uninstall    # removes binary + unit, keeps env file and user
```

## Notes

- The unit sets `WorkingDirectory=/`, so a repo-local `.env` is never read under systemd.
- Hardening directives are in the unit. If the service fails to start on a quirky kernel/arch, comment out `MemoryDenyWriteExecute=` in `/etc/systemd/system/youtube-randomizer.service` and `daemon-reload + restart`.
- **TLS and public exposure:** keep `HOST=127.0.0.1` and terminate TLS in nginx/Caddy proxying to that port.
