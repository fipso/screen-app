#!/bin/bash
scp index.html fipso@raspberrypi:/home/fipso/screen-app/
ssh fipso@raspberrypi DISPLAY=:0.0 xdotool key F5
