#!/usr/bin/env python3
"""
Utility script to enable or disable verbose logging in Shadowbox.
"""

import sys
from meta_ops.settings import get_verbose_logging, set_verbose_logging

def main():
    """Main function to handle verbose logging settings."""
    if len(sys.argv) > 1:
        action = sys.argv[1].lower()
        if action in ['on', 'enable', 'true', '1']:
            if set_verbose_logging(True):
                print("✅ Verbose logging enabled")
                print("   You will now see detailed lyrics information during downloads")
            else:
                print("❌ Failed to enable verbose logging")
        elif action in ['off', 'disable', 'false', '0']:
            if set_verbose_logging(False):
                print("✅ Verbose logging disabled")
            else:
                print("❌ Failed to disable verbose logging")
        else:
            print("❌ Invalid option. Use 'on' or 'off'")
            sys.exit(1)
    else:
        # Show current status
        current_status = get_verbose_logging()
        print(f"Current verbose logging status: {'ON' if current_status else 'OFF'}")
        print("\nUsage:")
        print("  python3 enable_verbose.py on   # Enable verbose logging")
        print("  python3 enable_verbose.py off  # Disable verbose logging")

if __name__ == "__main__":
    main()