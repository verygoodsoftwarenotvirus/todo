### frontend-build-stage
FROM node:alpine

WORKDIR /app

ADD frontend .

RUN npm install
RUN npm run build

EXPOSE 443 80

ENTRYPOINT [ "npm", "run", "watch" ]
