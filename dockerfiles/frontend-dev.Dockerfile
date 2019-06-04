### frontend-build-stage
FROM node:10

WORKDIR /app

ADD frontend/v1 .

RUN npm install
RUN npm run build

EXPOSE 443 80

ENTRYPOINT [ "npm", "run", "watch" ]
