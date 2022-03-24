# SNI-Proxy

[![Go](https://github.com/hatamiarash7/SNI-Proxy/actions/workflows/go.yml/badge.svg)](https://github.com/hatamiarash7/SNI-Proxy/actions/workflows/go.yml) [![Release](https://github.com/hatamiarash7/SNI-Proxy/actions/workflows/release.yml/badge.svg)](https://github.com/hatamiarash7/SNI-Proxy/actions/workflows/release.yml) [![Docker](https://github.com/hatamiarash7/SNI-Proxy/actions/workflows/docker.yml/badge.svg)](https://github.com/hatamiarash7/SNI-Proxy/actions/workflows/docker.yml) [![Publish](https://github.com/hatamiarash7/SNI-Proxy/actions/workflows/publish.yml/badge.svg)](https://github.com/hatamiarash7/SNI-Proxy/actions/workflows/publish.yml)

A Simple SNI Proxy with internal DNS server

## Binary

```bash
sudo ./SNI-Proxy -list domains -PIP <YOUR SERVER IP>
```

You can create service for the binary:

```ini
[Unit]
Description=SNI Proxy by A.Hatami - 2022

[Service]
User=root
WorkingDirectory=/home/ubuntu
ExecStart=sudo ./SNI-Proxy -PIP <YOUR SERVER IP> -list domains
Restart=always

[Install]
WantedBy=multi-user.target
```

## Docker

```bash
docker run -d -p 80:80 -p 443:443 -p 53:53 -v "$(pwd):/tmp/" --restart unless-stopped ghcr.io/hatamiarash7/sni-proxy:v1.1.1 -list /tmp/list -PIP YOUR SERVER IP
```
