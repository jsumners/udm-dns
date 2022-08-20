#!/bin/bash

set -e

if [ ! -z ${GITHUB_TOKEN} ]; then
  echo "Downloading remote configuration file: ${GITHUB_FILE}"
  curl -sOLf -H "authorization: token ${GITHUB_TOKEN}" \
    -H 'accept: application/vnd.github.v4.raw' \
    "https://api.github.com/repos/${GITHUB_ORG}/${GITHUB_REPO}/contents/${GITHUB_FILE}"
fi

hosts_dir=/etc/dnsmasq.hosts
unifi_hosts=${hosts_dir}/unifi.hosts

[ -d ${hosts_dir} ] || mkdir ${hosts_dir}

_DNSMASQ_OPTS=(
  --keep-in-foreground
  --hostsdir=${hosts_dir}
  --log-facility=-
)

if [ ! -z ${DNS_LOCAL_DOMAIN} ]; then
  _DNSMASQ_OPTS+=(
    --expand-hosts
    --domain=${DNS_LOCAL_DOMAIN}
  )
fi

if [ ! -z ${DNSMASQ_OPTS} ]; then
  _DNSMASQ_OPTS+=( ${DNSMASQ_OPTS} )
fi

dnsmasq $(IFS=' ' ; echo "${_DNSMASQ_OPTS[*]}") &
dnsmasq_pid=$!

function shutdown() {
    echo "Shutting down due to SIGTERM"
    kill $dnsmasq_pid
    wait $dnsmasq_pid
    exit 0
}
trap shutdown SIGTERM

while true; do
    api-client > /tmp/current_unifi.hosts \
    && ! diff -N ${unifi_hosts} /tmp/current_unifi.hosts \
    && mv /tmp/current_unifi.hosts ${unifi_hosts}
    sleep ${UDM_POLL_INTERVAL:-60} & wait $! # trapable sleep
done
