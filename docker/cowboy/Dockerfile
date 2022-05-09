FROM golang:1.18.1-alpine
WORKDIR /app
COPY . .

RUN apk add build-base librdkafka-dev
# RUN apk add --no-cache --virtual .make-deps bash make wget git gcc g++ && apk add --no-cache musl-dev zlib-dev openssl zstd-dev pkgconfig libc-dev
# RUN  wget https://github.com/edenhill/librdkafka/archive/v${LIBRD_VER}.tar.gz
# RUN tar -xvf v${LIBRD_VER}.tar.gz && cd librdkafka-${LIBRD_VER} && ./configure --prefix /usr && make && make install && make clean 
# RUN rm -rf librdkafka-${LIBRD_VER} && rm -rf v${LIBRD_VER}.tar.gz && apk del .make-deps

RUN go build -tags dynamic -o cowboy cmd/cowboy/main.go
ENTRYPOINT [ "/app/cowboy" ]
