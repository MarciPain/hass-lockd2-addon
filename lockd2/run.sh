#!/bin/sh
set -e

echo "Starting Lockd2 Addon..."

# Elindul a Go backendünk
exec /usr/bin/lockd2-addon
