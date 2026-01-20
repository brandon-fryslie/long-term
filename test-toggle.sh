#!/bin/bash
# Test script to verify toggle functionality
# This script repeatedly shows the terminal height
# Press Ctrl+\ three times quickly to toggle between fake and real size

while true; do
    clear
    echo "Terminal height: $(tput lines)"
    echo ""
    echo "Press Ctrl+\\ three times quickly (within 500ms) to toggle"
    echo "Press Ctrl+C to exit"
    sleep 1
done
