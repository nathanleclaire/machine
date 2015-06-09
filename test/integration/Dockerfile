FROM debian:jessie

RUN apt-get update && \
    apt-get install -y git curl openssl
RUN git clone https://github.com/sstephenson/bats.git && \
    cd bats && \
    ./install.sh /usr/local
RUN curl -SsL https://get.docker.com/builds/Linux/x86_64/docker-latest >/usr/local/bin/docker
RUN chmod +x /usr/local/bin/docker
WORKDIR /integration
ENTRYPOINT ["/integration/run-bats.sh"]
ADD . /integration
RUN mv /integration/docker-machine_linux-amd64 /docker-machine_linux-amd64
