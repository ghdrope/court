FROM debian:trixie-backports

ARG COMPONENT
# Must match GitHub repository name
ARG PROJECT_NAME="court"
ARG VERSION
ENV VERSION=${VERSION}

# Avoid interactive prompts during apt installs
ENV DEBIAN_FRONTEND=noninteractive

# Install certificates CA
RUN apt-get update \
    && apt-get install -y ca-certificates \
    && update-ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Create directory for binary
RUN mkdir -p /go/bin
ENV PATH="/go/bin:${PATH}"

# ---- COPY pre-built binary (CI/CD build job) ----
COPY .bin/${PROJECT_NAME}-${COMPONENT} /go/bin/${COMPONENT}

# ---- Execution permissions ----
RUN chmod +x /go/bin/${COMPONENT}
