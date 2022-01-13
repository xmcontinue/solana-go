FROM harbor.cplus.link/base/backend
MAINTAINER ops <ops@abit.com>
EXPOSE 8080
COPY ./market /data/bin/
ENV SERVICENAME="market"
ENTRYPOINT /data/bin/market