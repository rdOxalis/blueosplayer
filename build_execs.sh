#!/bin/bash

# Build script for bluosplayer
# Builds executables for multiple platforms

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check if we're in the right directory
if [ ! -d "src" ] || [ ! -d "release" ]; then
    echo -e "${RED}Error: 'src' and 'release' directories must exist!${NC}"
    echo "Please run this script from the project root directory."
    exit 1
fi

# Clean old builds
echo -e "${YELLOW}Cleaning old builds...${NC}"
rm -f release/bluosplayer-*

# Change to source directory
cd src

# Check if all required source files exist
required_files=("main.go" "common.go" "bluos.go" "sonos.go" "network.go" "localization.go")
for file in "${required_files[@]}"; do
    if [ ! -f "$file" ]; then
        echo -e "${RED}Error: Required source file '$file' not found in src directory!${NC}"
        exit 1
    fi
done

echo -e "${GREEN}Building bluosplayer for multiple platforms...${NC}"
echo ""

# Function to build for a specific platform
build_platform() {
    local goos=$1
    local goarch=$2
    local output=$3
    local extra_info=$4
    
    echo -e "${YELLOW}Building for ${extra_info}...${NC}"
    
    GOOS=$goos GOARCH=$goarch go build -o ../release/$output *.go
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Successfully built: $output${NC}"
        # Make Linux/Mac binaries executable
        if [[ "$goos" == "linux" || "$goos" == "darwin" ]]; then
            chmod +x ../release/$output
        fi
    else
        echo -e "${RED}✗ Failed to build: $output${NC}"
    fi
    echo ""
}

# Build for all platforms
build_platform "linux" "amd64" "bluosplayer-linux-amd64" "Linux amd64"
build_platform "windows" "amd64" "bluosplayer-win-amd64.exe" "Windows amd64"
build_platform "linux" "arm" "bluosplayer-linux-armv6" "Raspberry Pi 1/Zero/Zero W (ARMv6)"
build_platform "linux" "arm" "bluosplayer-linux-armv7" "Raspberry Pi 2/3/4/Zero 2 W (ARMv7)"
build_platform "linux" "arm64" "bluosplayer-linux-arm64" "Raspberry Pi 4/5 (ARM64)"
build_platform "darwin" "arm64" "bluosplayer-apple-arm64" "Apple Silicon (M1/M2/M3)"
build_platform "darwin" "amd64" "bluosplayer-apple-amd64" "Intel Mac"

# Return to project root
cd ..

# Show results
echo -e "${GREEN}Build complete! Executables are in the 'release' directory:${NC}"
ls -la release/bluosplayer-*

# Create version info file
echo -e "\n${YELLOW}Creating version info...${NC}"
echo "Build date: $(date)" > release/BUILD_INFO.txt
echo "Git commit: $(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')" >> release/BUILD_INFO.txt
echo "" >> release/BUILD_INFO.txt
echo "Platforms built:" >> release/BUILD_INFO.txt
ls release/bluosplayer-* | sed 's/release\//  - /' >> release/BUILD_INFO.txt

echo -e "${GREEN}Done!${NC}"