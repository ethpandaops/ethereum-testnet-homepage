# ethereum-testnet-homepage

A homepage that can be used to show important links and documentation for a given testnet.


## Running with docker

(Optional) Building with docker:

```sh
docker build -t skylenet/ethereum-testnet-homepage .
```

The example below shows you how to overwrite the `config.yaml` and how to set your custom .md file to be rendered on the page.

```sh
docker run -it --rm --name testnet-homepage \
           -p 1313:1313 \
           -v $PWD/config.yaml:/app/config.yaml \
           -v $PWD/custom-md-example.md:/app/layouts/partials/custom.md \
           skylenet/ethereum-testnet-homepage
```

## License

[MIT License](LICENSE)
