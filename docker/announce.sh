#!/bin/bash

set -Eeuo pipefail

test_state="withdrawn"

if [ -f "/exabgp/test_state" ]; then
    test_state="$(cat /exabgp/test_state)"
fi

if [ "${test_state}" == "withdrawn" ]; then
    # announce 192.168.88.0/24 as 32 /29s
    echo "neighbor 127.0.0.1 announce route 192.168.88.0/24 next-hop 192.168.1.2 split /29" \
        | tee /exabgp/exabgp.cmd

    # announce single ipv4 route
    echo "neighbor 127.0.0.1 announce route 192.168.0.0/24 next-hop 192.168.1.2" \
        | tee /exabgp/exabgp.cmd

    # announce single ipv4 route with as-path, community, localpref and MED
    echo "neighbor 127.0.0.1 announce route 10.0.0.0/24 next-hop 192.168.1.2 as-path [ 65001 ] community [ 65001:1234 ] local-preference 100 med 200" \
        | tee /exabgp/exabgp.cmd

    # announce single ipv4 route with multiple as-paths and communities
    echo "neighbor 127.0.0.1 announce route 10.0.1.0/24 next-hop 192.168.1.2 as-path [ 65001 65002 ] community [ 65001:1234 65001:5678 ] extended-community [ target:54591:6 l2info:19:0:1500:111 ]" \
        | tee /exabgp/exabgp.cmd

    # announce single ipv6 route
    echo "neighbor 127.0.0.1 announce route 2001:db8:1000::/64 next-hop 2001:db8:ffff::1" \
        | tee /exabgp/exabgp.cmd

    # announce single ipv6 route with as-path, community, localpref and MED
    echo "neighbor 127.0.0.1 announce route 2001:db8:2000::/64 next-hop 2001:db8:ffff::1 as-path [ 65001 ] community [ 65001:1234 ] local-preference 100 med 200" \
        | tee /exabgp/exabgp.cmd

    # announce single ipv6 route with multiple as-pathes and communities
    echo "neighbor 127.0.0.1 announce route 2001:db8:3000::/64 next-hop 2001:db8:ffff::1 as-path [ 65001 65002 ] community [ 65001:1234 65001:5678 ] extended-community [ target:54591:6 l2info:19:0:1500:111 ]" \
        | tee /exabgp/exabgp.cmd

    echo "announced" > "/exabgp/test_state"

    sleep 10
fi
