FROM golang:1.25-alpine as builder

RUN apk add git curl

WORKDIR /go/src/app

RUN git clone https://github.com/OctopusDeploy/terraform-provider-octopusdeploy.git
RUN cd terraform-provider-octopusdeploy && \
  git checkout mattc/spacebuilder && \
  go get ./... && \
  go build -o terraform-provider-octopusdeploy

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main cmd/web/main.go
RUN mkdir binaries
RUN curl -L -o binaries/opa_linux_amd64 https://github.com/open-policy-agent/opa/releases/download/v1.3.0/opa_linux_amd64
RUN chmod +x binaries/opa_linux_amd64
RUN curl -L -o binaries/tofu https://github.com/OctopusSolutionsEngineering/opentofu/releases/download/v1.10.6.3/tofu
RUN chmod +x binaries/tofu
RUN mkdir -p provider/registry.opentofu.org/octopusdeploy/octopusdeploy/1.18.2/linux_amd64
RUN cp terraform-provider-octopusdeploy/terraform-provider-octopusdeploy functions/provider/registry.opentofu.org/octopusdeploy/octopusdeploy/1.18.2/linux_amd64/terraform-provider-octopusdeploy_v1.18.2

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /go/src/app .
CMD ["/home/nonroot/main"]