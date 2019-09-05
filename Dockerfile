FROM golang:1.12.9

WORKDIR /app
# Just add the binary
ADD main /app/
ENTRYPOINT ["./main"]