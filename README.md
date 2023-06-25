# free5GC compose

This repository is a docker compose version of [free5GC](https://github.com/free5gc/free5gc) for stage 3. It's inspired by [free5gc-docker-compose](https://github.com/calee0219/free5gc-docker-compose) and also reference to [docker-free5gc](https://github.com/abousselmi/docker-free5gc).

You can setup your own config in [config](./config) folder and [docker-compose.yaml](docker-compose.yaml)

## Prerequisites

- [GTP5G kernel module](https://github.com/free5gc/gtp5g): needed to run the UPF
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

Sometimes, you need to drop data from DB:

```bash
docker exec -it mongodb mongo
> use free5gc
> db.dropDatabase()
> exit # (Or Ctrl-D)
```

You can see logs for each service using `docker logs` command. For example, to access the logs of the *SMF* you can use:

```console
docker logs smf
```

Please refer to the [wiki](https://github.com/free5gc/free5gc/wiki) for more troubleshooting information.

## Integration with external gNB/UE simulators

The integration with the [UERANSIM](https://github.com/aligungr/UERANSIM) eNB/UE simulator is documented [here](https://www.free5gc.org/installations/stage-3-sim-install/). 

You can also refer to this [issue](https://github.com/free5gc/free5gc-compose/issues/26) to find out how you can configure the UPF to forward traffic between the [UERANSIM](https://github.com/aligungr/UERANSIM) to the DN (eg. internet) in a docker environment.

This [issue](https://github.com/free5gc/free5gc-compose/issues/28) provides detailed steps that might be useful.

## Integration of WebUI with Nginx reverse proxy

Here you can find helpful guidelines on the integration of Nginx reverse proxy to set it in front of the WebUI: https://github.com/free5gc/free5gc-compose/issues/55#issuecomment-1146648600

## Vagrant Box Option

For Linux kernel version below 5.4 you can setup a working environment using a vagrant box: https://github.com/abousselmi/vagrant-free5gc
Please refer to [GTP5G kernel module](https://github.com/free5gc/gtp5g) for more information.

## ULCL Configuration 
You can check the following informations below:
- [ulcl-example branch](https://github.com/free5gc/free5gc-compose/tree/ulcl-example), or
- [patch file](https://github.com/ianchen0119/free5gc-compose-ulcl)

## Reference
- https://github.com/open5gs/nextepc/tree/master/docker
- https://github.com/abousselmi/docker-free5gc
