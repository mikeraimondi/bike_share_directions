FROM golang:onbuild

RUN apt-get update
RUN apt-get -y install npm
RUN ln -s /usr/bin/nodejs /usr/bin/node

RUN npm install --unsafe-perm
RUN node_modules/.bin/bower install --allow-root --config.interactive=false

ENV SSLPORT=6443
EXPOSE 6443
ENV PORT=6080
EXPOSE 6080
