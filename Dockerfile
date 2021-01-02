FROM node:12

RUN mkdir -p /app/node && \
    mkdir -p /app/go
    
WORKDIR /app

COPY package*.json /app/node/
COPY * /app/go/

RUN apt-get install -y golang && \
    cd node && \
    npm install . && \
    cd ../go &&
    go get -d -v ./... && \
    go install -v ./...

#if you want to start the service automatically upon docker start, then can run the relevant CMDs.
