#!/bin/bash
rsync -avz --exclude ".git" . fipso@raspberrypi:/home/fipso/screen-app/
ssh fipso@raspberrypi "killall screen-app"
