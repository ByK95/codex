#!/bin/bash
set -e

# --- Config ---
HEADER="./codex.h"
DLL="./codex.dll"
# PROJECT_ROOT="E:/Unreal Projects/OrbitSurvivors"
PROJECT_NAME="ResourceSystem"
PROJECT_ROOT="E:/Unreal Projects/ResourceSystem"
DEST="$PROJECT_ROOT/Source/$PROJECT_NAME/"
BINARIES="$PROJECT_ROOT/Binaries/Win64"

# --- Run generator ---
echo "Running generator..."
go run gen_wrapper.go -dest "$DEST"

# --- Copy DLL ---
echo "Copying DLL to $BINARIES"
cp "$DLL" "$BINARIES/"

echo "Done!"
