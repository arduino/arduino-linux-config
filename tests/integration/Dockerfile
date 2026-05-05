FROM debian:bookworm

ARG ARCH=arm64

COPY build/arduino-linux-config_*_${ARCH}.deb /tmp/arduino-linux-config.deb
RUN apt-get update && apt-get install -y /tmp/arduino-linux-config.deb && rm /tmp/arduino-linux-config.deb

# Create fake dtb directory and files
RUN mkdir -p /boot/efi/dtb/qcom

RUN apt-get install -y device-tree-compiler

RUN echo '/dts-v1/; / {};' > /tmp/empty.dts && \
    dtc -I dts -O dtb -o /boot/efi/dtb/qcom/qrb2210-arduino-imola-base.dtb /tmp/empty.dts && \
    dtc -I dts -O dtb -o /boot/efi/dtb/qcom/qrb2210-arduino-imola.dtb /tmp/empty.dts && \
    dtc -I dts -O dtb -o /boot/efi/dtb/qcom/qrb2210-arduino-imola-carrier-media.dtbo /tmp/empty.dts && \
    dtc -I dts -O dtb -o /boot/efi/dtb/qcom/qrb2210-arduino-imola-video_sound-usbc.dtbo /tmp/empty.dts && \
    dtc -I dts -O dtb -o /boot/efi/dtb/qcom/qrb2210-arduino-imola-carrier-media-camera-imx219-csi0-2lanes.dtbo /tmp/empty.dts && \
    dtc -I dts -O dtb -o /boot/efi/dtb/qcom/qrb2210-arduino-imola-carrier-media-camera-imx219-csi0-4lanes.dtbo /tmp/empty.dts && \
    dtc -I dts -O dtb -o /boot/efi/dtb/qcom/qrb2210-arduino-imola-carrier-media-camera-imx219-csi1-2lanes.dtbo /tmp/empty.dts && \
    dtc -I dts -O dtb -o /boot/efi/dtb/qcom/qrb2210-arduino-imola-carrier-media-camera-imx219-csi1-4lanes.dtbo /tmp/empty.dts && \
    dtc -I dts -O dtb -o /boot/efi/dtb/qcom/qrb2210-arduino-imola-carrier-media-panel-8in_touch_a-dsi.dtbo /tmp/empty.dts && \
    rm /tmp/empty.dts
