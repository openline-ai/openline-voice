FROM node:16.16-alpine as deps
RUN apk add --no-cache libc6-compat
WORKDIR /app
COPY package.json yarn.lock ./
RUN yarn install --frozen-lockfile

FROM node:16.16-alpine AS runner

WORKDIR /app

COPY --from=deps /app/node_modules ./node_modules

COPY . .

CMD ["node", "src/server.js"]