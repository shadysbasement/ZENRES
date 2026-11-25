# Builds one SvelteKit app from the monorepo: --build-arg APP=web-public | web-app
FROM node:24-alpine AS build
ARG APP
WORKDIR /repo
COPY package.json package-lock.json ./
COPY packages ./packages
COPY apps/web-public/package.json apps/web-public/
COPY apps/web-app/package.json apps/web-app/
RUN npm ci
COPY apps/${APP} apps/${APP}
RUN npm run build -w @zevra/${APP}

FROM node:24-alpine
ARG APP
WORKDIR /app
ENV NODE_ENV=production
COPY --from=build /repo/apps/${APP}/build ./build
COPY --from=build /repo/apps/${APP}/package.json ./
COPY --from=build /repo/node_modules ./node_modules
USER node
EXPOSE 3000
CMD ["node", "build"]
