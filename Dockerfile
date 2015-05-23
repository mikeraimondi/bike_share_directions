FROM golang:1.4

RUN apt-get update
RUN apt-get -y install npm
RUN ln -s /usr/bin/nodejs /usr/bin/node

# Install Node deps
ADD package.json /tmp/package.json
RUN cd /tmp && npm install --unsafe-perm

# Install Bower deps
ADD bower.json /tmp/bower.json
RUN cd /tmp && node_modules/.bin/bower install --allow-root --config.interactive=false

# Copy deps
RUN mkdir -p go/src/app && cp -a /tmp/node_modules go/src/app/ && cp -a /tmp/bower_components go/src/app/

WORKDIR /go/src/app

COPY . /go/src/app
RUN go-wrapper download
RUN go-wrapper install
CMD ["go-wrapper", "run"]

ENV SSLPORT=6443
EXPOSE 6443
ENV PORT=6080
EXPOSE 6080
