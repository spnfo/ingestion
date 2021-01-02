FROM node:12

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
