#!/bin/bash
rsync -avz --exclude ".git" --exclude config.json . fipso@raspberrypi:/home/fipso/screen-app/
ssh fipso@raspberrypi "killall screen-app"
