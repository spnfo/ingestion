FROM node:12

#FROM golang:latest #The golang dockerhub image, in case you want to split the two out

RUN mkdir -p /app/js && \
    mkdir -p /app/go
    
WORKDIR /app

COPY package*.json /app/js/
COPY * /app/go/

RUN apt-get install -y golang && \
    cd js && \
    npm install . && \
    cd ../go &&
    go get -d -v ./... && \
    go install -v ./...

#if you want to start the service automatically upon docker start, then can run the relevant CMDs.
