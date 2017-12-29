FROM golang:alpine as go-builder

# Get dep working and install dependencies
RUN apk --no-cache add git
RUN go get -u -v github.com/golang/dep/cmd/dep

# Separate vendor fetching step for better caching
COPY Gopkg.lock Gopkg.toml /go/src/github.com/baltimore-sun-data/track-changes/
WORKDIR /go/src/github.com/baltimore-sun-data/track-changes
RUN dep ensure -vendor-only

# Build Go binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-X 'main.applicationBuildDate=`date`'" \
    -o app .

# Node comes with yarn
FROM node:alpine as yarn-builder
WORKDIR /go/src/github.com/baltimore-sun-data/track-changes
COPY package.json yarn.lock ./
RUN yarn

COPY . .
RUN yarn run build

# Actual final image
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY templates/ templates/
COPY --from=go-builder /go/src/github.com/baltimore-sun-data/track-changes/app .
COPY --from=yarn-builder /go/src/github.com/baltimore-sun-data/track-changes/assets/ assets/

# Mount volume containing config/secrets
VOLUME [ "/var/track-changes" ]
ENV ENV_FILE /var/track-changes/track-changes-prod.json

# Set port
ENV PORT 80
EXPOSE 80

CMD [ "./app" ]
