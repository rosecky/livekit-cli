# Copyright 2023 LiveKit, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

FROM golang:1.23-alpine as builder

ARG TARGETPLATFORM
ARG TARGETARCH
RUN echo building for "$TARGETPLATFORM"

WORKDIR /workspace

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# Copy the server-sdk-go directory to resolve the replace directive
COPY ./server-sdk-go /workspace/server-sdk-go

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY cmd/ cmd/
COPY pkg/ pkg/
COPY version.go version.go

RUN CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH go build -a -o lk ./cmd/lk

FROM alpine:3.21

COPY --from=builder /workspace/lk /lk

# Expose port 9090 for metrics or other services
EXPOSE 9090

# Run the binary.
ENTRYPOINT ["/lk"]
