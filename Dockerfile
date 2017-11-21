FROM golang:latest as builder
WORKDIR /go/src/github.com/baltimore-sun-data/track-changes
COPY . .

# Make Yarn installable
RUN curl -sL https://deb.nodesource.com/setup_8.x | bash -
RUN apt-get update
RUN apt-get install -y --no-install-recommends \
    nodejs \
    build-essential

RUN go get -u -v github.com/golang/dep/cmd/dep
RUN npm install -g yarn
RUN dep ensure
RUN yarn
RUN yarn run build
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /go/src/github.com/baltimore-sun-data/track-changes/app .
COPY --from=builder /go/src/github.com/baltimore-sun-data/track-changes/assets/ assets/

ENV PORT 80
EXPOSE 80
CMD ./app
