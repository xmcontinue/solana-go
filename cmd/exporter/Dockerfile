FROM harbor.cplus.link/base/backend
MAINTAINER ops <ops@abit.com>
EXPOSE 8080
COPY ./exporter /data/bin/
ENV SERVICENAME="exporter"
ENTRYPOINT /data/bin/exporter