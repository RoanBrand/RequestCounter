#!/bin/sh

export PORT
export REQCOUNTER_ADDR

envsubst '${PORT} ${REQCOUNTER_ADDR}' < /nginx.conf.template > /etc/nginx/nginx.conf

exec /docker-entrypoint.sh "$@"
