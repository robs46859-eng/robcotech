#!/usr/bin/env python3
"""
Test IFC Parser with the BIMS_Structural.ifc file
"""

import asyncio
import sys
from pathlib import Path

# Add app to path
sys.path.insert(0, str(Path(__file__).parent.parent / "services" / "bim_ingestion"))

from app.parsers.ifc import IFCParser


async def test_ifc_parser():
    """Test the IFC parser with our test file"""
    ifc_file_path = Path(__file__).parent / "BIMS_Structural.ifc"
    
    if not ifc_file_path.exists():
        print(f"❌ IFC file not found: {ifc_file_path}")
        return False
    
    print(f"\n{'='*60}")
    print(f"IFC Parser Test")
    print(f"{'='*60}")
    print(f"Testing file: {ifc_file_path}")
    print(f"File size: {ifc_file_path.stat().st_size} bytes\n")
    
    parser = IFCParser()
    
    try:
        result = await parser.parse(ifc_file_path)
        
        print(f"✓ Parse successful!")
        print(f"\nMetadata:")
        for key, value in result.get('metadata', {}).items():
            print(f"  - {key}: {value}")
        
        print(f"\nElements extracted: {len(result.get('elements', []))}")
        for elem in result.get('elements', [])[:5]:
            print(f"  - [{elem.get('type')}] {elem.get('name', 'Unnamed')} ({elem.get('id', 'no-id')[:8]}...)")
        
        if len(result.get('elements', [])) > 5:
            print(f"  ... and {len(result.get('elements', [])) - 5} more")
        
        print(f"\nIssues detected: {len(result.get('issues', []))}")
        for issue in result.get('issues', []):
            print(f"  - [{issue.get('severity')}] {issue.get('description')}")
        
        print(f"\n{'='*60}")
        print(f"TEST PASSED")
        print(f"{'='*60}\n")
        return True
        
    except Exception as e:
        print(f"\n❌ Parse failed: {e}")
        import traceback
        traceback.print_exc()
        return False


if __name__ == "__main__":
    success = asyncio.run(test_ifc_parser())
    sys.exit(0 if success else 1)
