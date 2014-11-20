# consul-srv-updater

A somewhat-opinionated utility for managing DNS [SRV records][SRV] in Route53
for your Consul cluster.

## usage

Install this utility on one or more hosts in your Consul cluster; probably on
each of your Consul servers.  Schedule it to run via cron.

    consul-srv-updater \
        --log-file=/var/log/consul-srv-updater.log \
        --data-dir=/var/lib/consul-srv-updater \
        --zone=your_zone_id \
        --name=_consul._tcp.your_domain.com \
        --ttl=60

To use the SRV record when joining an agent to the cluster, use `dig`:

    dig +search +noall +answer _consul._tcp.your_domain.com SRV | \
        awk '{printf("%s:%d\n", $8, $7)}' | \
        xargs --no-run-if-empty consul join

## description

This tool stores the nodes providing the `consul` service as targets in a SRV
record.  This name can then be used as a coordination point for agents that need
to join the cluster for the first time.

Because it uses Consul's [leader election strategy][leader-elec], you don't have
to worry about updates occurring multiple times from multiple servers.  I run it
on all of my Consul servers via cron.

## limitations

aka "opinions"

* The servers' `serf_lan` port must be 8301; the port shown in the registered
  `consul` service is the server port, but that's not what agents use to join the
  cluster.  Since the `serf_lan` port is not discoverable, it's hard-coded here.

Per the spec, SRV targets must be domain names.  We use private IPs here to cope
with environments where you can't use a single DNS record to resolve a public IP
or private IP, depending on the lookup context - a la DNS in EC2.  You'll have to
deal with this in your tooling by watching out for trailing periods on A/AAAA
records.  We may or may not provide a flag in the future to toggle between the node
(FQDN) and address (listening IP).

## building

You need Go.  The included `Makefile` will create `stage/consul-srv-updater` for
you.

## future features

* health check to ensure the session JSON file exists, to prevent lost sessions

## references

* http://en.wikipedia.org/wiki/SRV_record
* https://github.com/hashicorp/consul/issues/110

[SRV]: http://en.wikipedia.org/wiki/SRV_record
[leader-elec]: http://www.consul.io/docs/guides/leader-election.html
