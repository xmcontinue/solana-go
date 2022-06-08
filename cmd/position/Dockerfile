FROM harbor.cplus.link/base/backend
MAINTAINER ops <ops@abit.com>
EXPOSE 8080
COPY ./position /data/bin/
ENV SERVICENAME="position"
ENTRYPOINT /data/bin/position
