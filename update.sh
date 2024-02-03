#!/bin/bash
rsync -avz --exclude ".git" . fipso@raspberrypi:/home/fipso/screen-app/
#ssh fipso@raspberrypi "cd ~/screen-app/crypto/; go build; killall crypto || true; cd ~/screen-app/bus/; go build; killall bus || true"
