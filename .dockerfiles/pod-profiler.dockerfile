FROM alpine:3.20.3

RUN mkdir pod-profiler
RUN mkdir pod-profiler/bin
RUN mkdir pod-profiler/logs
RUN mkdir pod-profiler/config

WORKDIR /pod-profiler

COPY bin/linux/amd64/pod-profiler ./pod-profiler

ENTRYPOINT ["/pod-profiler/pod-profiler"]
