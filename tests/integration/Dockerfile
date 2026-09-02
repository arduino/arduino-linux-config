FROM debian:trixie
COPY build/arduino-linux-config_*.deb /tmp/arduino-linux-config.deb

RUN apt-get update && apt-get install -y /tmp/arduino-linux-config.deb && rm /tmp/arduino-linux-config.deb

# Create fake device-tree compatible file for board detection
RUN mkdir -p /tmp/compat-root/sys/firmware/devicetree/base && \
    printf 'arduino,imola\0' > /tmp/compat-root/sys/firmware/devicetree/base/compatible
ENV COMPATIBLE_ROOT_DIR=/tmp/compat-root

# Create fake dtb directory and files
RUN mkdir -p /boot/efi/dtb/qcom

RUN apt-get install -y device-tree-compiler

RUN echo '/dts-v1/; / {};' > /tmp/empty.dts && \
    dtc -I dts -O dtb -o /boot/efi/dtb/qcom/qrb2210-arduino-imola-base.dtb /tmp/empty.dts && \
    dtc -I dts -O dtb -o /boot/efi/dtb/qcom/qrb2210-arduino-imola.dtb /tmp/empty.dts && \
    rm /tmp/empty.dts

# Each overlay adds a node named after itself, so that the generated device tree
# differs depending on which overlays get applied.
RUN set -eu; \
    for name in \
        carrier-media \
        video_sound-usbc \
        carrier-media-camera-imx219-csi0-2lanes \
        carrier-media-camera-imx219-csi0-4lanes \
        carrier-media-camera-imx219-csi1-2lanes \
        carrier-media-camera-imx219-csi1-4lanes \
        carrier-media-panel-8in_touch_a-dsi \
    ; do \
        node=$(echo "$name" | tr '_.' '--'); \
        printf '/dts-v1/;\n/plugin/;\n/ {\n  fragment@0 {\n    target-path = "/";\n    __overlay__ {\n      %s { status = "okay"; };\n    };\n  };\n};\n' "$node" > /tmp/overlay.dts; \
        dtc -I dts -O dtb -o "/boot/efi/dtb/qcom/qrb2210-arduino-imola-$name.dtbo" /tmp/overlay.dts; \
    done; \
    rm /tmp/overlay.dts
