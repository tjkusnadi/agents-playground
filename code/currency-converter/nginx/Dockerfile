FROM node:20-alpine AS frontend-builder
WORKDIR /app

COPY code/currency-converter/frontend/package.json ./
RUN npm install

COPY code/currency-converter/frontend ./
RUN npm run build

FROM nginx:1.25-alpine
COPY code/currency-converter/nginx/default.conf /etc/nginx/conf.d/default.conf
COPY --from=frontend-builder /app/dist /usr/share/nginx/html

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
