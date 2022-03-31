FROM harbor.cplus.link/base/backend
MAINTAINER ops <ops@abit.com>
EXPOSE 8080
COPY ./exchange /data/bin/
ENV SERVICENAME="exchange"
ENTRYPOINT /data/bin/exchange