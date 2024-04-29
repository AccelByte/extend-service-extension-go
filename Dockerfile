# gRPC Gen
FROM --platform=$BUILDPLATFORM rvolosatovs/protoc:4.1.0 as grpc-gen
WORKDIR /build
COPY pkg/proto pkg/proto
COPY proto.sh .
RUN mkdir -p apidocs pkg/pb
RUN bash proto.sh


# Extend App Builder			
FROM --platform=$BUILDPLATFORM golang:1.20-alpine as builder
ARG TARGETOS
ARG TARGETARCH
ARG GOOS=$TARGETOS
ARG GOARCH=$TARGETARCH
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=grpc-gen /build/pkg/pb pkg/pb
RUN go build -v -o $TARGETOS/$TARGETARCH/service


# Extend App
FROM alpine:3.17.0
ARG TARGETOS
ARG TARGETARCH
WORKDIR /app
COPY --from=builder /build/$TARGETOS/$TARGETARCH/service service
COPY --from=grpc-gen /build/apidocs apidocs
COPY third_party third_party
# gRPC server port, gRPC gateway port, Prometheus /metrics port
EXPOSE 6565 8000 8080
CMD [ "/app/service" ]