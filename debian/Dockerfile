# Used by the --platform of the dtc-addon stage
ARG ARCH

FROM golang:1.27.2 AS go

ARG BINARY_NAME
RUN test -n "${BINARY_NAME}" || (echo "Error: BINARY_NAME is not set" && exit 1)
ARG VERSION
RUN test -n "${VERSION}" || (echo "Error: VERSION is not set" && exit 1)
ARG ARCH
RUN test -n "${ARCH}" || (echo "Error: ARCH is not set" && exit 1)
ARG RELEASE

# Download dependencies first (this layer will be cached by docker for quicker image builds).
# This will be recomputed only when these two files change, instead of every build.
COPY go.mod go.sum* /app/
WORKDIR /app

RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
  go mod download

ADD . .

RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
  --mount=type=cache,target=/root/.cache/go-build,sharing=locked \
  export VERSION=$(echo "${VERSION}" | sed -e "s/^v\(.*\)/\1/") \
  export LDFLAGS=$([ -n "$RELEASE" ] && echo "-s -w" || echo "") \
  && GOOS=linux GOARCH=${ARCH} go build -ldflags "${LDFLAGS} -X 'main.Version=${VERSION}'" -o ${BINARY_NAME} ./cmd/${BINARY_NAME}

FROM --platform=linux/${ARCH} debian:bookworm AS dtc-addon

# fdtaddon and dtc from the device-tree addon RFC, with libfdt linked statically
RUN apt-get update && apt-get install -y git meson ninja-build gcc flex bison pkg-config
RUN git clone -b bootlin/support-for-addon-rfc-v2 https://github.com/bootlin/dtc.git /dtc && \
  git -C /dtc checkout c68038e0ff4cde5de37a21419df8a082032ef994
RUN meson setup /dtc/build /dtc -Ddefault_library=static -Dwerror=false -Dyaml=disabled -Dpython=disabled -Dtests=false && \
  meson compile -C /dtc/build dtc fdtaddon

FROM dtc-addon AS devicetree

# The base dtb and the camera addons, from the device-tree sources of the kernel
ARG KERNEL_REPO
RUN test -n "${KERNEL_REPO}" || (echo "Error: KERNEL_REPO is not set" && exit 1)
ARG KERNEL_REF
RUN test -n "${KERNEL_REF}" || (echo "Error: KERNEL_REF is not set" && exit 1)
WORKDIR /kernel
RUN git init -q . && git remote add origin ${KERNEL_REPO} && \
  git sparse-checkout set --no-cone /arch/arm64/boot/dts/qcom/ /include/dt-bindings/ /include/uapi/linux/input-event-codes.h && \
  git fetch -q --depth 1 --filter=blob:none origin ${KERNEL_REF} && git checkout -q FETCH_HEAD
RUN QCOM=arch/arm64/boot/dts/qcom && mkdir /out && \
  cpp -P -nostdinc -undef -D__DTS__ -DDT_ADDON_EXPORTS -x assembler-with-cpp -I ${QCOM} -I include \
    ${QCOM}/monaco-arduino-monza.dts -o /tmp/monza-base.dts && \
  /dtc/build/dtc -q -@ -I dts -O dtb -o /out/monza-base.dtb /tmp/monza-base.dts && \
  for addon in ${QCOM}/arduino-camera-*.dtsa; do \
    /dtc/build/dtc -q -I dts -O dtb -o /out/$(basename ${addon} .dtsa).dtba ${addon} || exit 1; \
  done && \
  /dtc/build/fdtaddon -i /out/monza-base.dtb -t /flat0-connector -o /tmp/check.dtb /out/arduino-camera-imx219-2lanes.dtba

FROM debian:bookworm AS debian

ARG VERSION
RUN test -n "${VERSION}" || (echo "Error: VERSION is not set" && exit 1)
ARG REVISION="1"
ARG ARCH
RUN test -n "${ARCH}" || (echo "Error: ARCH is not set" && exit 1)
ARG DEB_NAME
RUN test -n "${DEB_NAME}" || (echo "Error: DEB_NAME is not set" && exit 1)
ARG BINARY_NAME
RUN test -n "${BINARY_NAME}" || (echo "Error: BINARY_NAME is not set" && exit 1)

RUN apt-get update && apt-get install -y sed

COPY ./debian/${DEB_NAME} /${DEB_NAME}/
COPY --from=go /app/${BINARY_NAME} /${DEB_NAME}/usr/bin/${BINARY_NAME}
COPY --from=dtc-addon /dtc/build/dtc /dtc/build/fdtaddon /${DEB_NAME}/usr/lib/${BINARY_NAME}/addon/
COPY --from=devicetree /out/ /${DEB_NAME}/usr/lib/${BINARY_NAME}/addon/

# Go application are tagged with `v` prefix, this remove the first v if present
RUN export VERSION=$(echo "${VERSION}" | sed -e "s/^v\(.*\)/\1/") && \
  sed -i "s/\$ARCH/${ARCH}/" /${DEB_NAME}/DEBIAN/control && \
  sed -i "s/\$VERSION/${VERSION}/" /${DEB_NAME}/DEBIAN/control && \
  dpkg-deb --build --root-owner-group /${DEB_NAME} &&\
  mv /${DEB_NAME}.deb "/${DEB_NAME}_${VERSION}-${REVISION}_${ARCH}.deb"

FROM scratch

COPY --from=debian /*.deb /
