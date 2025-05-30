#!/usr/bin/env python3
"""
Utility script to enhance multiple music files with Spotify metadata.
This script will scan a directory for music files and try to enhance them with metadata from Spotify.
"""

import os
import sys
import argparse
import glob

# Add project root to Python path
script_dir = os.path.dirname(os.path.abspath(__file__))
project_root = os.path.dirname(script_dir)
sys.path.insert(0, project_root)

from meta_ops.spotify_metadata import enhance_existing_file_with_spotify

def main():
    parser = argparse.ArgumentParser(description='Enhance music files with Spotify metadata')
    parser.add_argument('path', help='Path to a music file or directory containing music files')
    parser.add_argument('--recursive', '-r', action='store_true', help='Recursively scan directories')
    parser.add_argument('--extensions', '-e', default='.mp3,.opus', help='Comma-separated list of file extensions to process (default: .mp3,.opus)')
    parser.add_argument('--dry-run', '-d', action='store_true', help='Show what would be done without making changes')
    
    args = parser.parse_args()
    
    # Check if path exists
    if not os.path.exists(args.path):
        print(f"Error: Path does not exist: {args.path}")
        return 1
    
    # Get list of extensions to process
    extensions = args.extensions.split(',')
    
    # Get list of files to process
    files_to_process = []
    
    if os.path.isfile(args.path):
        # Single file
        files_to_process.append(args.path)
    else:
        # Directory
        for ext in extensions:
            if args.recursive:
                # Recursive search
                pattern = os.path.join(args.path, f'**/*{ext}')
                files_to_process.extend(glob.glob(pattern, recursive=True))
            else:
                # Non-recursive search
                pattern = os.path.join(args.path, f'*{ext}')
                files_to_process.extend(glob.glob(pattern))
    
    # Sort files
    files_to_process.sort()
    
    # Process files
    print(f"Found {len(files_to_process)} files to process")
    
    success_count = 0
    fail_count = 0
    
    for file_path in files_to_process:
        print(f"\nProcessing: {file_path}")
        
        if args.dry_run:
            print("  [DRY RUN] Would enhance this file with Spotify metadata")
            success_count += 1
            continue
        
        # Extract title from filename
        filename = os.path.splitext(os.path.basename(file_path))[0]
        
        # Try to enhance the file
        success = enhance_existing_file_with_spotify(file_path)
        
        if success:
            success_count += 1
        else:
            fail_count += 1
    
    # Print summary
    print("\nSummary:")
    print(f"  Total files: {len(files_to_process)}")
    print(f"  Successfully processed: {success_count}")
    print(f"  Failed: {fail_count}")
    
    return 0 if fail_count == 0 else 1

if __name__ == '__main__':
    sys.exit(main())