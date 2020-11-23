FROM golang:latest AS builder
WORKDIR /go/src/grafzahl
COPY main.go .
RUN go get -v && CGO_ENABLED=0 go build -o /bin/grafzahl
FROM scratch
COPY --from=builder /bin/grafzahl /bin/grafzahl
EXPOSE 9696
ENV GRAFZAHL_PASSWORD=""
ENV GRAFZAHL_USERNAME=""
ENTRYPOINT ["/bin/grafzahl"]
