#!/bin/bash
# Quick script to run on Linode servers running prod freno

set -e

echo "$(date): Kicking off freno..."

export HOSTNAME=`hostname`
aws ecr get-login --no-include-email --region us-east-1 | bash
docker pull 769352775470.dkr.ecr.us-east-1.amazonaws.com/freno:latest

docker stop freno
docker rm freno
docker run -d --network=host --name=freno \
--restart=always 769352775470.dkr.ecr.us-east-1.amazonaws.com/freno:latest \
--http --raft-bind $(/sbin/ip -o -4 addr list eth0 | tail -1 | awk '{print $4}' | cut -d/ -f1)

echo "$(date): Done!"
