#!/command/with-contenv bash
if [[ "${EXABGP_VERSION}" =~ ^5 ]]; then
    export EXABGP_ROOT=/exabgp/
    exec /exabgp/venv/bin/exabgp server /exabgp/etc/exabgp/exabgp.conf
else
    exec /exabgp/venv/bin/exabgp --root=/exabgp/ /exabgp/etc/exabgp/exabgp.conf
fi
