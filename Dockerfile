FROM alpine:latest

COPY ./andromedad /usr/local/bin/andromedad

CMD andromedad
