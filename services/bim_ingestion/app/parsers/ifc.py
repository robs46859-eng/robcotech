"""
IFC File Parser

Parses IFC files using IfcOpenShell and extracts:
- Building elements (walls, doors, windows, etc.)
- Spatial structure (buildings, stories, spaces)
- Properties and quantities
- Materials and classifications
- Potential issues and warnings
"""

import logging
from pathlib import Path
from typing import Any, Dict, List, Optional
from uuid import uuid4

logger = logging.getLogger(__name__)


class IFCParser:
    """Parser for IFC files using IfcOpenShell"""
    
    def __init__(self):
        self.ifc_file = None
    
    async def parse(self, file_path: Path) -> Dict[str, Any]:
        """
        Parse an IFC file and extract normalized data

        Args:
            file_path: Path to the IFC file

        Returns:
            Dictionary containing elements, issues, and metadata
        """
        try:
            import ifcopenshell
            import ifcopenshell.util.element
            # Note: ifcopenshell.util.project may not be available in all versions
        except ImportError:
            logger.warning("IfcOpenShell not available - using mock parser")
            return await self._mock_parse(file_path)
        
        result = {
            "file_path": str(file_path),
            "elements": [],
            "issues": [],
            "metadata": {},
        }
        
        try:
            # Open IFC file
            self.ifc_file = ifcopenshell.open(file_path)
            
            # Extract project metadata
            result["metadata"] = self._extract_metadata()
            
            # Extract spatial structure
            building_elements = self._extract_elements()
            result["elements"] = building_elements
            
            # Detect issues
            issues = self._detect_issues()
            result["issues"] = issues
            
            logger.info(
                f"Parsed IFC: {len(building_elements)} elements, {len(issues)} issues"
            )
            
        except Exception as e:
            logger.error(f"Error parsing IFC file {file_path}: {e}")
            raise
        
        return result
    
    def _extract_metadata(self) -> Dict[str, Any]:
        """Extract project metadata from IFC file"""
        metadata = {}
        
        if not self.ifc_file:
            return metadata
        
        try:
            # Get project information
            projects = self.ifc_file.by_type("IfcProject")
            if projects:
                project = projects[0]
                metadata["name"] = getattr(project, "Name", "Unknown")
                metadata["description"] = getattr(project, "Description", None)
                metadata["phase"] = getattr(project, "Phase", None)
                
            # Get application information
            applications = self.ifc_file.by_type("IfcApplication")
            if applications:
                app = applications[0]
                metadata["author"] = getattr(app, "ApplicationFullName", None)
                metadata["organization"] = getattr(app, "ApplicationDeveloper", None)
                metadata["timestamp"] = getattr(app, "Version", None)
                
        except Exception as e:
            logger.warning(f"Error extracting metadata: {e}")
        
        return metadata
    
    def _extract_elements(self) -> List[Dict[str, Any]]:
        """Extract building elements from IFC file"""
        elements = []
        
        if not self.ifc_file:
            return elements
        
        # Element types to extract
        element_types = [
            "IfcWall", "IfcSlab", "IfcBeam", "IfcColumn",
            "IfcDoor", "IfcWindow", "IfcStair", "IfcRailing",
            "IfcSpace", "IfcBuildingStorey", "IfcBuilding",
            "IfcFurnishingElement", "IfcDistributionElement",
        ]
        
        try:
            for element_type in element_types:
                ifc_elements = self.ifc_file.by_type(element_type)
                
                for elem in ifc_elements:
                    element_data = {
                        "id": getattr(elem, "GlobalId", str(uuid4())),
                        "type": element_type,
                        "name": getattr(elem, "Name", None),
                        "description": getattr(elem, "Description", None),
                        "tag": getattr(elem, "Tag", None),
                        "properties": self._extract_properties(elem),
                        "quantities": self._extract_quantities(elem),
                        "materials": self._extract_materials(elem),
                        "spatial_container": self._get_spatial_container(elem),
                    }
                    elements.append(element_data)
                    
        except Exception as e:
            logger.error(f"Error extracting elements: {e}")
        
        return elements
    
    def _extract_properties(self, element) -> Dict[str, Any]:
        """Extract properties from an element"""
        properties = {}
        
        try:
            # Get property sets
            definitions = element.IsDefinedBy
            for definition in definitions or []:
                if hasattr(definition, "RelatingPropertyDefinition"):
                    prop_def = definition.RelatingPropertyDefinition
                    if hasattr(prop_def, "HasProperties"):
                        for prop in prop_def.HasProperties:
                            if hasattr(prop, "Name") and hasattr(prop, "NominalValue"):
                                properties[prop.Name] = prop.NominalValue.Value
                                
        except Exception as e:
            logger.debug(f"Error extracting properties: {e}")
        
        return properties
    
    def _extract_quantities(self, element) -> Dict[str, Any]:
        """Extract quantities from an element"""
        quantities = {}
        
        try:
            # Get quantity sets
            definitions = element.IsDefinedBy
            for definition in definitions or []:
                if hasattr(definition, "RelatingPropertyDefinition"):
                    prop_def = definition.RelatingPropertyDefinition
                    if hasattr(prop_def, "Quantities"):
                        for qty in prop_def.Quantities:
                            if hasattr(qty, "Name"):
                                if hasattr(qty, "LengthValue"):
                                    quantities[qty.Name] = qty.LengthValue
                                elif hasattr(qty, "AreaValue"):
                                    quantities[qty.Name] = qty.AreaValue
                                elif hasattr(qty, "VolumeValue"):
                                    quantities[qty.Name] = qty.VolumeValue
                                elif hasattr(qty, "CountValue"):
                                    quantities[qty.Name] = qty.CountValue
                                    
        except Exception as e:
            logger.debug(f"Error extracting quantities: {e}")
        
        return quantities
    
    def _extract_materials(self, element) -> List[str]:
        """Extract materials from an element"""
        materials = []
        
        try:
            # Get material association
            if hasattr(element, "HasAssociations"):
                for assoc in element.HasAssociations:
                    if assoc.is_type("IfcRelAssociatesMaterial"):
                        material = assoc.RelatingMaterial
                        if hasattr(material, "Name"):
                            materials.append(material.Name)
                        elif hasattr(material, "Materials"):
                            for mat in material.Materials:
                                if hasattr(mat, "Name"):
                                    materials.append(mat.Name)
                                    
        except Exception as e:
            logger.debug(f"Error extracting materials: {e}")
        
        return materials
    
    def _get_spatial_container(self, element) -> Optional[str]:
        """Get the spatial container of an element"""
        try:
            if hasattr(element, "ContainedInStructure"):
                container = element.ContainedInStructure[0].RelatingStructure
                return getattr(container, "GlobalId", None)
        except Exception:
            pass
        return None
    
    def _detect_issues(self) -> List[Dict[str, Any]]:
        """Detect potential issues in the IFC model"""
        issues = []
        
        if not self.ifc_file:
            return issues
        
        # Check for elements without names
        unnamed_count = 0
        for elem_type in ["IfcWall", "IfcDoor", "IfcWindow"]:
            for elem in self.ifc_file.by_type(elem_type):
                if not getattr(elem, "Name", None):
                    unnamed_count += 1
        
        if unnamed_count > 0:
            issues.append({
                "id": str(uuid4()),
                "type": "naming",
                "severity": "low",
                "description": f"{unnamed_count} elements without names",
            })
        
        # Check for duplicate GUIDs
        all_elements = self.ifc_file.by_type("IfcElement")
        guids = [getattr(e, "GlobalId", None) for e in all_elements]
        duplicates = len(guids) - len(set(guids))
        
        if duplicates > 0:
            issues.append({
                "id": str(uuid4()),
                "type": "duplicate_guid",
                "severity": "high",
                "description": f"{duplicates} duplicate GUIDs found",
            })
        
        return issues
    
    async def _mock_parse(self, file_path: Path) -> Dict[str, Any]:
        """Mock parser for when IfcOpenShell is not available"""
        logger.info(f"Using mock parser for {file_path}")
        return {
            "file_path": str(file_path),
            "elements": [
                {
                    "id": str(uuid4()),
                    "type": "IfcWall",
                    "name": "Mock Wall",
                    "description": "Mock element for testing",
                    "properties": {},
                    "quantities": {},
                    "materials": [],
                    "spatial_container": None,
                }
            ],
            "issues": [],
            "metadata": {
                "name": file_path.stem,
            },
        }
