# Free5GC Compose

This repository is a docker compose version of [Free5GC](https://github.com/free5gc/free5gc) for stage 3. It's inspired by [free5gc-docker-compose](https://github.com/calee0219/free5gc-docker-compose) and also reference to [docker-free5gc](https://github.com/abousselmi/docker-free5gc).

You can setup your own config in [config](./config) folder and [docker-compose.yaml](docker-compose.yaml)

## Prerequisites


- [GTP5G kernel module](https://github.com/free5gc/gtp5g): needed to run the UPF
- [Docker Engine](https://docs.docker.com/engine/install): needed to run the Free5GC containers
- [Docker Compose](https://docs.docker.com/compose/install): needed to bootstrap the Free5GC stack

## Start Free5gc

Because we need to create tunnel interface, we need to use privileged container with root permission.

```bash
# Clone the project
git clone https://github.com/free5gc/free5gc-compose.git
cd free5gc-compose

# Build the images
make base
docker-compose build

# Run it
sudo docker-compose up # add -d to run in background mode
```

Destroy the established container resource after testing:
```
docker-compose rm
```

## Troubleshooting

Sometimes, you need to drop data from DB:

```bash
docker exec -it mongodb mongo
> use free5gc
> db.subscribers.drop()
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

## Vagrant Box Option

You can setup a working environment without the fuss of updating your kernel version just by using a vagrant box. You can follow the instructions provided here: https://github.com/abousselmi/vagrant-free5gc

## Reference
- https://github.com/open5gs/nextepc/tree/master/docker
- https://github.com/abousselmi/docker-free5gc
