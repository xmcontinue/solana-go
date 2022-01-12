FROM harbor.cplus.link/base/backend
MAINTAINER ops <ops@abit.com>
EXPOSE 8080
COPY ./wacher /data/bin/
ENV SERVICENAME="wacher"
ENTRYPOINT /data/bin/wacher