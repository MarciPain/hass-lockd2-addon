#!/usr/bin/with-contenv bashio
set -e

bashio::log.info "Starting Lockd2 Addon..."

# Elindul a Go backendünk a konténer környezeti változóival
/usr/bin/lockd2-addon
