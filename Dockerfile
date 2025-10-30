FROM golang:1.25.1 as build
ARG version=unknown
ARG hash=unknown
ARG ldflags=""
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="$ldflags -X 'github.com/conduktor/ctl/utils.version=$version' -X 'github.com/conduktor/ctl/utils.hash=$hash'" -o /conduktor . && rm -rf /app
CMD ["/bin/conduktor"]

FROM alpine:3.19
RUN adduser -D conduktor
USER conduktor
COPY --from=build /conduktor /bin/conduktor
ENTRYPOINT ["/bin/conduktor"]
