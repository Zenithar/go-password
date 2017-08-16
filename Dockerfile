FROM        frolvlad/alpine-glibc:latest
MAINTAINER  Thibault NORMAND <me@zenithar.org>

COPY bin/password_server /usr/bin/password_server
COPY entrypoint.sh /

RUN apk add --no-cache su-exec tini \
    && chmod +x /usr/bin/password_server \
    && chmod +x /entrypoint.sh \
    && addgroup tokenizr \
    && adduser -s /bin/false -G nogroup -S -D nobody

EXPOSE     5555
WORKDIR    /srv
ENTRYPOINT [ "/entrypoint.sh" ]
CMD        [ "app:help" ]
