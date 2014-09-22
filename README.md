https://github.com/hashicorp/consul/issues/110
http://en.wikipedia.org/wiki/SRV_record

Per the spec, SRV targets must be A or AAAA records.  Route53 doesn't bitch
about CNAMEs, not sure how it'd feel about IP addresses.  My environment uses
resolvable FQDNs for Consul node names so that's what I'm going with.  This tool
may not be general-purpose enough for some (most? any?) people.
