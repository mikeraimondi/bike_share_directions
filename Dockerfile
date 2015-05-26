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
ENV project /go/src/github.com/mikeraimondi/bike_share_directions
RUN mkdir -p $project \
  && cp -a /tmp/node_modules $project \
  && cp -a /tmp/bower_components $project

WORKDIR $project
COPY . $project
RUN go install

ENTRYPOINT ["/go/bin/bike_share_directions"]

EXPOSE 80
EXPOSE 443
