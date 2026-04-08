#!/bin/bash

# Initialize status variables
PLAYBACK_1_STATUS="Pending"
PLAYBACK_2_STATUS="Pending"
PLAYBACK_3_STATUS="Pending"
RECORD_STATUS="Pending"
PLAY_RECORDED_STATUS="Pending"
RESET_STATUS="Pending"
CAMERA_STATUS="Pending"

echo "Starting full hardware validation sequence..."
echo "------------------------------------------------------------"

# --- 1. AUDIO PLAYBACK TEST (Standard hw:0,1) ---
echo "[1/7] Testing Audio Playback (Standard hw:0,1)..."
{
    sudo amixer -c0 cset iface=MIXER,name='RX_CODEC_DMA_RX_0 Audio Mixer MultiMedia2' 1
    sudo amixer -c0 cset iface=MIXER,name='RX_MACRO RX0 MUX' 1
    sudo amixer -c0 cset iface=MIXER,name='RX INT0_1 MIX1 INP0' 'RX0'
    sudo amixer -c0 cset iface=MIXER,name='RX INT0 DEM MUX' 1
    sudo amixer -c0 cset iface=MIXER,name='EAR_RDAC Switch' 1
    sudo amixer -c0 cset iface=MIXER,name='HPHL Switch' 1
    sudo amixer -c0 cset iface=MIXER,name='RX_RX0 Digital Volume' 80
    sudo aplay -D hw:0,1 /usr/share/sounds/alsa/Front_Center.wav
} > /dev/null 2>&1
[ $? -eq 0 ] && PLAYBACK_1_STATUS="SUCCESS" || PLAYBACK_1_STATUS="FAILED"

# --- 2. AUDIO PLAYBACK TEST (Headphone plughw:0,0) ---
echo "[2/7] Testing Audio Playback (Headphone plughw:0,0)..."
{
    sudo amixer -c0 cset iface=MIXER,name='RX_CODEC_DMA_RX_0 Audio Mixer MultiMedia1' 1
    sudo amixer -c0 cset iface=MIXER,name='RX_MACRO RX0 MUX' 'AIF1_PB'
    sudo amixer -c0 cset iface=MIXER,name='RX_MACRO RX1 MUX' 'AIF1_PB'
    sudo amixer -c0 cset iface=MIXER,name='RX INT0_1 MIX1 INP0' 'RX0'
    sudo amixer -c0 cset iface=MIXER,name='RX INT1_1 MIX1 INP0' 'RX1'
    sudo amixer -c0 cset iface=MIXER,name='RX INT0 DEM MUX' 'CLSH_DSM_OUT'
    sudo amixer -c0 cset iface=MIXER,name='RX INT1 DEM MUX' 'CLSH_DSM_OUT'
    sudo amixer -c0 cset iface=MIXER,name='RX_COMP1 Switch' 1
    sudo amixer -c0 cset iface=MIXER,name='RX_COMP2 Switch' 1
    sudo amixer -c0 cset iface=MIXER,name='HPHL_RDAC Switch' 1
    sudo amixer -c0 cset iface=MIXER,name='HPHR_RDAC Switch' 1
    sudo amixer -c0 cset iface=MIXER,name='HPHL_COMP Switch' 1
    sudo amixer -c0 cset iface=MIXER,name='HPHR_COMP Switch' 1
    sudo amixer -c0 cset iface=MIXER,name='HPHR Switch' 1
    sudo amixer -c0 cset iface=MIXER,name='HPHL Switch' 1
    sudo amixer -c0 cset iface=MIXER,name='RX_RX0 Digital Volume' 80
    sudo amixer -c0 cset iface=MIXER,name='RX_RX1 Digital Volume' 80
    sudo aplay -D plughw:0,0 /usr/share/sounds/alsa/Front_Center.wav
} > /dev/null 2>&1
[ $? -eq 0 ] && PLAYBACK_2_STATUS="SUCCESS" || PLAYBACK_2_STATUS="FAILED"

# --- 3. AUDIO PLAYBACK TEST (Line-Out plughw:0,1) ---
echo "[3/7] Testing Audio Playback (Line-Out plughw:0,1)..."
{
    sudo amixer -c0 cset iface=MIXER,name='RX_CODEC_DMA_RX_0 Audio Mixer MultiMedia2' 1
    sudo amixer -c0 cset iface=MIXER,name='RX_MACRO RX0 MUX' 1
    sudo amixer -c0 cset iface=MIXER,name='RX INT0_1 MIX1 INP0' 'RX0'
    sudo amixer -c0 cset iface=MIXER,name='RX INT0 DEM MUX' 1
    sudo amixer -c0 cset iface=MIXER,name='LO_RDAC Switch' 1
    sudo amixer -c0 cset iface=MIXER,name='HPHL Switch' 1
    sudo amixer -c0 cset iface=MIXER,name='RX_RX0 Digital Volume' 200
    sudo aplay -D plughw:0,1 /usr/share/sounds/alsa/Front_Center.wav
} > /dev/null 2>&1
[ $? -eq 0 ] && PLAYBACK_3_STATUS="SUCCESS" || PLAYBACK_3_STATUS="FAILED"

