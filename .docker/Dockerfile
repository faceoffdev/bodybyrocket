ARG TDLIB_HASH=${TDLIB_HASH:-"22d49d5"}
ARG GO_VERSION=${GO_VERSION:-"1.23"}
ARG GO_ALPINE_VERSION=${GO_ALPINE_VERSION:-"3.20"}

FROM ghcr.io/zelenin/tdlib-docker:${TDLIB_HASH}-alpine AS tdlib

FROM golang:${GO_VERSION}-alpine${GO_ALPINE_VERSION} AS golang

# Пакеты необходимые для Oh My Bash
RUN apk add --no-cache --update --repository=https://dl-cdn.alpinelinux.org/alpine/edge/community \
    curl \
    git \
    bash

# Дополнительные пакеты
RUN apk add --no-cache --update --repository=https://dl-cdn.alpinelinux.org/alpine/edge/community \
    nano \
    make

# Пакеты для tdlib
RUN apk add --no-cache --update --repository=https://dl-cdn.alpinelinux.org/alpine/edge/community \
    build-base \
    ca-certificates \
    linux-headers \
    openssl-dev \
    zlib-dev

# Настраиваем Europe/Moscow
RUN apk add --no-cache --update --repository=https://dl-cdn.alpinelinux.org/alpine/edge/community \
    tzdata

ENV TZ=Europe/Moscow

# Создаем базовую директорию и удаляем кэши, логи и базовые файлы
RUN mkdir -p /app \
    && rm -rf /tmp/* \
    /var/{cache,log}/* \
    /etc/crontabs/*

# Создаем базовую директорию и выдаем новому пользователю права
ARG UID=1000
ARG GID=1000
ENV USER='application'
ENV WORKDIR='/app'

RUN addgroup --gid ${GID} --system ${USER} \
    && adduser --uid ${UID} --system --ingroup ${USER} ${USER} \
    && chown -R ${USER}:${USER} ${WORKDIR} /home/${USER}

RUN chown -R ${USER}:${USER} /app \
    && chown -R ${USER}:${USER} /run \
     && chown -R ${USER}:${USER} /go

USER ${USER}

WORKDIR ${WORKDIR}

COPY --from=tdlib /usr/local/include/td /usr/local/include/td/
COPY --from=tdlib /usr/local/lib/libtd* /usr/local/lib/


FROM golang AS golang_prod

USER root

RUN apk add --no-cache --update --repository=https://dl-cdn.alpinelinux.org/alpine/edge/community \
    tini \
    flock \
    supercronic \
    supervisor

COPY --link .docker/rootfs /

RUN chown -R ${USER}:${USER} /etc/crontabs \
    && chown -R ${USER}:${USER} /etc/supervisor

USER ${USER}

COPY --chown=${USER}:${USER} . /app/src
RUN cd /app/src/ \
    && go mod download \
    && go build -buildvcs=false -gcflags "all=-N -l" -o /app/bbr . \
    && rm -rf /app/src

ENTRYPOINT ["/sbin/tini", "--", "/usr/bin/docker-entrypoint.sh"]
CMD ["/usr/bin/supervisord", "-c", "/etc/supervisord.conf"]


FROM golang AS golang_dev

RUN go install github.com/go-delve/delve/cmd/dlv@latest
