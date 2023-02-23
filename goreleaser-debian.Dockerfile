FROM debian:latest
COPY ethereum-testnet-homepage* /ethereum-testnet-homepage
ENTRYPOINT ["/ethereum-testnet-homepage"]
