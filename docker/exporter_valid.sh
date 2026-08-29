#!/command/with-contenv bash
# this is an `standalone` mode exporter that should be able to successfully call exabgpcli listening on a different port
if [[ "${EXABGP_VERSION}" =~ ^5 ]]; then
    exabgpcli="exabgp-cli"
else
    exabgpcli="exabgpcli"
fi
exec /exabgp/exabgp_exporter --web.listen-address=":9570" --log.format="json" standalone --exabgp.cli.command="/exabgp/venv/bin/${exabgpcli}" --exabgp.root="/exabgp/"
