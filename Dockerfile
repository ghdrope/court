FROM golang:1.26 AS builder

ARG TARGETOS
ARG TARGETARCH

ARG COMPONENT
# Must match GitHub repository name
ARG PROJECT_NAME="court"
ARG BUILD_DATE
ARG GIT_COMMIT
ARG VERSION
ENV VERSION=${VERSION}

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 \
    GOOS=${TARGETOS} \
    GOARCH=${TARGETARCH} \
    go build \
      -ldflags "\
        -s -w \
        -X github.com/ghdrope/go-version.Version=${VERSION} \
        -X github.com/ghdrope/go-version.GitCommit=${GIT_COMMIT} \
        -X github.com/ghdrope/go-version.BuildDate=${BUILD_DATE}" \
      -o /out/${PROJECT_NAME}-${COMPONENT} \
      ./cmd  

      

FROM debian:trixie-backports

ARG VERSION
ENV VERSION=${VERSION}

# Avoid interactive prompts during apt installs
ENV DEBIAN_FRONTEND=noninteractive

# Install certificates CA
RUN apt-get update \
    && apt-get install -y ca-certificates \
    && update-ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /out/${PROJECT_NAME}-${COMPONENT} /usr/local/bin/${PROJECT_NAME}-${COMPONENT}

# ---- Execution permissions ----
RUN chmod +x /usr/local/bin/${PROJECT_NAME}-${COMPONENT}
