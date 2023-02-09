FROM harbor.cplus.link/base/backend
MAINTAINER ops <ops@abit.com>
RUN update-ca-certificates
EXPOSE 8080
COPY ./process2 /data/bin/
ENV SERVICENAME="process2"
ENTRYPOINT /data/bin/process2