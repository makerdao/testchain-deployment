FROM testchain-deployment-base:latest

ENV TCD_PORT 5001

EXPOSE $TCD_PORT

VOLUME /root/.ssh

COPY certs /etc/ssl/certs/
COPY ./bin/linux-amd64/service /service

ENTRYPOINT ["./service"]