#!/bin/sh

set -e

ME=$(basename $0)

generate() {
  local domains="${DOMAIN_LIST}"
  local email="${CERTBOT_EMAIL}"
  if [ -z "$domains" ]; then
    echo "$ME: not generation certificates because domain list is empty"
    return 0
  fi

  apt-get install -y certbot python-certbot-nginx wget
  echo "$ME: generating certificates"
  certbot certonly --standalone --agree-tos -m "${email}" -n -d "$domains"

  rm -rf /var/lib/apt/lists/* &&
    echo "PATH=$PATH" >/etc/cron.d/certbot-renew &&
    echo "@monthly certbot renew --nginx >> /var/log/cron.log 2>&1" >>/etc/cron.d/certbot-renew &&
    crontab /etc/cron.d/certbot-renew
}

generate

exit 0
