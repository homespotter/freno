#!/bin/bash
# Quick script to run on Linode servers running staging freno

set -e

echo "$(date): Kicking off freno..."

export HOSTNAME=`hostname`
aws ecr get-login --no-include-email --region us-east-1 | bash
docker pull 769352775470.dkr.ecr.us-east-1.amazonaws.com/freno:staging

docker stop freno
docker rm freno
docker run -d --network=host --name=freno \
--restart=always 769352775470.dkr.ecr.us-east-1.amazonaws.com/freno:staging

echo "$(date): Done!"
