#!/bin/bash

# --- 1. Configuration ---
WIDTH=1280
HEIGHT=720
FRAME_COUNT=10
FOLDER_NAME="captures_$(date +%Y%m%d_%H%M%S)"
REAL_USER=${SUDO_USER:-$(whoami)}

if [ "$(id -u)" -ne 0 ]; then
    echo "Please run as root."
    exit 1
fi

# --- 2. Discovery ---
SENSOR_PATH=$(cam -l 2>/dev/null | grep -o "/base/soc[^ ]*" | head -n 1 | tr -d '()')

if [ -z "$SENSOR_PATH" ]; then
    echo "Error: Camera sensor not found."
    exit 1
fi

mkdir -p "$FOLDER_NAME"

# --- 3. Execute Capture (Universal Approach) ---
echo "Capturing $FRAME_COUNT frames from $SENSOR_PATH..."

# We use the 'buffers' property on the sink if available, 
# or we use a stream-terminator.
timeout 5 gst-launch-1.0 -v libcamerasrc camera-name="$SENSOR_PATH" ! \
    video/x-raw,width=$WIDTH,height=$HEIGHT ! \
    videoconvert ! \
    jpegenc ! \
    multifilesink location="$FOLDER_NAME/frame_%06d.jpg"

# --- 4. Permission Fix ---
echo "Adjusting permissions for user: $REAL_USER"
chown -R "$REAL_USER":"$REAL_USER" "$FOLDER_NAME"
chmod 755 "$FOLDER_NAME"
chmod 644 "$FOLDER_NAME"/*.jpg 2>/dev/null

# --- 5. Verify ---
FILE_COUNT=$(ls -1 "$FOLDER_NAME" | wc -l)
if [ "$FILE_COUNT" -gt 0 ]; then
    echo -e "\n\033[1;32mSUCCESS: $FILE_COUNT frames saved in $FOLDER_NAME\033[0m"
else
    echo -e "\n\033[1;31mFAILED: No images captured.\033[0m"
    echo "Try: libcamera-still -t 5000 -n -o test.jpg (to verify hardware)"
fi