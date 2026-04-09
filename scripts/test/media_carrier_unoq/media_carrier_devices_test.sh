#!/bin/bash

# --- Color Definitions ---
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# --- Test Functions ---

test_mi2s() {
    echo -e "\n[1] Testing Audio Playback (MI2S hw:0,3)..."
    {
        amixer -c0 cset iface=MIXER,name='SEC_MI2S_RX Audio Mixer MultiMedia4' 1
        aplay -D hw:0,3 /usr/share/sounds/alsa/Front_Center.wav
        amixer -c0 cset iface=MIXER,name='SEC_MI2S_RX Audio Mixer MultiMedia4' 0
    } > /dev/null 2>&1
    [ $? -eq 0 ] && echo -e "${GREEN}MI2S SUCCESS${NC}" || echo -e "${RED}MI2S FAILED${NC}"
}

test_standard() {
    echo -e "\n[2] Testing Audio Playback (Standard hw:0,1)..."
    {
        amixer -c0 cset iface=MIXER,name='RX_CODEC_DMA_RX_0 Audio Mixer MultiMedia2' 1
        amixer -c0 cset iface=MIXER,name='RX_MACRO RX0 MUX' 1
        amixer -c0 cset iface=MIXER,name='RX INT0_1 MIX1 INP0' 'RX0'
        amixer -c0 cset iface=MIXER,name='RX INT0 DEM MUX' 1
        amixer -c0 cset iface=MIXER,name='EAR_RDAC Switch' 1
        amixer -c0 cset iface=MIXER,name='HPHL Switch' 1
        amixer -c0 cset iface=MIXER,name='RX_RX0 Digital Volume' 80
        aplay -D hw:0,1 /usr/share/sounds/alsa/Front_Center.wav
        amixer -c0 cset iface=MIXER,name='RX_CODEC_DMA_RX_0 Audio Mixer MultiMedia2' 0
        amixer -c0 cset iface=MIXER,name='RX_MACRO RX0 MUX' 'ZERO'
        amixer -c0 cset iface=MIXER,name='RX INT0_1 MIX1 INP0' 'ZERO'
        amixer -c0 cset iface=MIXER,name='RX INT0 DEM MUX' 'NORMAL_DSM_OUT'
        amixer -c0 cset iface=MIXER,name='EAR_RDAC Switch' 0
        amixer -c0 cset iface=MIXER,name='HPHL Switch' 0
    } > /dev/null 2>&1
    [ $? -eq 0 ] && echo -e "${GREEN}STANDARD SUCCESS${NC}" || echo -e "${RED}STANDARD FAILED${NC}"
}

test_headphone() {
    echo -e "\n[3] Testing Audio Playback (Headphone plughw:0,0)..."
    {
        amixer -c0 cset iface=MIXER,name='RX_CODEC_DMA_RX_0 Audio Mixer MultiMedia1' 1
        amixer -c0 cset iface=MIXER,name='RX_MACRO RX0 MUX' 'AIF1_PB'
        amixer -c0 cset iface=MIXER,name='RX_MACRO RX1 MUX' 'AIF1_PB'
        amixer -c0 cset iface=MIXER,name='RX INT0_1 MIX1 INP0' 'RX0'
        amixer -c0 cset iface=MIXER,name='RX INT1_1 MIX1 INP0' 'RX1'
        amixer -c0 cset iface=MIXER,name='RX INT0 DEM MUX' 'CLSH_DSM_OUT'
        amixer -c0 cset iface=MIXER,name='RX INT1 DEM MUX' 'CLSH_DSM_OUT'
        amixer -c0 cset iface=MIXER,name='RX_COMP1 Switch' 1
        amixer -c0 cset iface=MIXER,name='RX_COMP2 Switch' 1
        amixer -c0 cset iface=MIXER,name='HPHL_RDAC Switch' 1
        amixer -c0 cset iface=MIXER,name='HPHR_RDAC Switch' 1
        amixer -c0 cset iface=MIXER,name='HPHL_COMP Switch' 1
        amixer -c0 cset iface=MIXER,name='HPHR_COMP Switch' 1
        amixer -c0 cset iface=MIXER,name='HPHR Switch' 1
        amixer -c0 cset iface=MIXER,name='HPHL Switch' 1
        amixer -c0 cset iface=MIXER,name='RX_RX0 Digital Volume' 80
        amixer -c0 cset iface=MIXER,name='RX_RX1 Digital Volume' 80
        aplay -D plughw:0,0 /usr/share/sounds/alsa/Front_Center.wav
        amixer -c0 cset iface=MIXER,name='RX_CODEC_DMA_RX_0 Audio Mixer MultiMedia1' 0
        amixer -c0 cset iface=MIXER,name='HPHR Switch' 0
        amixer -c0 cset iface=MIXER,name='HPHL Switch' 0
    } > /dev/null 2>&1
    [ $? -eq 0 ] && echo -e "${GREEN}HEADPHONE SUCCESS${NC}" || echo -e "${RED}HEADPHONE FAILED${NC}"
}

