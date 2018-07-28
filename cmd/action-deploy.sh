#!/bin/sh

cd ../

echo "### Pulling"
git fetch origin
git reset --hard origin/master

echo "### Building"
dep ensure
go build

echo "### Copying crontab"
cp ./crontab /etc/cron.d/steamauthority

echo "### Updating Varnish"
cp ./varnish.vcl /etc/varnish/default.vcl
/etc/init.d/varnish reload

echo "### Talking to Rollbar"
curl https://api.rollbar.com/api/1/deploy/ \
  -F access_token=${STEAM_ROLLBAR_PRIVATE} \
  -F environment=${ENV} \
  -F revision=$(git log -n 1 --pretty=format:"%H") \
  -F local_username=${STEAM_ROLLBAR_USER} \
  --silent > /dev/null

echo "### Restarting"
/etc/init.d/steam restart
