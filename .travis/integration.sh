#!/bin/bash

set -e
set -x

TAG=test

# before_install
#curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
#sudo add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"
#sudo apt-get update
#sudo apt-get -y install docker-ce

# install
sudo docker pull busybox

# default root source is /data
sudo mkdir /data

# make the plugin
sudo PLUGIN_TAG=$TAG make
# enable the plugin
sudo docker plugin enable hectorqin/local-mapping:$TAG
# list plugins
sudo docker plugin ls

# test1: simple
# create volume
sudo docker volume create -d hectorqin/local-mapping:$TAG -o mountpoint=/tmp localvolume
# sudo cat /var/lib/docker/plugins/local-mapping.json
# check the state
grep -Fxq "/mnt/root/tmp" /var/lib/docker/plugins/local-mapping.json
# write the volume in container
sudo docker run --rm -v localvolume:/write busybox sh -c "echo hello > /write/world"
# check the volume on host
grep -Fxq hello /data/tmp/world
# read the volume on other container
sudo docker run --rm -v localvolume:/read busybox grep -Fxq hello /read/world
# remove volume
sudo docker volume rm localvolume

# test2: change state source
# set the state source
sudo docker plugin disable hectorqin/local-mapping:$TAG
sudo docker plugin set hectorqin/local-mapping:$TAG state.source=/tmp
sudo docker plugin enable hectorqin/local-mapping:$TAG
# create volume
sudo docker volume create -d hectorqin/local-mapping:$TAG -o mountpoint=/tmp localvolume
# sudo cat /tmp/local-mapping.json
# check the state
grep -Fxq "/mnt/root/tmp" /tmp/local-mapping.json
# write the volume in container
sudo docker run --rm -v localvolume:/write busybox sh -c "echo hello > /write/world"
# check the volume on host
grep -Fxq hello /data/tmp/world
# read the volume on other container
sudo docker run --rm -v localvolume:/read busybox grep -Fxq hello /read/world
# remove volume
sudo docker volume rm localvolume

# test3: change root source
# set the state source
sudo docker plugin disable hectorqin/local-mapping:$TAG
sudo docker plugin set hectorqin/local-mapping:$TAG root.source=/tmp
sudo docker plugin enable hectorqin/local-mapping:$TAG
# create volume
sudo docker volume create -d hectorqin/local-mapping:$TAG -o mountpoint=/tmp localvolume
# sudo cat /var/lib/docker/plugins/local-mapping.json
# check the state
grep -Fxq "/mnt/root/tmp" /var/lib/docker/plugins/local-mapping.json
# write the volume in container
sudo docker run --rm -v localvolume:/write busybox sh -c "echo hello > /write/world"
# check the volume on host
grep -Fxq hello /tmp/tmp/world
# read the volume on other container
sudo docker run --rm -v localvolume:/read busybox grep -Fxq hello /read/world
# remove volume
sudo docker volume rm localvolume