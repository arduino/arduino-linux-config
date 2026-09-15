FROM debian:trixie
COPY build/arduino-linux-config_*.deb /tmp/arduino-linux-config.deb

RUN apt-get update && apt-get install -y /tmp/arduino-linux-config.deb && rm /tmp/arduino-linux-config.deb

# Create fake device-tree compatible file for board detection
RUN mkdir -p /tmp/compat-root/sys/firmware/devicetree/base && \
    printf 'arduino,imola\0' > /tmp/compat-root/sys/firmware/devicetree/base/compatible

# COMPATIBLE_ROOT_DIR is also used to resolve the boot-id, so it must be provided too
RUN mkdir -p /tmp/compat-root/proc/sys/kernel/random && \
    printf 'test-boot-id\n' > /tmp/compat-root/proc/sys/kernel/random/boot_id

# osrelease holds only the release field of "uname -a", e.g. from
# "Linux uno 7.0.0-g122c2c22d838 #1 SMP PREEMPT Tue May 19 15:08:13 UTC 2026 aarch64 GNU/Linux"
RUN printf '7.0.0-g122c2c22d838\n' > /tmp/compat-root/proc/sys/kernel/osrelease

ENV COMPATIBLE_ROOT_DIR=/tmp/compat-root

# Create fake dtb directory and files
RUN mkdir -p /boot/efi/dtb/qcom

RUN apt-get install -y device-tree-compiler

# Each base dtb gets a distinct node so its md5 is unique and fdtoverlay has a
# valid, non-empty tree to merge overlays into.
RUN set -eu; \
    for pair in "base:qrb2210-arduino-imola-base.dtb" "imola:qrb2210-arduino-imola.dtb"; do \
        node=${pair%%:*}; file=${pair#*:}; \
        printf '/dts-v1/;\n/ {\n  compatible = "arduino,imola";\n  model = "Arduino Imola";\n  %s { status = "okay"; };\n};\n' "$node" > /tmp/root.dts; \
        dtc -I dts -O dtb -o "/boot/efi/dtb/qcom/$file" /tmp/root.dts; \
    done; \
    rm /tmp/root.dts

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
