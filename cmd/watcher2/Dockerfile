FROM harbor.cplus.link/base/backend
MAINTAINER ops <ops@abit.com>
EXPOSE 8080
COPY ./watcher2 /data/bin/
ENV SERVICENAME="watcher2"
ENTRYPOINT /data/bin/watcher2