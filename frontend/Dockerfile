FROM node:12.18-alpine3.11 as bin
WORKDIR /immune
COPY . .

# Install modules and TSC
RUN npm install --unsafe-perm --production
# Run typescript compiler
RUN npm run build-prod

ARG NAME
ARG VERSION

ENV NAME=${NAME}
ENV VERSION=${VERSION}

ENTRYPOINT ["node", "/immune"]
CMD ["--config=/configs/config.json"]
