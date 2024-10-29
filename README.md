# free5GC compose

This repository is a docker compose version of [free5GC](https://github.com/free5gc/free5gc) for stage 3. It's inspired by [free5gc-docker-compose](https://github.com/calee0219/free5gc-docker-compose) and also reference to [docker-free5gc](https://github.com/abousselmi/docker-free5gc).

You can setup your own config in [config](./config) folder and [docker-compose.yaml](docker-compose.yaml)

## Prerequisites

- [GTP5G kernel module](https://github.com/free5gc/gtp5g): needed to run the UPF (Currently, UPF only supports GTP5G versions between 0.8.6 and 0.8.10 (use git clone --branch v0.8.10 --depth 1 https://github.com/free5gc/gtp5g.git).)
- [Docker Engine](https://docs.docker.com/engine/install): needed to run the Free5GC containers
- [Docker Compose v2](https://docs.docker.com/compose/install): needed to bootstrap the free5GC stack

**Note: AVX for MongoDB**: some HW does not support MongoDB releases above`4.4` due to use of the new AVX instructions set. To verify if your CPU is compatible you can check CPU flags by running `grep avx /proc/cpuinfo`. A workaround is suggested [here](https://github.com/free5gc/free5gc-compose/issues/30#issuecomment-897627049).

## Start free5gc

Because we need to create tunnel interface, we need to use privileged container with root permission.

### Pull docker images from Docker Hub

```bash
docker compose pull
```

### [Optional] Build docker images from local sources

```bash
# Clone the project
git clone https://github.com/free5gc/free5gc-compose.git
cd free5gc-compose

# clone free5gc sources
cd base
git clone --recursive -j `nproc` https://github.com/free5gc/free5gc.git
cd ..

# Build the images
make all
docker compose -f docker-compose-build.yaml build

# Alternatively you can build specific NF image e.g.:
make amf
docker compose -f docker-compose-build.yaml build free5gc-amf
```

Note:

Dangling images may be created during the build process. It is advised to remove them from time to time to free up disk space.

```bash
docker rmi $(docker images -f "dangling=true" -q)
```

### Run free5GC

You can create free5GC containers based on local images or docker hub images:

```bash
# use local images
docker compose -f docker-compose-build.yaml up
# use images from docker hub
docker compose up # add -d to run in background mode
```

Destroy the established container resource after testing:

```bash
# Remove established containers (local images)
docker compose -f docker-compose-build.yaml rm
# Remove established containers (remote images)
docker compose rm
```

## Troubleshooting

Please refer to the [Troubleshooting](./TROUBLESHOOTING.md) for more troubleshooting information.

## Integration with (external) gNB/UE

### UERANSIM Notes

The integration with the [UERANSIM](https://github.com/aligungr/UERANSIM) eNB/UE simulator is documented [here](https://free5gc.org/guide/5-install-ueransim/).

This [issue](https://github.com/free5gc/free5gc-compose/issues/28) provides detailed steps that might be useful.

#### Option 1: Run UE inside gNB container

You can launch a UE using:

```console
docker exec -it ueransim bash
root@host:/ueransim# ./nr-ue -c config/uecfg.yaml
```

#### Option 2: Run UE on a separate container

By default, the provided UERANSIM service on this `docker-compose.yaml` will only act as a gNB. If you want to create a UE you'll need to:

1. Create a subscriber through the WebUI. Follow the steps [here](https://free5gc.org/guide/Webconsole/Create-Subscriber-via-webconsole/#4-open-webconsole)
1. Copy the `UE ID` field
1. Change the value of `supi` in `config/uecfg.yaml` to the UE ID that you just copied
1. Change the `linkIp` in `config/gnbcfg.yaml` to `gnb.free5gc.org` (which is also present in the `gnbSearchList` in `config/uecfg.yaml`) to enable communication between the UE and gNB services
1. Add an UE service on `docker-compose.yaml` as it follows:

```yaml
ue:
  container_name: ue
  image: free5gc/ueransim:latest
  command: ./nr-ue -c ./config/uecfg.yaml
  volumes:
    - ./config/uecfg.yaml:/ueransim/config/uecfg.yaml
  cap_add:
    - NET_ADMIN
  devices:
    - "/dev/net/tun"
  networks:
    privnet:
      aliases:
        - ue.free5gc.org
  depends_on:
    - ueransim
```

5. Run `docker-compose.yaml`

### srsRAN Notes

You can check this [issue](https://github.com/free5gc/free5gc-compose/issues/94) for some sample configuration files of srsRAN + free5GC

## Integration of WebUI with Nginx reverse proxy

Here you can find helpful guidelines on the integration of Nginx reverse proxy to set it in front of the WebUI: https://github.com/free5gc/free5gc-compose/issues/55#issuecomment-1146648600

## ULCL Configuration

To start the core with a I-UPF and PSA-UPF ULCL configuration, use

```bash
docker compose -f docker-compose-ulcl.yaml up
```

> Note: This configuration have been tested using release [free5gc-compose v3.4.3](https://github.com/free5gc/free5gc-compose/tree/v3.4.3)

Check out the used configuration files at `config/ULCL`.

## Reference

- https://github.com/open5gs/nextepc/tree/master/docker
- https://github.com/abousselmi/docker-free5gc
