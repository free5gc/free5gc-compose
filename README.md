# Free5GC Compose

This repository is a docker compose version of [free5GC](https://github.com/free5gc/free5gc) for stage 3. It's inspire by [free5gc-docker-compose](https://github.com/calee0219/free5gc-docker-compose) and also reference to [docker-free5GC](https://github.com/abousselmi/docker-free5gc).

You can change your own config in [config](./config) folder and [docker-compose.yaml](docker-compose.yaml)

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->


- [How to use it: bare metal](#how-to-use-it-bare-metal)
  - [Prerequisites](#prerequisites)
  - [Install Docker](#install-docker)
    - [Ubuntu](#ubuntu)
    - [CentOS](#centos)
  - [Install docker-compose](#install-docker-compose)
  - [Run Up](#run-up)
- [How to use it: vagrant box](#how-to-use-it-vagrant-box)
- [Troubleshooting](#troubleshooting)
- [NF](#nf)
- [Reference](#reference)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## How to use it: bare metal

### Prerequisites

Due to the UPF issue, the host must using kernel `5.0.0-23-generic`. And it should contain `gtp5g` kernel module.

On you host OS:
```
git clone https://github.com/PrinzOwO/gtp5g.git
cd gtp5g
make
sudo make install
```

### Install Docker

#### Ubuntu
Reference: https://docs.docker.com/install/linux/docker-ce/ubuntu/
```bash
$ sudo apt-get update
$ sudo apt-get install \
    apt-transport-https \
    ca-certificates \
    curl \
    gnupg-agent \
    software-properties-common

$ curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -

$ sudo add-apt-repository \
   "deb [arch=amd64] https://download.docker.com/linux/ubuntu \
   $(lsb_release -cs) \
   stable"

$ sudo apt-get update
$ sudo apt-get install docker-ce docker-ce-cli containerd.io
```

#### CentOS
Reference: https://docs.docker.com/install/linux/docker-ce/centos/
```bash
$ sudo yum install -y yum-utils \
  device-mapper-persistent-data \
  lvm2

$ sudo yum-config-manager \
    --add-repo \
    https://download.docker.com/linux/centos/docker-ce.repo

$ sudo yum install docker-ce docker-ce-cli containerd.io

$ sudo systemctl start docker
$ sudo systemctl enable docker
```

#### Add docker group
To let you use docker without root permission.

Reference: https://docs.docker.com/engine/install/linux-postinstall/
```bash
$ sudo groupadd docker
$ sudo usermod -aG docker $USER
$ sudo reboot
```


### Install docker-compose
Reference: https://docs.docker.com/compose/install/
```bash
$ sudo curl -L https://github.com/docker/compose/releases/download/1.25.5/docker-compose-`uname -s`-`uname -m` -o /usr/local/bin/docker-compose
$ sudo chmod +x /usr/local/bin/docker-compose
```

### Run Up
Because we need to create tunnel interface, we need to use privileged container with root permission.
```bash
$ git clone https://github.com/free5gc/free5gc-compose.git
$ cd free5gc-compose
$ make base
$ docker-compose build
$ sudo docker-compose up # Recommand use with tmux to run in frontground
$ sudo docker-compose up -d # Run in backbround if needed
```

## How to use it: vagrant box
You can setup a working environment without the fuss of updating your kernel version just by using a vagrant box.

Please find follow the instructions provided here: https://github.com/abousselmi/vagrant-free5gc


## Troubleshooting
Sometimes, you need to drop data from DB(See #Troubleshooting from https://www.free5gc.org/installation).
```bash
$ docker exec -it mongodb mongo
> use free5gc
> db.subscribers.drop()
> exit # (Or Ctrl-D)
```

Another way to drop DB data is just remove db data. Outside your container, run:
```bash
$ rm -rf ./mongodb
```

## NF

For my default setting.

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
