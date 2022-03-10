FROM harbor.cplus.link/base/backend
MAINTAINER ops <ops@abit.com>
EXPOSE 8080
COPY ./process /data/bin/
ENV SERVICENAME="process"
ENTRYPOINT /data/bin/process