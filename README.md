# free5GC Compose

This repository is a docker compose version of [free5GC](https://github.com/free5gc/free5gc) for stage 3. It's inspire by [free5gc-docker-compose](https://github.com/calee0219/free5gc-docker-compose) and also reference to [docker-free5GC](https://github.com/abousselmi/docker-free5gc).

You can change your own config in [config](./config) folder and [docker-compose.yaml](docker-compose.yaml)

## Pre-request

Due to the UPF issue, the host must using kernel `5.0.0-23-generic`. And it should contain `gtp5g` kernel module.

On you host OS:
```
git clone https://github.com/PrinzOwO/gtp5g.git
cd gtp5g
make
sudo make install
```

## How to use it
### Install docker.

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

### Install docker-compose
Reference: https://docs.docker.com/v17.09/compose/install/
```bash
$ sudo curl -L https://github.com/docker/compose/releases/download/1.18.0/docker-compose-`uname -s`-`uname -m` -o /usr/local/bin/docker-compose
$ sudo chmod +x /usr/local/bin/docker-compose
```

### Run Up
Because we need to create tunnel interface, we need to use privileged container with root permission.
```bash
$ git clone https://github.com/calee0219/free5gc-docker.git
$ cd free5gc-docker
$ docker-compose build
$ sudo docker-compose up # Recommand use with tmux to run in frontground
$ sudo docker-compose up -d # Run in backbround if need
```

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

| NF | IP | Exposed Ports | Dependencies | Dependencies URI |
|:-:|:-:|:-:|:-:|:-:|
| amf | 10.200.200.3 | 29518 | nrf | nrfUri: https://nrf:29510 |
| ausf | 10.200.200.4 | 29509 | nrf | nrfUri: https://nrf:29510 |
| n3iwf |
| nrf | 10.200.200.2 | 29510 | db | MongoDBUrl: mongodb://db:27017 |
| nssf | 10.200.200.5 | 29531 | nrf | nrfUri: https://nrf:29510gg/,<br/>nrfId: https://nrf:29510/nnrf-nfm/v1/nf-instances |
| pcf | 10.200.200.6 | 29507 | nrf | nrfUri: https://nrf:29510 |
| smf | 10.200.200.7 | 29502 | nrf, upf | nrfUri: https://nrf:29510,<br/>node_id: upf1, node_id: upf2, node_id: upf3 |
| udm | 10.200.200.8 | 29503 | nrf | nrfUri: https://nrf:29510 |
| udr | 10.200.200.9 | 29504 | nrf, db | nrfUri: https://nrf:29510,<br/>url: mongodb://db:27017 |
| upf1 | 10.200.200.101 | N/A | pfcp, gtpu, apn | pfcp: upf1, gtpu: upf1, apn: internet |
| upf2 | 10.200.200.103 | N/A | pfcp, gtpu, apn | pfcp: upf2, gtpu: upf2, apn: internet |
| upfb (ulcl) | 10.200.200.102 | N/A | pfcp, gtpu, apn | pfcp: upfb, gtpu: upfb, apn: intranet |

## Reference
- https://github.com/open5gs/nextepc/tree/master/docker
- https://github.com/abousselmi/docker-free5gc
