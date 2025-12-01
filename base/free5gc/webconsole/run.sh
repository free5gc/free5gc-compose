#!/bin/bash

########################################################
#
# This script is used to build and run the free5gc-Webconsole
#
# For quickly developing used
#
##########

cd frontend

# check yarn install
if [ ! -d "node_modules" ]; then
    echo "node_modules not found, installing..."
    yarn install
else
    echo "node_modules found, skipping installation"
fi

# yarn build
echo "building frontend..."
yarn build

# copy build to public
echo "copying build to public..."
rm -rf ../public
cp -R build ../public
cd ..

# run server
echo "running server..."
go run server.go
