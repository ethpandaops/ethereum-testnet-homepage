FROM gcr.io/distroless/static-debian11:latest
COPY ethereum-testnet-homepage* /ethereum-testnet-homepage
ENTRYPOINT ["/ethereum-testnet-homepage"]
