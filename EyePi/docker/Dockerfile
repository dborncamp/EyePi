FROM ubuntu

COPY ./settings.toml /etc/EyePi/settings.toml
COPY ./EyePi /usr/bin/EyePi
COPY ./entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]

EXPOSE 10000
EXPOSE 10001
EXPOSE 8080
