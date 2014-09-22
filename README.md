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

* Consul server node names must be resolvable DNS names
* The servers' `serf_lan` port must be 8301; the port shown in the registered
  `consul` service is the server port, but that's not what agents use to join the
  cluster.  Since the `serf_lan` port is not discoverable, it's hard-coded here.

Per the spec, SRV targets must be A or AAAA records.  Route53 doesn't bitch
about CNAMEs, not sure how it'd feel about IP addresses.  My environment uses
resolvable FQDNs for Consul node names so that's what I'm going with.  This tool
may not be general-purpose enough for some (most? any?) people.

## future features

* health check to ensure the session JSON file exists, to prevent lost sessions

## references

http://en.wikipedia.org/wiki/SRV_record
https://github.com/hashicorp/consul/issues/110

[SRV]: http://en.wikipedia.org/wiki/SRV_record
[leader-elec]: http://www.consul.io/docs/guides/leader-election.html
