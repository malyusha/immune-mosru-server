FROM nginx:stable

ARG CERTBOT_EMAIL=info@domain.com
ARG DOMAIN_LIST

COPY envsubst-templates.sh /usr/bin/envsubst-templates.sh
COPY gen-certs.sh /usr/bin/gen-certs.sh

RUN apt-get update
RUN apt-get install -y cron bash curl
RUN curl -L https://github.com/a8m/envsubst/releases/download/v1.2.0/envsubst-Linux-x86_64 -o envsubst
RUN chmod +x envsubst && mv envsubst /usr/bin/envsubst-go

# install all required utils
RUN sh -c "CERTBOT_EMAIL=${CERTBOT_EMAIL} DOMAIN_LIST=${DOMAIN_LIST} gen-certs.sh"

VOLUME /etc/letsencrypt

CMD [ "sh", "-c", "envsubst-templates.sh && cron && nginx -g 'daemon off;'" ]