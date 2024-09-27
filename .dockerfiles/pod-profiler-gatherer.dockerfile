FROM alpine:3.20.3

RUN mkdir pod-profiler-gatherer
RUN mkdir pod-profiler-gatherer/bin
RUN mkdir pod-profiler-gatherer/logs
RUN mkdir pod-profiler-gatherer/config

WORKDIR /pod-profiler-gatherer

COPY bin/linux/amd64/pod-profiler-gatherer ./pod-profiler-gatherer

ENTRYPOINT ["/pod-profiler-gatherer/pod-profiler-gatherer"]
