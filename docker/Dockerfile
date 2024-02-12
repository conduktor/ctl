FROM golang:1.22
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /conduktor . && rm -rf /app
CMD ["/bin/conduktor"]

FROM scratch
COPY --from=0 /conduktor /bin/conduktor
ENTRYPOINT ["/bin/conduktor"]
