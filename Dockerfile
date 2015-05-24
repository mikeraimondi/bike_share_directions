FROM golang:1.4

# Install Node deps
RUN curl -sL https://deb.nodesource.com/setup | bash - \
  && apt-get install -y nodejs \
  && npm install npm -g \
  && adduser --disabled-password --gecos '' node
USER node
COPY package.json /tmp/package.json
RUN cd /tmp && npm install

# Install Bower deps
COPY bower.json /tmp/bower.json
RUN cd /tmp && node_modules/.bin/bower install --config.interactive=false

# Copy deps
USER root
RUN mkdir -p go/src/app \
  && cp -a /tmp/node_modules go/src/app/ \
  && cp -a /tmp/bower_components go/src/app/

WORKDIR /go/src/app

COPY . /go/src/app
RUN go-wrapper download
RUN go-wrapper install
CMD ["go-wrapper", "run"]

ENV SSLPORT=6443
EXPOSE 6443
ENV PORT=6080
EXPOSE 6080
