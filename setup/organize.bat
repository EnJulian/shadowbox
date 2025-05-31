@echo off
REM Script to organize the project structure and create symbolic links
REM Run this script after cloning the repository to set up the project structure

echo Setting up project structure...

REM Get the project root directory
set "SCRIPT_DIR=%~dp0"
set "PROJECT_ROOT=%SCRIPT_DIR%.."
cd "%PROJECT_ROOT%"

REM Create directories if they don't exist
if not exist core mkdir core
if not exist exec_bin mkdir exec_bin
if not exist meta_ops mkdir meta_ops
if not exist tests mkdir tests

REM Move files to appropriate directories
REM Only move if the files exist and are not already in their target directories

REM Core files
if exist shbox.py (
    if not exist core\shbox.py (
        move shbox.py core\
    )
)
if exist main.py (
    if not exist core\main.py (
        move main.py core\
    )
)

REM Meta operations files
if exist metadata.py (
    if not exist meta_ops\metadata.py (
        move metadata.py meta_ops\
    )
)
if exist cover_art.py (
    if not exist meta_ops\cover_art.py (
        move cover_art.py meta_ops\
    )
)
if exist downloader.py (
    if not exist meta_ops\downloader.py (
        move downloader.py meta_ops\
    )
)

REM Executable scripts
if exist shbox.sh (
    if not exist exec_bin\shbox.sh (
        move shbox.sh exec_bin\
    )
)
if exist shbox.bat (
    if not exist exec_bin\shbox.bat (
        move shbox.bat exec_bin\
    )
)
if exist run.sh (
    if not exist exec_bin\run.sh (
        move run.sh exec_bin\
    )
)
if exist run.bat (
    if not exist exec_bin\run.bat (
        move run.bat exec_bin\
    )
)

REM Test files
if exist test.py (
    if not exist tests\test.py (
        move test.py tests\
    )
)

REM Setup files
if exist requirements.txt (
    if not exist setup\requirements.txt (
        move requirements.txt setup\
    )
)
if exist setup.py (
    if not exist setup\setup.py (
        move setup.py setup\
    )
)
if exist install.sh (
    if not exist setup\install.sh (
        move install.sh setup\
    )
)
if exist install.bat (
    if not exist setup\install.bat (
        move install.bat setup\
    )
)

REM Create symbolic links in the project root
REM Windows requires administrator privileges to create symbolic links
REM We'll use mklink command which requires admin rights
echo Creating symbolic links (may require administrator privileges)...
mklink shbox.sh exec_bin\shbox.sh 2>nul
mklink run.sh exec_bin\run.sh 2>nul
mklink shbox.bat exec_bin\shbox.bat 2>nul
mklink run.bat exec_bin\run.bat 2>nul

REM If mklink fails (no admin rights), copy files instead
if not exist shbox.bat (
    echo Symbolic links failed, copying files instead...
    copy exec_bin\shbox.bat shbox.bat >nul
    copy exec_bin\run.bat run.bat >nul
    copy exec_bin\shbox.sh shbox.sh >nul
    copy exec_bin\run.sh run.sh >nul
)

echo Project structure set up successfully!
echo You can now run the application using:
echo shbox.bat or run.bat

pause