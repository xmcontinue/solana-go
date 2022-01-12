FROM harbor.cplus.link/base/backend
MAINTAINER ops <ops@abit.com>
EXPOSE 8080
COPY ./tool /data/bin/
ENV SERVICENAME="tool"
ENTRYPOINT /data/bin/tool