# --- 4. AUDIO RECORD TEST (5 Seconds) ---
echo "[4/7] Testing Audio Recording (hw:0,2)..."
{
    sudo amixer -c0 cset iface=MIXER,name='MultiMedia3 Mixer TX_CODEC_DMA_TX_3' 1
    sudo amixer -c0 cset iface=MIXER,name='TX DEC0 MUX' 'SWR_MIC'
    sudo amixer -c0 cset iface=MIXER,name='TX SMIC MUX0' 'SWR_MIC1'
    sudo amixer -c0 cset iface=MIXER,name='TX_AIF1_CAP Mixer DEC0' 1
    sudo amixer -c0 cset iface=MIXER,name='ADC2 Switch' 1
    sudo amixer -c0 cset iface=MIXER,name='ADC2 Volume' 7
    sudo amixer -c0 cset iface=MIXER,name='ADC2_MIXER Switch' 1
    sudo amixer -c0 cset iface=MIXER,name='ADC2 MUX' 'INP2'
    sudo amixer -c0 cset iface=MIXER,name='TX_DEC0 Volume' 80
    sudo arecord -D hw:0,2 -f S16_LE -c 1 -r 48000 -d 5 record.wav
    sudo amixer -c0 cset iface=MIXER,name='ADC2_MIXER Switch' 0
    sudo amixer -c0 cset iface=MIXER,name='MultiMedia3 Mixer TX_CODEC_DMA_TX_3' 0
} > /dev/null 2>&1
[ $? -eq 0 ] && RECORD_STATUS="SUCCESS" || RECORD_STATUS="FAILED"

# --- 5. PLAYBACK OF RECORDED FILE ---
echo "[5/7] Playing back recorded file..."
if [ -f record.wav ]; then
    sudo aplay -D hw:0,1 record.wav > /dev/null 2>&1
    [ $? -eq 0 ] && PLAY_RECORDED_STATUS="SUCCESS" || PLAY_RECORDED_STATUS="FAILED"
else
    PLAY_RECORDED_STATUS="FAILED (File not found)"
fi

# --- 7. CAMERA CONFIG ---
echo "[7/7] Configuring camera..."
sudo arduino-linux-config carrier config media-carrier camera0=type1-4lane > /dev/null 2>&1
[ $? -eq 0 ] && CAMERA_STATUS="SUCCESS" || CAMERA_STATUS="FAILED"

# --- FINAL SUMMARY REPORT ---
echo ""
echo "==============================================="
echo "         HARDWARE TEST SUMMARY REPORT          "
echo "==============================================="
printf "%-25s %s\n" "Playback (Standard):" "$PLAYBACK_1_STATUS"
printf "%-25s %s\n" "Playback (Headphone):" "$PLAYBACK_2_STATUS"
printf "%-25s %s\n" "Playback (Line-Out):" "$PLAYBACK_3_STATUS"
printf "%-25s %s\n" "Recording (5s):" "$RECORD_STATUS"
printf "%-25s %s\n" "Play Recorded File:" "$PLAY_RECORDED_STATUS"
printf "%-25s %s\n" "Camera Config:" "$CAMERA_STATUS"
echo "==============================================="

# Check for camera success and prompt for reboot
if [ "$CAMERA_STATUS" = "SUCCESS" ]; then
    echo ""
    echo ">>> Camera configuration applied successfully."
    printf "A reboot is required. Reboot now? (y/n): "
    read choice
    if [ "$choice" = "y" ] || [ "$choice" = "Y" ]; then
        echo "Rebooting..."
        sudo reboot
    else
        echo "Please remember to reboot later for changes to take effect."
    fi
fi

# Exit with 1 if any failure occurred (POSIX syntax)
if [ "$PLAYBACK_1_STATUS" = "FAILED" ] || [ "$PLAYBACK_2_STATUS" = "FAILED" ] || \
   [ "$PLAYBACK_3_STATUS" = "FAILED" ] || [ "$RECORD_STATUS" = "FAILED" ] || \
   [ "$PLAY_RECORDED_STATUS" = "FAILED" ] || [ "$RESET_STATUS" = "FAILED" ] || \
   [ "$CAMERA_STATUS" = "FAILED" ]; then
    exit 1
fi