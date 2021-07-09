FROM node:12.18-alpine3.11 as bin
WORKDIR /immune
COPY . .

# Install modules and TSC
RUN npm install --unsafe-perm --production
# Run typescript compiler

ARG NAME
ARG VERSION

ENV NAME=${NAME}
ENV VERSION=${VERSION}
ENV GATEWAY_GRAPH_VARIANT=""
ENV GATEWAY_REGISTRY_URL=""

ENTRYPOINT ["node", "/youla-graph-federation"]
CMD ["--config=/configs/config.json"]
