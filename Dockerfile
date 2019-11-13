FROM golang:1.9-stretch

# install nodejs via nvm

# Replace shell with bash so we can source files
RUN rm /bin/sh && ln -s /bin/bash /bin/sh

# Install base dependencies
RUN apt-get update && apt-get install -y -q --no-install-recommends \
        apt-transport-https \
        build-essential \
        ca-certificates \
        curl \
        git \
        libssl-dev \
        wget

ENV NVM_DIR /usr/local/nvm
ENV NODE_VERSION 6.11.1

WORKDIR $NVM_DIR

RUN curl https://raw.githubusercontent.com/creationix/nvm/master/install.sh | bash \
    && . $NVM_DIR/nvm.sh \
    && nvm install $NODE_VERSION \
    && nvm alias default $NODE_VERSION \
    && nvm use default

ENV NODE_PATH $NVM_DIR/versions/node/v$NODE_VERSION/lib/node_modules
ENV PATH      $NVM_DIR/versions/node/v$NODE_VERSION/bin:$PATH


# 
ADD . /go/src/bitbucket.org/softwarehouseio/victory

WORKDIR /go/src/bitbucket.org/softwarehouseio/victory/victory-frontend

RUN tar xfz gopath_src.tar.gz -C /go/

RUN go-wrapper download

RUN apt-get install -y -q --no-install-recommends \
	libpng-dev

RUN npm install

