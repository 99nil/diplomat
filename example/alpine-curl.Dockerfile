# docker buildx build --push --platform linux/arm64,linux/amd64,linux/arm,linux/ppc64le,linux/s390x \
# -t $(REPO)/alpine-curl:$(TAG) -f alpine-curl.Dockerfile .
FROM alpine:latest

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories && \
    apk update && \
    apk add --no-cache curl

CMD ["tail", "-f", "/dev/null"]