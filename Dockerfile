FROM golang:1.14
WORKDIR /go/src/app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN ./env.sh
RUN go build main.go
EXPOSE 8080
CMD ["./setup.sh"]