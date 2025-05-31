# -*- mode: python ; coding: utf-8 -*-

import sys
import os
from PyInstaller.utils.hooks import collect_data_files, collect_submodules

# Collect all data files and submodules
datas = []
datas += collect_data_files('yt_dlp')
datas += collect_data_files('mutagen')
datas += collect_data_files('spotipy')
datas += collect_data_files('requests')
datas += collect_data_files('certifi')

# Add project-specific data files
datas += [
    ('meta_ops', 'meta_ops'),
    ('core', 'core'),
]

# Collect hidden imports
hiddenimports = []
hiddenimports += collect_submodules('yt_dlp')
hiddenimports += collect_submodules('mutagen')
hiddenimports += collect_submodules('spotipy')
hiddenimports += [
    # Core dependencies
    'requests',
    'PIL',
    'PIL.Image',
    'PIL.ImageFile',
    'spotipy',
    'spotipy.oauth2',
    
    # Standard library modules that might need explicit inclusion
    'urllib3',
    'urllib.parse',
    'urllib.request',
    'certifi',
    'charset_normalizer',
    'idna',
    'json',
    'base64',
    'logging',
    'subprocess',
    'time',
    'os',
    'sys',
    're',
    'glob',
    'shutil',
    'argparse',
    
    # Additional dependencies that might be needed
    'http.client',
    'ssl',
    'socket',
    'threading',
    'queue',
    'functools',
    'itertools',
    'collections',
    'datetime',
    'hashlib',
    'hmac',
    'email',
    'email.mime',
    'email.mime.text',
    'email.mime.multipart',
]

block_cipher = None

a = Analysis(
    ['core/main.py'],
    pathex=[],
    binaries=[],
    datas=datas,
    hiddenimports=hiddenimports,
    hookspath=[],
    hooksconfig={},
    runtime_hooks=[],
    excludes=[],
    win_no_prefer_redirects=False,
    win_private_assemblies=False,
    cipher=block_cipher,
    noarchive=False,
)

pyz = PYZ(a.pure, a.zipped_data, cipher=block_cipher)

exe = EXE(
    pyz,
    a.scripts,
    a.binaries,
    a.zipfiles,
    a.datas,
    [],
    name='shadowbox',
    debug=False,
    bootloader_ignore_signals=False,
    strip=False,
    upx=True,
    upx_exclude=[],
    runtime_tmpdir=None,
    console=True,
    disable_windowed_traceback=False,
    argv_emulation=False,
    target_arch=None,
    codesign_identity=None,
    entitlements_file=None,
)