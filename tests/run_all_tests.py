#!/usr/bin/env python3
"""
Test runner script to execute all tests in the tests directory.
"""

import os
import sys
import subprocess

def run_test_file(test_file):
    """Run a single test file and return the result."""
    print(f"\n{'='*60}")
    print(f"Running {test_file}")
    print('='*60)
    
    try:
        result = subprocess.run([sys.executable, test_file], 
                              capture_output=False, 
                              text=True)
        return result.returncode == 0
    except Exception as e:
        print(f"❌ Error running {test_file}: {e}")
        return False

def main():
    """Run all test files in the tests directory."""
    print("Shadowbox Test Suite Runner")
    print("="*60)
    
    # Get the directory where this script is located
    tests_dir = os.path.dirname(os.path.abspath(__file__))
    
    # Find all test files
    test_files = []
    for file in os.listdir(tests_dir):
        if file.startswith('test_') and file.endswith('.py'):
            test_files.append(os.path.join(tests_dir, file))
    
    # Also include the main test.py file
    main_test = os.path.join(tests_dir, 'test.py')
    if os.path.exists(main_test):
        test_files.append(main_test)
    
    if not test_files:
        print("❌ No test files found!")
        return False
    
    # Sort test files for consistent execution order
    test_files.sort()
    
    print(f"Found {len(test_files)} test files:")
    for test_file in test_files:
        print(f"  - {os.path.basename(test_file)}")
    
    # Run all tests
    results = []
    for test_file in test_files:
        success = run_test_file(test_file)
        results.append((os.path.basename(test_file), success))
    
    # Summary
    print(f"\n{'='*60}")
    print("TEST SUMMARY")
    print('='*60)
    
    passed = 0
    for test_name, success in results:
        status = "✅ PASS" if success else "❌ FAIL"
        print(f"{status} {test_name}")
        if success:
            passed += 1
    
    total = len(results)
    print(f"\nResults: {passed}/{total} tests passed")
    
    if passed == total:
        print("🎉 All tests passed!")
        return True
    else:
        print("⚠️  Some tests failed")
        return False

if __name__ == "__main__":
    success = main()
    sys.exit(0 if success else 1)