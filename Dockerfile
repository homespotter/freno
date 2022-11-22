FROM golang:1.16-alpine

COPY . /go/src/github.com/github/freno
WORKDIR /go/src/github.com/github/freno

RUN go build -ldflags '-w -s' -o freno cmd/freno/main.go

FROM alpine

COPY --from=0 /go/src/github.com/github/freno/freno /usr/local/bin/freno

RUN adduser --system freno
RUN mkdir -p /var/lib/freno && chown -R freno /var/lib/freno

ENV BREAK_CACHE 202006031453

RUN apk add --no-cache aws-cli

ARG AWS_ACCESS_KEY_ID
ARG AWS_SECRET_ACCESS_KEY

RUN mkdir /root/.aws/ && \
    echo "[default]" > /root/.aws/credentials && \
    echo "aws_access_key_id = ${AWS_ACCESS_KEY_ID}" >> /root/.aws/credentials && \
    echo "aws_secret_access_key = ${AWS_SECRET_ACCESS_KEY}" >> /root/.aws/credentials && \
    aws s3 cp s3://homespotter-vault/freno/prod/ /  --recursive && \
    rm -rf /root/.aws

EXPOSE 3000

USER freno
ENTRYPOINT ["freno"]
CMD ["--http", "--verbose"]
