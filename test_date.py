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

# Test cases
test_dates = [
    'DATE:20240419',
    '20240419',
    '2024-04-19',
    '2024',
    'Unknown date'
]

for date in test_dates:
    print(f"Original: {date} -> Extracted: {extract_year_from_date(date)}")