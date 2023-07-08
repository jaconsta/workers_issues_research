FROM node:20-bullseye as base
RUN groupadd docker
RUN useradd -r -u 1001 -g docker user
USER user
WORKDIR /home/user
COPY --chown=user:docker tsconfig.json package.json package-lock.json ./
RUN npm ci --no-audit --no-fund

# FROM base as development
COPY --chown=user:docker src ./src
CMD npx nodemon src/index.ts

