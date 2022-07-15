FROM  golang:1.18.4-bullseye

ADD . /work

WORKDIR /work

RUN make && mv /work/bin/initc /work/initc

ENTRYPOINT ["/work/initc"]