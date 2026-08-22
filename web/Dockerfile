FROM node:22.23.2-alpine@sha256:c610fcdfb1d5b4740dd70c284ed3cb16bb857e0f7166196e36a5501df7a3aa32

ENV NEXT_TELEMETRY_DISABLED=1
ARG NEXT_PUBLIC_APP_ENV=production
ARG NEXT_PUBLIC_PASSAGE_MODE=jwks
ENV NEXT_PUBLIC_APP_ENV=$NEXT_PUBLIC_APP_ENV NEXT_PUBLIC_PASSAGE_MODE=$NEXT_PUBLIC_PASSAGE_MODE

WORKDIR /app

RUN corepack enable

COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY web/package.json ./web/package.json
RUN pnpm install --frozen-lockfile --filter @heard/web...

COPY web ./web
WORKDIR /app/web

RUN pnpm build

EXPOSE 3010

CMD ["pnpm", "start"]
