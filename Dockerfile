FROM alpine

ENV TIME_ZONE=Asia/Shanghai
RUN mkdir /lib64 && \
    ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld.linux-x86_64.so.2 && \
    ln -snf /usr/share/zoneinfo/$TIME_ZONE /etc/localtime && \
    echo $TIME_ZONE > /etc/timezone

COPY bin/linux/pwc-controller /pwc-controller
COPY scripts/ /scripts/
COPY conf/config-example.json /etc/xwc-controller/config.json

EXPOSE 7000

ENTRYPOINT ["/pwc-contoller"]