FROM harbor.cplus.link/base/backend
MAINTAINER ops <ops@abit.com>
EXPOSE 8080
COPY ./watcher /data/bin/
ENV SERVICENAME="watcher"
ENTRYPOINT /data/bin/watcher