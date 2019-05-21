FROM node:latest AS frontend-build-stage

WORKDIR /app

ADD frontend .

RUN npm install && npm run build

# COPY --from=frontend-build-stage /app/dist /destination
