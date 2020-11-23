FROM golang:alpine AS builder
WORKDIR /go/src/grafzahl
RUN apk --no-cache add ca-certificates git
COPY main.go .
RUN go get -v && CGO_ENABLED=0 go build -o /bin/grafzahl
FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /bin/grafzahl /bin/grafzahl
EXPOSE 6969
ENV GRAFZAHL_PASSWORD=""
ENV GRAFZAHL_USERNAME=""
ENTRYPOINT ["/bin/grafzahl"]
