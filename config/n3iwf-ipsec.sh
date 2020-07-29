#!/bin/sh

### N3iwf IPSec tunnel configuration

ip route del default # deleta a rota default para funcionar com várias redes via docker-compose
ip l add name ipsec0 type vti local 192.168.127.1 remote 0.0.0.0 key 5
ip a add 10.0.0.1/24 dev ipsec0
ip l set dev ipsec0 up
./n3iwf -n3iwfcfg ../config/n3iwfcfg.conf