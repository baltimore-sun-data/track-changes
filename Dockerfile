FROM golang:alpine as go-builder
WORKDIR /go/src/github.com/baltimore-sun-data/track-changes
COPY . .

# Get dep working and install dependencies
RUN apk --no-cache add git
RUN go get -u -v github.com/golang/dep/cmd/dep
RUN dep ensure

RUN CGO_ENABLED=0 GOOS=linux go build -o app .

# Node comes with yarn
FROM node:alpine as yarn-builder
WORKDIR /go/src/github.com/baltimore-sun-data/track-changes
COPY . .

RUN yarn
RUN yarn run build

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=go-builder /go/src/github.com/baltimore-sun-data/track-changes/app .
COPY --from=yarn-builder /go/src/github.com/baltimore-sun-data/track-changes/assets/ assets/

ENV PORT 80
EXPOSE 80
CMD ["./app"]
