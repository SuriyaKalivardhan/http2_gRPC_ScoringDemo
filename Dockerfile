FROM golang
RUN mkdir /app
RUN git clone https://github.com/SuriyaKalivardhan/http2_gRPC_ScoringDemo.git /app/grpcserver
WORKDIR /app/grpcserver/server
RUN go build server.go
EXPOSE 5001
ENTRYPOINT ["/app/grpcserver/server/server"]
