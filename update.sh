#!/bin/bash
rsync -avz --exclude ".git" --exclude "screen-app" . fipso@raspberrypi:/home/fipso/screen-app/
ssh fipso@raspberrypi "killall screen-app || true; cd screen-app; go build .; DISPLAY=:0.0 ./screen-app &"
