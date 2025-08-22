FROM bluenviron/mediamtx:latest AS mtx

FROM alpine:3.20
RUN apk add --no-cache ffmpeg ca-certificates tzdata
COPY --from=mtx /mediamtx /mediamtx
COPY mediamtx.yml /mediamtx.yml
EXPOSE 8554 8889 9997
EXPOSE 8189/udp 8890/udp
ENTRYPOINT ["/mediamtx"]