FROM golang:1.13-alpine AS dev
WORKDIR /
COPY scripts/setup.sh /setup.sh
RUN /setup.sh
RUN rm setup.sh
COPY .bashrc /root/.bashrc
ENTRYPOINT ["/bin/bash"]
