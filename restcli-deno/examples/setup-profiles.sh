#!/bin/bash

echo "Setting up HTTP TUI profiles..."

# Copy example files if they don't exist
if [ ! -f ../.session.json ]; then
  cp .session.json.example ../.session.json
  echo "✓ Created ../.session.json from example"
else
  echo "⚠ ../.session.json already exists, skipping"
fi

if [ ! -f ../.profiles.json ]; then
  cp .profiles.json.example ../.profiles.json
  echo "✓ Created ../.profiles.json from example"
else
  echo "⚠ ../.profiles.json already exists, skipping"
fi

echo ""
echo "Next steps:"
echo "1. Edit .session.json to add your tokens and variables"
echo "2. Edit .profiles.json to customize your header profiles"
echo "3. Run 'deno task dev' to start the TUI"
echo "4. Press 'p' to cycle through profiles"
echo ""
echo "See PROFILES.md for detailed usage guide"
