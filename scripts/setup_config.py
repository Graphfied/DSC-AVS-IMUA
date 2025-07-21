#!/usr/bin/env python3
import os
import yaml
from pathlib import Path

def convert_to_absolute_paths(config_path):
    # Get the absolute path of the project root
    project_root = Path(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
    
    # Read the config file
    with open(config_path, 'r') as f:
        config = yaml.safe_load(f)
    
    # List of keys that should contain file paths
    path_keys = [
        'avs_ecdsa_private_key_store_path',
        'operator_ecdsa_private_key_store_path',
        'bls_private_key_store_path'
    ]
    
    # Convert relative paths to absolute paths
    for key in path_keys:
        if key in config and config[key]:
            # Convert relative path to absolute path
            relative_path = config[key]
            absolute_path = str(project_root / relative_path)
            config[key] = absolute_path
    
    # Write back to the config file
    with open(config_path, 'w') as f:
        yaml.dump(config, f, default_flow_style=False)
    
    print("Configuration updated with absolute paths:")
    for key in path_keys:
        if key in config:
            print(f"{key}: {config[key]}")

if __name__ == "__main__":
    config_path = os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))), "config.yaml")
    convert_to_absolute_paths(config_path) 