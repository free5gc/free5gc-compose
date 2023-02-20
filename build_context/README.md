# Note
> All the images be built when `make all` command executed except `upf`, `ueransim` and `webui`.

## Build the images (UPF, UERANSIM, WebUI)

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

## DEBUG_TOOL

If you want to use debug tool, you can build the images with `DEBUG_TOOL` option:
1. Replace the DEBUG_ENABLE with `true` in `Makefile`
2. Replace all the DEBUG_ENABLE with `true` in `docker-compose-build.yaml`