#!/bin/bash

./nr-gnb -c ./config/gnbcfg.yaml & ./nr-ue -c config/uecfg.yaml
