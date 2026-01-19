#!/bin/bash

# Configuration
MODEL_URL="https://alphacephei.com/vosk/models/vosk-model-small-ru-0.22.zip"
MODEL_NAME="vosk-model-small-ru-0.22"
MODELS_DIR="models"

echo "🐶 Bobik Model Downloader"

# Create models directory if it doesn't exist
mkdir -p $MODELS_DIR

if [ -d "$MODELS_DIR/$MODEL_NAME" ]; then
    echo "✓ Model $MODEL_NAME already exists in $MODELS_DIR."
    exit 0
fi

echo "Downloading Russian model from $MODEL_URL..."
curl -L $MODEL_URL -o $MODELS_DIR/model.zip

echo "Extracting..."
unzip $MODELS_DIR/model.zip -d $MODELS_DIR/
rm $MODELS_DIR/model.zip

echo "✓ Done! Model is ready at $MODELS_DIR/$MODEL_NAME"
