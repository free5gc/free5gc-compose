#!/bin/bash

sudo docker-compose down
sudo rm -rf dbdata/
sudo docker-compose up -d
