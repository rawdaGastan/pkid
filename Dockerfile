FROM golang:1.19-alpine

# Create app directory
WORKDIR /usr/src/app

# Install app packages
COPY . .
RUN apk add --update gcc musl-dev
RUN go mod download

# Build app
RUN go build -o /pkid

EXPOSE 3000

CMD [ "/pkid" ]
