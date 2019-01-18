FROM bash:5.0.0

ENV TCD_PORT 5001

EXPOSE $TCD_PORT

RUN mkdir /deployment && \
    mkdir /downloaded

COPY certs /etc/ssl/certs/
COPY ./bin/linux-amd64/service /service

ENTRYPOINT ["./service"]