#!/bin/sh

echo "Stop old PICS"
forever stopall

echo "Start new PICS"
npm start --prefix ${STEAM_PICS_PATH} &
