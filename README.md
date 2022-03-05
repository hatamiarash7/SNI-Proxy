# SNI-Proxy

[![Go](https://github.com/hatamiarash7/SNI-Proxy/actions/workflows/go.yml/badge.svg)](https://github.com/hatamiarash7/SNI-Proxy/actions/workflows/go.yml) [![Release](https://github.com/hatamiarash7/SNI-Proxy/actions/workflows/release.yml/badge.svg)](https://github.com/hatamiarash7/SNI-Proxy/actions/workflows/release.yml) [![Docker](https://github.com/hatamiarash7/SNI-Proxy/actions/workflows/docker.yml/badge.svg)](https://github.com/hatamiarash7/SNI-Proxy/actions/workflows/docker.yml) [![Publish](https://github.com/hatamiarash7/SNI-Proxy/actions/workflows/publish.yml/badge.svg)](https://github.com/hatamiarash7/SNI-Proxy/actions/workflows/publish.yml) 

A Simple SNI Proxy with internal DNS server

```bash
docker run -d -p 80:80 -p 443:443 -p 53:53 -v "$(pwd):/tmp/" --restart unless-stopped ghcr.io/hatamiarash7/sni-proxy:latest -list /tmp/list -PIP 185.235.42.191
```