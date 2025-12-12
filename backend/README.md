<p align="center">
  <a href="http://nestjs.com/" target="blank"><img src="https://nestjs.com/img/logo-small.svg" width="120" alt="Nest Logo" /></a>
</p>

[circleci-image]: https://img.shields.io/circleci/build/github/nestjs/nest/master?token=abc123def456
[circleci-url]: https://circleci.com/gh/nestjs/nest

  <p align="center">A progressive <a href="http://nodejs.org" target="_blank">Node.js</a> framework for building efficient and scalable server-side applications.</p>
    <p align="center">
<a href="https://www.npmjs.com/~nestjscore" target="_blank"><img src="https://img.shields.io/npm/v/@nestjs/core.svg" alt="NPM Version" /></a>
<a href="https://www.npmjs.com/~nestjscore" target="_blank"><img src="https://img.shields.io/npm/l/@nestjs/core.svg" alt="Package License" /></a>
<a href="https://www.npmjs.com/~nestjscore" target="_blank"><img src="https://img.shields.io/npm/dm/@nestjs/common.svg" alt="NPM Downloads" /></a>
<a href="https://circleci.com/gh/nestjs/nest" target="_blank"><img src="https://img.shields.io/circleci/build/github/nestjs/nest/master" alt="CircleCI" /></a>
<a href="https://discord.gg/G7Qnnhy" target="_blank"><img src="https://img.shields.io/badge/discord-online-brightgreen.svg" alt="Discord"/></a>
<a href="https://opencollective.com/nest#backer" target="_blank"><img src="https://opencollective.com/nest/backers/badge.svg" alt="Backers on Open Collective" /></a>
<a href="https://opencollective.com/nest#sponsor" target="_blank"><img src="https://opencollective.com/nest/sponsors/badge.svg" alt="Sponsors on Open Collective" /></a>
  <a href="https://paypal.me/kamilmysliwiec" target="_blank"><img src="https://img.shields.io/badge/Donate-PayPal-ff3f59.svg" alt="Donate us"/></a>
    <a href="https://opencollective.com/nest#sponsor"  target="_blank"><img src="https://img.shields.io/badge/Support%20us-Open%20Collective-41B883.svg" alt="Support us"></a>
  <a href="https://twitter.com/nestframework" target="_blank"><img src="https://img.shields.io/twitter/follow/nestframework.svg?style=social&label=Follow" alt="Follow us on Twitter"></a>
</p>
  <!--[![Backers on Open Collective](https://opencollective.com/nest/backers/badge.svg)](https://opencollective.com/nest#backer)
  [![Sponsors on Open Collective](https://opencollective.com/nest/sponsors/badge.svg)](https://opencollective.com/nest#sponsor)-->

## Description

Monad Flow backend service built with NestJS. It ingests metrics and traces from the sidecars (network/system) and exposes APIs for the frontend.

---

## 1. MongoDB & Docker Compose

The backend requires a MongoDB instance. Before starting MongoDB with Docker Compose, you must define the required environment variables.

### 1.1 Environment variables

Define the following environment variables in a `.env` file at the **backend project root** (or in your deployment environment):

```
MONGO_ROOT_USERNAME=root
MONGO_ROOT_PASSWORD=<password>
MONGO_DATABASE=<database-name>
```

- `MONGO_ROOT_USERNAME` / `MONGO_ROOT_PASSWORD`: credentials for the MongoDB root user.
- `MONGO_DATABASE`: default database used by the backend.

These values must be set as environment variables (for example via a `.env` file, `export` in your shell, or your process manager configuration) **before** you start MongoDB with Docker Compose and run the backend.

### 1.2 Start MongoDB with Docker Compose

From the `backend` directory:

```bash
cd backend
docker compose up -d
```

This will start MongoDB (and any other services defined in `docker-compose.yml`) in the background.

---

## 2. Project setup

Install dependencies:

```bash
cd backend
npm install
```

---

## 3. Build & run

### 3.1 Build

```bash
cd backend
npm run build
```

### 3.2 Run with Node

```bash
cd backend

# development
npm run start

# watch mode
npm run start:dev

# production mode (compiled)
npm run start:prod
```

### 3.3 Run with PM2

To run the backend as a managed process using PM2:

```bash
# install pm2 globally (once)
npm install -g pm2

cd backend

# production run using the compiled dist/main.js
pm2 start dist/main.js --name monad-flow-backend
```

---

## License

Nest is [MIT licensed](https://github.com/nestjs/nest/blob/master/LICENSE).
