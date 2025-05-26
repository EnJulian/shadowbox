from setuptools import setup, find_packages

setup(
    name="shadowbox",
    version="0.1.0",
    packages=find_packages(),
    install_requires=[
        "yt-dlp>=2025.5.22",
        "mutagen>=1.47.0",
        "requests>=2.32.3",
        "pillow>=10.0.0",
    ],
    entry_points={
        'console_scripts': [
            'shadowbox=main:main',
        ],
    },
    python_requires='>=3.6',
    author="lsnen",
    description="A Python application that downloads music from YouTube or Bandcamp, adds metadata, and embeds album art",
)