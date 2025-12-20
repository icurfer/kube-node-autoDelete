# ---------- build stage ----------
FROM golang:1.22 AS builder

WORKDIR /workspace

# go.mod 먼저 복사 (캐시 최적화)
COPY go.mod go.sum ./
RUN go mod download

# 소스 복사
COPY cmd/ cmd/
COPY controllers/ controllers/

# 바이너리 빌드
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -a -o manager ./cmd

# ---------- runtime stage ----------
FROM gcr.io/distroless/static:nonroot

WORKDIR /
COPY --from=builder /workspace/manager /manager

USER nonroot:nonroot
ENTRYPOINT ["/manager"]
