import os
import subprocess
import sys
from pathlib import Path
from time import sleep
from typing import List
from platform import uname

def is_wsl() -> bool:
    """Detects if running under wsl
    Returns:
        bool: true if running under wsl
    """
    return 'microsoft-standard' in uname().release

def run(cmd: str, silent:bool=True) -> None:
    """Run command and print output if not silent
    Args:
        cmd (str): command to execute
        silent (bool, optional): suppress output if true (default).
    """

    if not silent: print(f"Executing {cmd}")
    with subprocess.Popen(cmd, stderr=subprocess.STDOUT, stdout=subprocess.PIPE, shell=True) as p:
        while p.poll() == None:
            stdout = p.stdout.read1()
            if not silent:
                print(stdout.decode('utf-8'), end="")
            sleep(0.1)

def samba_mount(remote_node: str, remote_drive: str, debug: bool) -> Path:
    """ Given a remote node name and a path on that node:
        - Create the mount point if it does not exist
        - Mount the remote drive on the mount point if not already mounted
        - Return the absolute Path of the mount point
    Example: samba_mount('camille', 'c')
        - Creates /mnt/camille/c if it does not already exist
        - Mounts '\\camille\c' on '/mnt/camille/c'
        - Returns the mount point, '/mnt/camille/c'
    Args:
        remote_node (str): network node of the remote
        remote_drive (str): directory on the remote, must start with leading slash and must exist
        debug (bool): Enable output if true
    Returns:
        Path: path to local mount point, will be created if necessary
    """

    slash = '' if remote_drive.startswith('/') else '/'
    mount_point = f"/mnt/{remote_node}{slash}{remote_drive}"
    if not os.path.isdir(mount_point):
        run(f"sudo mkdir -p {mount_point}", silent=not debug)
    if not os.path.ismount(mount_point):
        run(f"sudo mount -t drvfs '\\\\{remote_node}\\{remote_drive}' {mount_point}", silent=False)
    return Path(mount_point)

# Parses windows paths of the form C:/path/to/file
# Returns mount point for a path on the node
# Example: samba_parse('c:/blah/ick.poo')
#  returns ['c', '/blah/ick.poo']
def samba_parse(win_path: Path|str) -> List[str]:
    paths = str(win_path).split(':')
    if len(paths) != 2:
        exit(f"Invalid windows path '{win_path}'. Check dl.config.")
    return paths

def win_home() -> Path:
    if not is_wsl: exit("Error: not running in WSL.")

    linux_dir = subprocess.run(['cmd.exe', '/c', "<nul set /p=%UserProfile%"], capture_output=True, text=True).stdout.strip()
    wsl_dir = subprocess.run(['/usr/bin/wslpath', linux_dir], capture_output=True, text=True).stdout.strip()
    return Path(wsl_dir)

# TODO: Delete, not used.
# TODO: The logic is flawed. Completely rewrite
# TODO: change enumerate to scan for mounted windows drives
# @return absolute path Path of subdir on mounted drive
def wsl_subdir(subdir: str) -> Path:
    for i, dir in enumerate(['c', 'e', 'f']):
        dir_fq = Path(f"/mnt/{dir}/{subdir}")
        if dir_fq.is_dir():
            return dir_fq
    sys.exit(f"Unable to find '{subdir}' within /mnt/")
