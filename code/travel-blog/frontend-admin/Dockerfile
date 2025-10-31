# syntax=docker/dockerfile:1

FROM nginx:1.25-alpine
WORKDIR /usr/share/nginx/html
COPY index.html ./
COPY styles.css ./
COPY main.js ./
COPY nginx.conf /etc/nginx/conf.d/default.conf
