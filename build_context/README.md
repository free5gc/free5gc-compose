# Note
> All the images be built when `make all` command executed except `upf`, `ueransim` and `webui`.

# Build the images (UPF, UERANSIM, WebUI)

For UPF:
```bash
make upf
docker compose -f docker-compose-build.yaml build free5gc-upf
```

For UERANSIM:
```bash
docker compose -f docker-compose-build.yaml build ueransim
```

For WebUI:
```bash
make webconsole
docker compose -f docker-compose-build.yaml build free5gc-webui
```
