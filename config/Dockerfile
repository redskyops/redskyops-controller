FROM alpine:latest

ENV HELM_VERSION="v3.4.1" \
    HELM_SHA256="538f85b4b73ac6160b30fd0ab4b510441aa3fa326593466e8bf7084a9c288420"

ENV KUBECTL_VERSION="v1.16.15" \
    KUBECTL_SHA256="e8913069293156ddf55f243814a22d2384fc18b165efb6200606fdeaad146605"

ENV KUSTOMIZE_VERSION="v3.8.7" \
    KUSTOMIZE_SHA256="4a3372d7bfdffe2eaf729e77f88bc94ce37dc84de55616bfe90aac089bf6fd02"

ENV KONJURE_VERSION="v0.2.1" \
    KONJURE_SHA256="8bf2a82b389076d80a9bd5f379c330e5d74353ef8fac95f851dd26c26349b61c"

ENV HELM_URL="https://get.helm.sh/helm-${HELM_VERSION}-linux-amd64.tar.gz" \
    KUBECTL_URL="https://storage.googleapis.com/kubernetes-release/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl" \
    KUSTOMIZE_URL="https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2F${KUSTOMIZE_VERSION}/kustomize_${KUSTOMIZE_VERSION}_linux_amd64.tar.gz" \
    KONJURE_URL="https://github.com/carbonrelay/konjure/releases/download/${KONJURE_VERSION}/konjure-linux-amd64.tar.gz"

RUN apk --no-cache add curl && \
    curl -L "$HELM_URL" | tar xz -C /usr/local/bin --exclude '*/*[^helm]' --strip-components=1 && \
    curl -L "$KUBECTL_URL" -o /usr/local/bin/kubectl && chmod +x /usr/local/bin/kubectl && \
    curl -L "$KUSTOMIZE_URL" | tar xz -C /usr/local/bin && \
    curl -L "$KONJURE_URL" | tar xz -C /usr/local/bin && \
    mkdir -p /workspace/base && \
    addgroup -g 1000 -S setup && \
    adduser -u 1000 -S setup -G setup

COPY . /workspace/
RUN chown -R setup /workspace

USER setup:setup
RUN konjure kustomize init

WORKDIR "/workspace/base"
ENTRYPOINT ["/workspace/docker-entrypoint.sh"]
CMD ["build"]
