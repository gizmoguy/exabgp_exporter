#!/bin/bash

set -Eeuo pipefail

test_state="withdrawn"

if [ -f "/exabgp/test_state" ]; then
    test_state="$(cat /exabgp/test_state)"
fi

if [ "${test_state}" == "announced" ]; then
    echo "neighbor 127.0.0.1 withdraw route 192.168.88.0/24 next-hop 192.168.1.2 split /29" \
        | tee /exabgp/exabgp.cmd
    echo "neighbor 127.0.0.1 withdraw route 192.168.0.0/24 next-hop 192.168.1.2" \
        | tee /exabgp/exabgp.cmd
    echo "neighbor 127.0.0.1 withdraw route 10.0.0.0/24 next-hop 192.168.1.2 as-path [ 65001 ] community [ 65001:1234 ] local-preference 100 med 200" \
        | tee /exabgp/exabgp.cmd
    echo "neighbor 127.0.0.1 withdraw route 10.0.1.0/24 next-hop 192.168.1.2 as-path [ 65001 65002 ] community [ 65001:1234 65001:5678 ] extended-community [ target:54591:6 l2info:19:0:1500:111 ]" \
        | tee /exabgp/exabgp.cmd
    echo "neighbor 127.0.0.1 withdraw route 2001:db8:1000::/64 next-hop 2001:db8:ffff::1" \
        | tee /exabgp/exabgp.cmd
    echo "neighbor 127.0.0.1 withdraw route 2001:db8:2000::/64 next-hop 2001:db8:ffff::1 as-path [ 65001 ] community [ 65001:1234 ] local-preference 100 med 200" \
        | tee /exabgp/exabgp.cmd
    echo "neighbor 127.0.0.1 withdraw route 2001:db8:3000::/64 next-hop 2001:db8:ffff::1 as-path [ 65001 65002 ] community [ 65001:1234 65001:5678 ] extended-community [ target:54591:6 l2info:19:0:1500:111 ]" \
        | tee /exabgp/exabgp.cmd

    echo "withdrawn" > "/exabgp/test_state"

    sleep 10
fi
