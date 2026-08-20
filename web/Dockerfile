FROM node:24-alpine

ENV NEXT_TELEMETRY_DISABLED=1
ARG NEXT_PUBLIC_APP_ENV=production
ENV NEXT_PUBLIC_APP_ENV=$NEXT_PUBLIC_APP_ENV

WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci

COPY . .

RUN npm run build

EXPOSE 3010

CMD ["npm", "run", "start"]
