# See here for image contents: https://github.com/microsoft/vscode-dev-containers/tree/v0.245.2/containers/go/.devcontainer/base.Dockerfile

ARG VARIANT="1.21"
ARG LINT="1.53.3"

ARG USERNAME=vscode
ARG USER_UID=1000
ARG USER_GID=$USER_UID

FROM mcr.microsoft.com/vscode/devcontainers/go:${VARIANT}

# Update packages
RUN apt-get update

# [Package] Install golangci-lint 
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.55.2

# [Package] Install act
RUN curl -s https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash

# [Pakage] Install openapi
RUN go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@latest

RUN chmod -R a+rwX /go/pkg

RUN BIN="/usr/local/bin" && \
    VERSION="1.26.1" && \
    curl -sSL \
    "https://github.com/bufbuild/buf/releases/download/v${VERSION}/buf-$(uname -s)-$(uname -m)" \
    -o "${BIN}/buf" && \
    chmod +x "${BIN}/buf"

# [Package] Terraform

# RUN sudo apt-get update && sudo apt-get install -y gnupg software-properties-common
# RUN wget -O- https://apt.releases.hashicorp.com/gpg | \
#     gpg --dearmor | \
#     sudo tee /usr/share/keyrings/hashicorp-archive-keyring.gpg

# RUN gpg --no-default-keyring \
#     --keyring /usr/share/keyrings/hashicorp-archive-keyring.gpg \
#     --fingerprint

# RUN echo "deb [signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] \
#     https://apt.releases.hashicorp.com $(lsb_release -cs) main" | \
#     sudo tee /etc/apt/sources.list.d/hashicorp.list

# RUN sudo apt update
# RUN sudo apt-get install terraform

# [Package] jsonnet 

RUN apt-get install 
RUN sudo apt-get install jsonnet

# [Package] yq
RUN go install github.com/mikefarah/yq/v4@latest

# [Package] yc 
RUN curl -sSL https://storage.yandexcloud.net/yandexcloud-yc/install.sh | bash

# [Package] node 

RUN apt-get install 
RUN sudo apt-get install -y nodejs
RUN sudo apt-get install -y npm

# [Package] migrate 

RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-arm64.tar.gz | tar xvz && \
    mv migrate /usr/local/bin/migrate

# [Package] wire

RUN go install github.com/google/wire/cmd/wire@latest