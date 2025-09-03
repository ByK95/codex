#!/bin/bash
set -e

# --- Config ---
HEADER="./codex.h"
DLL="./codex.dll"
PROJECT_ROOT="E:/Unreal Projects/OrbitSurvivors"
DEST="$PROJECT_ROOT/Source/OrbitSurvivors"
BINARIES="$PROJECT_ROOT/Binaries/Win64"

# --- Run generator ---
echo "Running generator..."
go run gen_wrapper.go

# --- Copy DLL ---
echo "Copying DLL to $BINARIES"
cp "$DLL" "$BINARIES/"

echo "Done!"
