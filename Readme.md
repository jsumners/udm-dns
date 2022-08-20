# udm-dns

`udm-dns` provides a [dnsmasq][] instance that registers hosts from
a [Ubiquity UDM-PRO][udm-pro] system. It is based upon the work done in
[unifi-dns][].

This is known to work with `UniFi OS UDM Pro 1.12.22` and `Network 7.1.68`.

## How It Works

The Docker container periodically polls the UDM-PRO for a list of clients,
parses that information, and writes a hosts file based upon it. Dnsmasq picks
up any changes in the hosts file and serves the content within.

Polling of the UDM-PRO is done by the included [api-client](./api-client/)
tool.

## Install

Install is done by writing a [docker-compose][docker-compose] file and running
it on a local Docker instance. An example compose file is:

```yaml
services:
  udm-dns:
    image: ghcr.io/jsumners/udm-dns
    container_name: udm-dns
    restart: unless-stopped

    cap_add:
      - NET_ADMIN
    ports:
      - 53:53/tcp
      - 53:53/udp
    dns:
      # Upstream DNS servers for recursive lookups,
      # i.e. the DNS servers Dnsamasq will use.
      # At least one should probably be a local ISP server.
      - 8.8.8.8
      - 8.8.4.4

    environment:
      API_CLIENT_ADDRESS: 192.168.1.1
      API_CLIENT_USERNAME: api
      API_CLIENT_PASSWORD: super-secret
      API_CLIENT_SITE: default
```

## Configuration

There are two items in `udm-dns` that can be configured: the container and
the `api-client`.

### Container

The container is configured completely through environment variables. The
variables available are:

* `GITHUB_TOKEN`: GitHub [personal access
  token](https://github.com/settings/tokens) with `repo` level access.
* `GITHUB_ORG`: GitHub organization/username.
* `GITHUB_REPO`: GitHub repository name.
* `GITHUB_FILE`: path to a file in the `GITHUB_REPO`, e.g. `api-client.yaml` or
`some/dir/api-client.toml`.
* `DNS_LOCAL_DOMAIN`: domain, e.g. `example.com`, that Dnsmasq will act as an
  authoritative server for. Setting this automatically enables the
  `--expand-hosts` option for Dnsmasq. Default: ``.
* `DNSMASQ_OPTS`: a string of Dnsmasq commandline switches to include when
  starting Dnsmasq. Default: ``.
* `UDM_POLL_INTERVAL`: number of seconds to wait between polls of the UDM-PRO.
Default: `60`.

#### GitHub Download

If `GITHUB_TOKEN` is set the container will attempt to download an `api-client`
configuration file from a private GitHub repository when starting. When setting
`GITHUB_TOKEN`, all other `GITHUB_*` environment variables must also be set or
the download will fail due to malformation of the download request.

The remote configuration file will be downloaded to the working directory of
the container as the remote file name. The easiest configuration is a private
repository that contains a single file: `api-client.yaml`. This will result
in `/app/api-client.yaml` being created and the `api-client` tool automatically
finding the configuration without any additional environment setup. For example,
the above `docker-compose` file could be modified to:

```yaml
services:
  udm-dns:
    image: ghcr.io/jsumners/udm-dns
    container_name: udm-dns
    restart: unless-stopped

    cap_add:
      - NET_ADMIN
    ports:
      - 53:53/tcp
      - 53:53/udp
    dns:
      - 8.8.8.8
      - 8.8.4.4

    environment:
      GITHUB_TOKEN: ghp_super-secret-hash
      GITHUB_ORG: octocat
      GITHUB_REPO: udm-dns-config
      GITHUB_FILE: api-client.yaml
```

### api-client

The `api-client` can be configured with any sort of configuration file supported
by [Viper][viper]. All features of `api-client`, except one, are also
configurable by environment variables. The exception is host aliases (described
below).

A full `api-client` configuration in yaml is:

```yaml
# Specifies the remote IP address of the UDM-PRO.
# Env var: API_CLIENT_ADDRESS
address: 192.168.1.1

# Specifies the username to connect to the UDM-PRO with.
# Env var: API_CLIENT_USERNAME
username: api

# Specifies the password to connect to the UDM-PRO with.
# Env var: API_CLIENT_PASSWORD
password: super-secret

# Indicates if only statically assigned clients should be considered.
# These are clients that have had an IP address assigned to them through the
# Network OS interface instead of randomly assigned from the pool of available
# dynamic IP addresses.
# Env var: API_CLIENT_FIXED_ONLY
# Default: true
fixed_only: true

# Indicates if all discovered hostnames should be lowercased before writing to
# the hosts file that Dnsmasq will read.
# Env var: API_CLIENT_LOWERCASE_HOSTNAMES
# Default: true
lowercase_hostnames: true

# A set of names and IP addresses to add to the hosts file that Dnsmasq will
# read. This allows for setting of names that the UDM-PRO does not manage, For
# example, the UDM-PRO itself.
# This is not configurable via envionment variables.
# Default: an empty list.
host_aliases:
  - name: router
    ip_address: 192.168.1.1
```


[dnsmasq]: https://dnsmasq.org
[udm-pro]: https://store.ui.com/products/udm-pro
[unifi-dns]: https://github.com/wicol/unifi-dns
[docker-compose]: https://docs.docker.com/compose/compose-file/
[viper]: https://github.com/spf13/viper/tree/5247643f02358b40d01385b0dbf743b659b0133f#reading-config-files
