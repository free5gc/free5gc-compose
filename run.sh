#!/bin/bash

sudo docker-compose down
sudo rm -rf dbdata/
sudo docker-compose up -d
sudo ./sample/config/subscribers.sh

sudo docker-compose exec my5gcore-n3iwf sh 
