#!/bin/bash
rsync -avz --exclude ".git" --exclude config.json . fipso@screen2:/home/fipso/screen-app/
ssh fipso@screen2 "killall screen-app"
