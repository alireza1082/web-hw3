FROM node:alpine
WORKDIR /usr/src/app
RUN pwd
COPY package*.json ./
RUN npm install
COPY . .
CMD ["node", "main.js"]