#!/usr/bin/env python3
"""
Test script for date extraction functionality.
"""

import os
import sys

# Add the project root to the Python path
project_root = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
sys.path.insert(0, project_root)

def extract_year_from_date(date_str):
    """
    Extract the year from a date string in various formats.
    
    Args:
        date_str (str): The date string to parse
        
    Returns:
        str: The extracted year, or the original string if no year could be extracted
    """
    # Check if it's in the format DATE:YYYYMMDD
    if date_str.startswith('DATE:') and len(date_str) >= 9 and date_str[5:9].isdigit():
        return date_str[5:9]  # Extract year from DATE:YYYYMMDD
    # Check if it's in the format YYYYMMDD
    elif len(date_str) >= 4 and date_str[:4].isdigit():
        return date_str[:4]  # Just take the year
    # Return the original string if no pattern matches
    return date_str

def test_date_extraction():
    """Test the date extraction functionality."""
    print("Testing date extraction functionality...")
    print("=" * 50)
    
    # Test cases
    test_dates = [
        'DATE:20240419',
        '20240419',
        '2024-04-19',
        '2024',
        'Unknown date'
    ]
    
    all_passed = True
    expected_results = ['2024', '2024', '2024', '2024', 'Unknown date']
    
    for i, date in enumerate(test_dates):
        result = extract_year_from_date(date)
        expected = expected_results[i]
        status = "✅ PASS" if result == expected else "❌ FAIL"
        print(f"{status} Original: {date} -> Extracted: {result}")
        if result != expected:
            all_passed = False
            print(f"      Expected: {expected}")
    
    return all_passed

def main():
    """Run all date extraction tests."""
    print("Date Extraction Test Suite")
    print("=" * 60)
    
    success = test_date_extraction()
    
    print("\n" + "=" * 60)
    if success:
        print("🎉 All date extraction tests passed!")
    else:
        print("⚠️  Some date extraction tests failed")
    
    return success

if __name__ == "__main__":
    success = main()
    sys.exit(0 if success else 1)