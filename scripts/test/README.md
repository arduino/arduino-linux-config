# Hardware Validation Tool

A comprehensive Bash utility designed to validate audio interfaces and camera functionality on Linux-based systems. This tool is specifically optimized for hardware using ALSA mixers and GStreamer-compatible camera sensors.

## Tests

- **MI2S Audio Test:** Validates the MI2S digital audio path.
- **Stereo Playback:** Tests Standard, Headphone, and Line-Out analog outputs.
- **Loopback Recording:** Captures 5 seconds of audio and immediately plays it back to verify the Microphone/ADC path.
- **Camera Capture:** Automatically detects the sensor path and captures a 10-frame burst at 720p.
- **Auto-Permissions:** Ensures captured media is owned by the logged-in user rather than root.
- **Hardware Configuration:** Integrated tool to set camera carrier modes.

## 📋 Prerequisites

The following packages must be installed on your system:

- `alsa-utils` (`amixer`, `aplay`, `arecord`)
- `gstreamer1.0-tools`
- `libcamera-tools`
- `sudo`
