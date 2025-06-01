# -*- mode: python ; coding: utf-8 -*-

import sys
import os
from PyInstaller.utils.hooks import collect_data_files, collect_submodules

# Collect essential data files only
datas = []
# Only collect essential data files, not all submodules
try:
    datas += collect_data_files('certifi')
except:
    pass

# Add project-specific data files
datas += [
    ('meta_ops', 'meta_ops'),
    ('core', 'core'),
]

# Essential hidden imports only
hiddenimports = []
# Don't collect all submodules - too resource intensive
# Only include essential ones
hiddenimports += [
    # Core dependencies - only essential ones
    'yt_dlp',
    'yt_dlp.extractor',
    'mutagen',
    'mutagen.mp3',
    'mutagen.mp4',
    'mutagen.flac',
    'requests',
    'PIL',
    'PIL.Image',
    'spotipy',
    'spotipy.oauth2',
    
    # Essential system modules
    'urllib3',
    'certifi',
    'charset_normalizer',
    'json',
    'logging',
    'subprocess',
    'argparse',
    'dis',  # Needed by PyInstaller runtime hooks
    'inspect',  # Needed by PyInstaller runtime hooks
    'locale',  # Needed by subprocess
    'multiprocessing',
    'multiprocessing.util',
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
    excludes=[
        # Exclude only clearly unnecessary GUI and development modules
        'tkinter',
        'matplotlib',
        'numpy',
        'scipy',
        'pandas',
        'jupyter',
        'IPython',
        'notebook',
        'pytest',
        'turtle',
        'turtledemo',
        'curses',
        'readline',
        'rlcompleter',
        'pdb',
        'profile',
        'pstats',
        'cProfile',
    ],
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
    upx=False,
    upx_exclude=[],
    runtime_tmpdir=None,
    console=True,
    disable_windowed_traceback=False,
    argv_emulation=False,
    target_arch=None,
    codesign_identity=None,
    entitlements_file=None,
)