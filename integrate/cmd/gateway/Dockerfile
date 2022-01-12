FROM harbor.cplus.link/base/backend
MAINTAINER ops <ops@abit.com>
EXPOSE 8080
ADD ./gateway /data/bin/gateway
ENV SERVICENAME="gateway"
ENTRYPOINT /data/bin/gateway
