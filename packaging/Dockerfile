FROM alpine:3.11.5

RUN apk add --no-cache ca-certificates

# https://github.com/hairyhenderson/gomplate/releases
ENV GOMPLATE_VERSION v3.6.0

RUN set -eux; \
	apkArch="$(apk --print-arch)"; \
	case "$apkArch" in \
		x86_64)  binArch='amd64'; checksum='acc075c4af7fac980457bf20a9c546f98c9c6ccae515c25bac883e0f98cf7a97b53bd1d5442dcdad6abd4ea4d0d3f5f9cc653b0abb006f0fc77ee8413e2ea180' ;; \
		armv7)   binArch='arm'; checksum='d1143816415d50f492eb45f696b269425a7b2e3185b08719ca8154bbe8f7eed4a0bb2069cfe240a59cc9f97e069a2a99935586a1e9defcc8fffbe298dce0a62c' ;; \
		aarch64) binArch='arm64'; checksum='052cb36f6cdeb9b9bc88867de2c5ce54d638f4d432d78332025484fe93906f5ae7c714f6e34a299380aaf45864cd9640b8a44e16afa0bcb2babe0a5b193c9959' ;; \
		*) echo >&2 "error: unsupported architecture ($apkArch)"; exit 1 ;;\
	esac; \
	filename=gomplate_linux-${binArch}-slim; \
	wget -O /bin/gomplate "https://github.com/hairyhenderson/gomplate/releases/download/v3.6.0/${filename}"; \
	echo "$checksum  /bin/gomplate" | sha512sum -c; \
	chmod +x /bin/gomplate; \
	gomplate --version

LABEL org.opencontainers.image.version=v3.6.0 \
	org.opencontainers.image.title=gomplate \
	org.opencontainers.image.url=https://gomplate.ca \
	org.opencontainers.image.documentation=https://docs.gomplate.ca \
	org.opencontainers.image.licenses=MIT \
	org.opencontainers.image.source="https://github.com/hairyhenderson/gomplate"

ENTRYPOINT [ "/bin/gomplate" ]