test_recording() {
    echo -e "\n[4] Testing Audio Recording (hw:0,2)..."
    {
        amixer -c0 cset iface=MIXER,name='MultiMedia3 Mixer TX_CODEC_DMA_TX_3' 1
        amixer -c0 cset iface=MIXER,name='TX DEC0 MUX' 'SWR_MIC'
        amixer -c0 cset iface=MIXER,name='TX SMIC MUX0' 'SWR_MIC1'
        amixer -c0 cset iface=MIXER,name='TX_AIF1_CAP Mixer DEC0' 1
        amixer -c0 cset iface=MIXER,name='ADC2 Switch' 1
        amixer -c0 cset iface=MIXER,name='ADC2 Volume' 7
        amixer -c0 cset iface=MIXER,name='ADC2_MIXER Switch' 1
        amixer -c0 cset iface=MIXER,name='ADC2 MUX' 'INP2'
        amixer -c0 cset iface=MIXER,name='TX_DEC0 Volume' 80
        arecord -D hw:0,2 -f S16_LE -c 1 -r 48000 -d 5 record.wav
        amixer -c0 cset iface=MIXER,name='ADC2_MIXER Switch' 0
        amixer -c0 cset iface=MIXER,name='MultiMedia3 Mixer TX_CODEC_DMA_TX_3' 0
    } > /dev/null 2>&1
    if [ $? -eq 0 ] && [ -f record.wav ]; then
        echo -e "${GREEN}RECORDING SUCCESS (record.wav created)${NC}"
        echo "Playing back recorded file..."
        aplay -D hw:0,1 record.wav > /dev/null 2>&1
    else
        echo -e "${RED}RECORDING FAILED${NC}"
    fi
}

test_camera_capture() {
    echo -e "\n[5] Executing Camera Capture Test..."
    
    WIDTH=1280
    HEIGHT=720
    FRAME_COUNT=10
    FOLDER_NAME="captures_$(date +%Y%m%d_%H%M%S)"
    REAL_USER=${SUDO_USER:-$(whoami)}

    SENSOR_PATH=$(cam -l 2>/dev/null | grep -o "/base/soc[^ ]*" | head -n 1 | tr -d '()')

    if [ -z "$SENSOR_PATH" ]; then
        echo -e "${RED}Error: Camera sensor not found.${NC}"
        return 1
    fi

    mkdir -p "$FOLDER_NAME"
    echo "Capturing $FRAME_COUNT frames from $SENSOR_PATH..."

    timeout 5 gst-launch-1.0 -v libcamerasrc camera-name="$SENSOR_PATH" ! \
        video/x-raw,width=$WIDTH,height=$HEIGHT ! \
        videoconvert ! \
        jpegenc ! \
        multifilesink location="$FOLDER_NAME/frame_%06d.jpg" > /dev/null 2>&1

    echo "Adjusting permissions for user: $REAL_USER"
    chown -R "$REAL_USER":"$REAL_USER" "$FOLDER_NAME"
    chmod 755 "$FOLDER_NAME"
    chmod 644 "$FOLDER_NAME"/*.jpg 2>/dev/null

    FILE_COUNT=$(ls -1 "$FOLDER_NAME" 2>/dev/null | wc -l)
    if [ "$FILE_COUNT" -gt 0 ]; then
        echo -e "${GREEN}SUCCESS: $FILE_COUNT frames saved in $FOLDER_NAME${NC}"
    else
        echo -e "${RED}FAILED: No images captured.${NC}"
        echo "Try: libcamera-still -t 5000 -n -o test.jpg"
    fi
}

config_camera_hardware() {
    echo -e "\n[6] Configuring camera hardware type..."
    sudo arduino-linux-config carrier config media-carrier camera0=type1-4lane > /dev/null 2>&1
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}CAMERA CONFIG SUCCESS${NC}"
        echo "A reboot is required to apply hardware changes."
    else
        echo -e "${RED}CAMERA CONFIG FAILED${NC}"
    fi
}

# --- Main Menu Loop ---

# Check for root
if [ "$(id -u)" -ne 0 ]; then
    echo "Please run this script as root (sudo ./mc_test.sh)."
    exit 1
fi

while true; do
    echo -e "\n==============================================="
    echo "       HARDWARE VALIDATION TOOL MENU          "
    echo "==============================================="
    echo "1) Run MI2S Playback Test (hw:0,3)"
    echo "2) Run Standard Playback Test (hw:0,1)"
    echo "3) Run Headphone Playback Test (plughw:0,0)"
    echo "4) Run Audio Recording & Playback Test (5s)"
    echo "5) Run Camera Capture Test (GStreamer)"
    echo "6) Configure Camera Hardware (Type1-4lane)"
    echo "7) RUN ALL TESTS"
    echo "q) Exit"
    echo "==============================================="
    printf "Select an option: "
    read choice

    case "$choice" in
        1) test_mi2s ;;
        2) test_standard ;;
        3) test_headphone ;;
        4) test_recording ;;
        5) test_camera_capture ;;
        6) config_camera_hardware ;;
        7) 
            test_mi2s
            test_standard
            test_headphone
            test_recording
            test_camera_capture
            config_camera_hardware
            ;;
        q|Q) 
            echo "Exiting..."
            break
            ;;
        *) 
            echo -e "${RED}Invalid option, please try again.${NC}" 
            ;;
    esac
done