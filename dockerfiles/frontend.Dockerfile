FROM node:latest AS frontend-build-stage

WORKDIR /app

ADD frontend/v1 .

RUN npm install && npm run build

# COPY --from=frontend-build-stage /app/dist /destination
