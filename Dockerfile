FROM alpine:latest

ARG HUGO_VERSION=0.110.0

RUN wget -O - https://github.com/gohugoio/hugo/releases/download/v${HUGO_VERSION}/hugo_${HUGO_VERSION}_Linux-64bit.tar.gz | tar -xz -C /tmp \
    && mkdir -p /usr/local/sbin \
    && mv /tmp/hugo /usr/local/sbin/hugo \
    && rm -rf /tmp/*

WORKDIR /app
COPY . /app
EXPOSE 1313
ENTRYPOINT ["hugo","server", "--bind","0.0.0.0", "--disableLiveReload", "--disableBrowserError", "-b" ,"" ,"--appendPort=false"]
