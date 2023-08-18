import os
import subprocess
import sys
from os import environ
from time import sleep
from platform import uname


def is_wsl() -> bool:
    return 'microsoft-standard' in uname().release

def run(cmd, silent=True):
    if not silent: print(f"Executing {cmd}")
    with subprocess.Popen(cmd, stderr=subprocess.STDOUT, stdout=subprocess.PIPE, shell=True) as p:
        while p.poll() == None:
            stdout = p.stdout.read1()
            if not silent:
                print(stdout.decode('utf-8'), end="")
            sleep(0.1)

# Given a remote node name and a path on that node,
#  - Create the mount point if it does not exist
#  - Mount the remote drive on the mount point if not already mounted
#  - Return the fully qualified path of the mount point
# The remote_path is assumed to start with a leading slash
# Example: samba_mount('camille', 'c')
#  - Creates /mnt/camille/c if it does not already exist
#  - Mounts '\\camille\c' on '/mnt/camille/c'
#  - Returns the mount point, '/mnt/camille/c'
def samba_mount(remote_node, remote_drive, debug):
    slash = '' if remote_drive.startswith('/') else '/'
    mount_point = f"/mnt/{remote_node}{slash}{remote_drive}"
    if not os.path.isdir(mount_point):
        run(f"sudo mkdir {mount_point}", silent=not debug)
    if not os.path.ismount(mount_point):
        run(f"sudo mount -t drvfs '\\\\{remote_node}\{remote_drive}' {mount_point}", silent=False)
    return mount_point

# Parses windows paths of the form C:/path/to/file
# Returns mount point for node_name and local path on the node
# Example: samba_parse('camille', 'c:/blah/ick.poo')
#  returns ['c', '/blah/ick.poo']
def samba_parse(node_name, win_path):
    paths = win_path.split(':')
    if len(paths) != 2:
        sys.abort(f"Invalid windows path '{win_path}'. Check dl.config.")
    return paths

def win_home() -> str:
    if not is_wsl: os.exit("Error: not running in WSL.")

    linux_dir = subprocess.run(['cmd.exe', '/c', "<nul set /p=%UserProfile%"], capture_output=True, text=True).stdout.strip()
    return subprocess.run(['/usr/bin/wslpath', linux_dir], capture_output=True, text=True).stdout.strip()

def wsl_subdir(subdir):
    for i, dir in enumerate(['c', 'e', 'f']):
        dir_fq = f"/mnt/{dir}/{subdir}"
        if os.path.isdir(dir_fq):
            return dir_fq
    sys.exit(f"Unable to find '{subdir}' within /mnt/")
