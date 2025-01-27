FROM node:12

#FROM golang:latest #The golang dockerhub image, in case you want to split the two out

RUN mkdir -p /app/
    
WORKDIR /app

#since it's a private repo, there may need to be some extra configuration here. Or if you've already cloned according to the directories here and are building locally, then this can be commented out
#RUN git clone --depth=1 https://github.com/spnfo/ingestion ./app

COPY package*.json /app/js/
COPY * /app/go/

RUN apt-get install -y golang && \
    cd js && \
    npm install . && \
    cd ../go &&
    go get -d -v ./... && \
    go install -v ./... && \
    cd ../

#these are more for port documentation purposes - they serve no real function. To actually publish the ports, it's the -p flag in docker run
EXPOSE 8000
EXPOSE 8001

#if you want to start the service automatically upon docker start, then can run the relevant CMDs
CMD ["node",".js/index.js","&&","go","run","./go/server/server.go"]
