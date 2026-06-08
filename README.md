# ssh-website

A personal "business card" served over SSH. When someone runs `ssh denizlg24.com`,
they get a rendered terminal portrait + info card (built with
[charmbracelet/wish](https://github.com/charmbracelet/wish) + lipgloss).

The card's age is computed live from the birthdate, so it never goes stale, and the
ASCII portrait is rendered with a vertical forest-green→brown gradient.

## Local preview

See exactly what visitors get, straight to your own terminal — no SSH needed:

```sh
PREVIEW=1 go run .
```

Or via the Makefile (runs a real server on port 2222, then `ssh -p 2222 localhost`):

```sh
make run
```

## Build

```sh
make build       # native binary for the current machine
make build-pi    # cross-compiled for Raspberry Pi Zero W (linux/arm/ARMv6)
```

## Deploy to a Raspberry Pi Zero W

The goal: this app answers normal SSH (`ssh denizlg24.com`, port **22**), so the
Pi's own SSH daemon must move out of the way to port **2222**.

### 1. Move the system SSH daemon to port 2222

On the Pi, edit `/etc/ssh/sshd_config`:

```sh
sudo sed -i 's/^#\?Port .*/Port 2222/' /etc/ssh/sshd_config
sudo systemctl restart ssh
```

> ⚠️ Keep your current SSH session open until you've confirmed the new port works.
> From your laptop, test in a second terminal: `ssh -p 2222 denizlg24@<pi-ip>`.
> Only close the old session once 2222 logs you in.

If you use UFW or another firewall, open the new admin port:

```sh
sudo ufw allow 2222/tcp
sudo ufw allow 22/tcp
```

### 2. Cross-compile and copy the app to the Pi

From your dev machine (the Pi Zero W is too slow to build Go comfortably):

```sh
make build-pi
scp -P 2222 ssh-server ssh-server.service denizlg24@<pi-ip>:~/ssh-website/
```

(The repo expects the app to live at `/home/denizlg24/ssh-website/` — see the service file.)

### 3. Host key

`wish` auto-generates an ed25519 host key at `HOST_KEY_PATH` on first run if it
doesn't exist. To pin one yourself instead:

```sh
mkdir -p ~/ssh-website/.ssh
ssh-keygen -t ed25519 -f ~/ssh-website/.ssh/host_key -N ""
```

### 4. Install and run the systemd service

The service binds port **22**, which requires root (or `CAP_NET_BIND_SERVICE`).
It runs as root by default since no `User=` is set.

```sh
sudo cp ~/ssh-website/ssh-server.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now ssh-server
```

Check it:

```sh
systemctl status ssh-server
journalctl -u ssh-server -f      # live logs
```

### 5. Test it

From anywhere:

```sh
ssh me.denizlg24.com          # or ssh <pi-ip>
```

You should see the card. Admin access is now `ssh -p 2222 denizlg24@<pi-ip>`.

## Updating after a code change

```sh
make build-pi
scp -P 2222 ssh-server denizlg24@<pi-ip>:~/ssh-website/
ssh -p 2222 denizlg24@<pi-ip> 'sudo systemctl restart ssh-server'
```

## Environment variables

| Var             | Default          | Purpose                          |
|-----------------|------------------|----------------------------------|
| `HOST`          | `0.0.0.0`        | Bind address                     |
| `PORT`          | `22`             | Listen port                      |
| `HOST_KEY_PATH` | `.ssh/host_key`  | SSH host key location            |
| `PREVIEW`       | (unset)          | Set to `1` to print card & exit  |
