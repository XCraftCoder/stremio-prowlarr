FROM golang:1.22-alpine AS build
WORKDIR /src

COPY go.* .
RUN go mod download

COPY . .
RUN go build -o /bin/server /src/cmd/server/main.go

FROM alpine
COPY --from=build /bin/server /bin/server
CMD ["/bin/server"]
