# Free5GC Compose

This repository is a docker compose version of [free5GC](https://github.com/free5gc/free5gc) for stage 3. It's inspire by [free5gc-docker-compose](https://github.com/calee0219/free5gc-docker-compose) and also reference to [docker-free5GC](https://github.com/abousselmi/docker-free5gc).

You can setup your own config in [config](./config) folder and [docker-compose.yaml](docker-compose.yaml)

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->


- [Prerequisites](#prerequisites)
  - [GTP5G kernel module](#gtp5g-kernel-module)
  - [Docker engine](#docker-engine)
  - [Docker compose](#docker-compose)
- [Start Free5gc](#start-free5gc)
- [Troubleshooting](#troubleshooting)
- [Vagrant Box Option](#vagrant-box-option)
- [NF dependencies and ports](#nf-dependencies-and-ports)
- [Reference](#reference)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Prerequisites

### GTP5G kernel module

Due to the UPF issue, the host must using kernel `5.0.0-23-generic`. And it should contain `gtp5g` kernel module.

On you host OS:
```
git clone https://github.com/PrinzOwO/gtp5g.git
cd gtp5g
make
sudo make install
```

### Docker engine

To install docker on your favorite OS, you can follow instruction here: https://docs.docker.com/engine/install/

### Docker compose

You also need to install docker compose as detailed here: https://docs.docker.com/compose/install/

## Start Free5gc

Because we need to create tunnel interface, we need to use privileged container with root permission.

```bash
$ git clone https://github.com/free5gc/free5gc-compose.git
$ cd free5gc-compose
$ make base
$ docker-compose build
$ sudo docker-compose up # Recommend use with tmux to run in frontground
$ sudo docker-compose up -d # Run in backbround if needed
```

## Troubleshooting

Sometimes, you need to drop data from DB(See #Troubleshooting from https://www.free5gc.org/installation).

```bash
$ docker exec -it mongodb mongo
> use free5gc
> db.subscribers.drop()
> exit # (Or Ctrl-D)
```

You can see logs for each service using `docker logs` command. For example, to access the logs of the *SMF* you can use:

```console
docker logs smf
```

Another way to drop DB data is just remove db data. Outside your container, run:
```bash
$ rm -rf ./mongodb
```

## Vagrant Box Option

You can setup a working environment without the fuss of updating your kernel version just by using a vagrant box. You can follow the instructions provided here: https://github.com/abousselmi/vagrant-free5gc


## NF dependencies and ports

| NF | Exposed Ports | Dependencies | Dependencies URI |
|:-:|:-:|:-:|:-:|
| amf | 29518 | nrf | nrfUri: https://nrf:29510 |
| ausf | 29509 | nrf | nrfUri: https://nrf:29510 |
| nrf | 29510 | db | MongoDBUrl: mongodb://db:27017 |
| nssf | 29531 | nrf | nrfUri: https://nrf:29510gg/,<br/>nrfId: https://nrf:29510/nnrf-nfm/v1/nf-instances |
| pcf | 29507 | nrf | nrfUri: https://nrf:29510 |
| smf | 29502 | nrf, upf | nrfUri: https://nrf:29510,<br/>node_id: upf1, node_id: upf2, node_id: upf3 |
| udm | 29503 | nrf | nrfUri: https://nrf:29510 |
| udr | 29504 | nrf, db | nrfUri: https://nrf:29510,<br/>url: mongodb://db:27017 |
| n3iwf | N/A | amf, smf, upf |  |
| upf1 | N/A | pfcp, gtpu, apn | pfcp: upf1, gtpu: upf1, apn: internet |
| upf2 | N/A | pfcp, gtpu, apn | pfcp: upf2, gtpu: upf2, apn: internet |
| upfb (ulcl) | N/A | pfcp, gtpu, apn | pfcp: upfb, gtpu: upfb, apn: intranet |
| webui | 5000 | db | MongoDBUrl: mongodb://db:27017  |

## Reference
- https://github.com/open5gs/nextepc/tree/master/docker
- https://github.com/abousselmi/docker-free5gc
