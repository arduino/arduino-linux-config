# Fixture for the VentunoQ (arduino,monza) board.
FROM ubuntu:24.04
COPY build/arduino-linux-config_*.deb /tmp/arduino-linux-config.deb

RUN apt-get update && apt-get install -y \
    /tmp/arduino-linux-config.deb \
    device-tree-compiler \
    dosfstools \
    && rm /tmp/arduino-linux-config.deb

ENV COMPATIBLE_ROOT_DIR=/tmp/compat-root
ENV KERNEL_VERSION=6.8.0-1-qcom

# Fake root used for board detection. The image is Debian, but the VentunoQ
# code path (and therefore the addons) is only enabled on Ubuntu.
RUN mkdir -p /tmp/compat-root/sys/firmware/devicetree/base \
             /tmp/compat-root/proc/sys/kernel/random \
             /tmp/compat-root/etc \
             /tmp/compat-root/boot/grub \
             "/tmp/compat-root/lib/firmware/${KERNEL_VERSION}/device-tree/qcom" && \
    printf 'arduino,monza\0' > /tmp/compat-root/sys/firmware/devicetree/base/compatible && \
    printf 'test-boot-id\n' > /tmp/compat-root/proc/sys/kernel/random/boot_id && \
    printf 'ID=ubuntu\n' > /tmp/compat-root/etc/os-release && \
    printf 'linux /boot/vmlinuz-%s root=/dev/sda1\n' "${KERNEL_VERSION}" > /tmp/compat-root/boot/grub/grub.cfg

# The base device tree is shipped as a combined dtb: a concatenation of flattened
# device trees, only one of which is compatible with arduino,monza.
RUN set -eu; \
    printf '/dts-v1/;\n/ { compatible = "arduino,other"; };\n' > /tmp/other.dts; \
    printf '/dts-v1/;\n/ { compatible = "arduino,monza"; };\n' > /tmp/monza.dts; \
    dtc -I dts -O dtb -o /tmp/other.dtb /tmp/other.dts; \
    dtc -I dts -O dtb -o /tmp/monza.dtb /tmp/monza.dts; \
    cat /tmp/other.dtb /tmp/monza.dtb > "/tmp/compat-root/lib/firmware/${KERNEL_VERSION}/device-tree/qcom/combined-dtb.dtb"; \
    rm /tmp/other.dts /tmp/monza.dts /tmp/other.dtb /tmp/monza.dtb

# Each addon overlay adds a node named after itself, so that the generated device
# tree differs depending on which addon is enabled.
RUN set -eu; \
    mkdir -p /var/lib/arduino-linux-config/overlays; \
    for name in \
        monaco-monza-automation-hat \
        monaco-addons-iqaudio-codeczero-monza \
    ; do \
        printf '/dts-v1/;\n/plugin/;\n/ {\n  fragment@0 {\n    target-path = "/";\n    __overlay__ {\n      %s { status = "okay"; };\n    };\n  };\n};\n' "$name" > /tmp/overlay.dts; \
        dtc -I dts -O dtb -o "/var/lib/arduino-linux-config/overlays/$name.dtbo" /tmp/overlay.dts; \
    done; \
    rm /tmp/overlay.dts
