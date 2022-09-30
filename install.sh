#!/bin/bash
set -eu

curl -o githubssh https://cdn.kastelo.net/prm/a0r4sa6g.githubssh
chmod 755 githubssh
sudo chown root:root githubssh
sudo mv githubssh /usr/local/bin
(crontab -l | grep -v githubssh ; echo '30 */8 * * * /usr/local/bin/githubssh') | crontab -
/usr/local/bin/githubssh

