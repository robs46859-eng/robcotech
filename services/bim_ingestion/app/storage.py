"""BIM Storage Module

Handles storage of BIM files and data
"""

import logging
import shutil
from pathlib import Path
from typing import Optional
from uuid import uuid4

from fastapi import UploadFile

logger = logging.getLogger(__name__)


class BIMStorage:
    """Storage handler for BIM files and data"""
    
    def __init__(self, storage_path: str):
        self.storage_path = Path(storage_path)
        self.uploads_dir = self.storage_path / "uploads"
        self.processed_dir = self.storage_path / "processed"
        self.temp_dir = self.storage_path / "temp"
    
    async def initialize(self):
        """Initialize storage directories"""
        self.uploads_dir.mkdir(parents=True, exist_ok=True)
        self.processed_dir.mkdir(parents=True, exist_ok=True)
        self.temp_dir.mkdir(parents=True, exist_ok=True)
        logger.info(f"BIM storage initialized at {self.storage_path}")
    
    async def close(self):
        """Cleanup storage resources"""
        pass
    
    async def save_upload(self, file: UploadFile) -> Path:
        """
        Save an uploaded file to storage
        
        Args:
            file: The uploaded file
            
        Returns:
            Path to the saved file
        """
        # Generate unique filename
        file_extension = Path(file.filename).suffix if file.filename else ""
        unique_filename = f"{uuid4()}{file_extension}"
        file_path = self.uploads_dir / unique_filename
        
        # Save file
        with open(file_path, "wb") as buffer:
            shutil.copyfileobj(file.file, buffer)
        
        logger.info(f"Saved upload to {file_path}")
        return file_path
    
    async def move_to_processed(self, file_path: Path, project_id: str) -> Path:
        """Move a file from uploads to processed"""
        new_path = self.processed_dir / project_id / file_path.name
        new_path.parent.mkdir(parents=True, exist_ok=True)
        shutil.move(str(file_path), str(new_path))
        return new_path
    
    async def get_file(self, file_id: str) -> Optional[Path]:
        """Get a file by ID"""
        # Search in uploads and processed
        for directory in [self.uploads_dir, self.processed_dir]:
            for file_path in directory.rglob(f"*{file_id}*"):
                if file_path.is_file():
                    return file_path
        return None
    
    async def delete_file(self, file_path: Path) -> bool:
        """Delete a file"""
        try:
            if file_path.exists():
                file_path.unlink()
                logger.info(f"Deleted file {file_path}")
                return True
        except Exception as e:
            logger.error(f"Error deleting file {file_path}: {e}")
        return False
