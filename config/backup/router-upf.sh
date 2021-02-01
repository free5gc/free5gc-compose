#!/bin/bash

#./free5gc-upfd -f ../config/upfcfg.yaml & 


sleep 3

#Configuração da rede pare teste do ping
ip link set lo up 
ip addr add 60.60.0.101 dev lo

#Configuração de regras de reteamento
iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE

#Configuração tunel gtp
#ip link set dev upfgtp mtu 1500


echo "Finish run script !!!"
