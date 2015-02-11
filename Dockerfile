FROM golang:onbuild

RUN apt-get update
RUN apt-get -y install npm
RUN ln -s /usr/bin/nodejs /usr/bin/node

RUN npm install --unsafe-perm
RUN node_modules/.bin/bower install --allow-root --config.interactive=false

ENV PORT=8080
EXPOSE 8080
