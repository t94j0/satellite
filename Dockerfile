FROM golang:1.13.1 as builder
WORKDIR /go/src/github.com/t94j0/satellite
COPY . .
RUN cd satellite && CGO_ENABLED=0 GOOS=linux go build -a  -o /root/satellite .


FROM alpine:latest  
RUN apk --no-cache add ca-certificates openssl
# Configure satellite
## Run postinstall.sh
COPY ./.config/scripts/postinstall.sh /
RUN sh /postinstall.sh
## Merge .config with filesystem
RUN mkdir -p /etc/satellite /var/lib/satellite
COPY ./.config/etc/satellite/config.yml /etc/satellite/
COPY ./.config/var/lib/satellite/GeoLite2-Country.mmdb /var/lib/satellite/

WORKDIR /root/
COPY --from=builder /root/satellite .
RUN ls -la /root
EXPOSE 443
CMD ["/root/satellite"]
