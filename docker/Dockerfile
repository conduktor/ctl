FROM golang:1.22
ARG version=unknown
ARG hash=unknown
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-X 'github.com/conduktor/ctl/cmd.version=$version' -X 'github.com/conduktor/ctl/cmd.hash=$hash'" -o /conduktor . && rm -rf /app
CMD ["/bin/conduktor"]

FROM scratch
COPY --from=0 /conduktor /bin/conduktor
ENTRYPOINT ["/bin/conduktor"]
