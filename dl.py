#!/usr/bin/env python3

import argparse
import os
import re
import shutil
import subprocess
import sys
import textwrap
import yaml
import yt_dlp
from argparse import ArgumentParser, HelpFormatter
from os import environ
from platform import uname
from time import sleep

class RawFormatter(HelpFormatter):
    def _fill_text(self, text, width, indent):
        return "\n".join([textwrap.fill(line, width) for line in textwrap.indent(textwrap.dedent(text), indent).splitlines()])

def is_wsl() -> bool:
    return 'microsoft-standard' in uname().release

def media_name(URL):
    with yt_dlp.YoutubeDL({'quiet': True}) as ydl:
        info = ydl.extract_info(URL, download=False)

    name = re.sub(r'[^A-Za-z0-9 ]+', '', info['title']).strip().replace(' ', '_')
    name = re.sub('__+', '_', name)
    return name

def parse_args():
    global action, dl_options, format
    description = f"Downloads media.\nDefaults to just downloading an MP3, even when the original is a video.\nMP3s are downloaded to {config['local']['mp3s']}."
    parser = ArgumentParser(prog='dl',
                            description=os.path.expandvars(description),
                            epilog=f"Last modified 2023-07-16.",
                            formatter_class=RawFormatter)
    parser.add_argument('url')
    parser.add_argument('-d', '--debug', action='store_true', help="Enable debug mode")
    parser.add_argument('-v', '--keep-video', action='store_true', help=os.path.expandvars(f"Download video to {vdest}"))
    parser.add_argument('-x', '--xrated', action='store_true', help=os.path.expandvars(f"Download video to {xdest}"))
    parser.add_argument('-V', '--video_dest', help=f"download video to the specified directory")
    args = parser.parse_args()

    action = 'mp3'
    format = 'mp3'
    if args.keep_video or args.video_dest or args.xrated:
        action = 'video'
        format = 'mp4'
    return args

def read_config():
    global config
    config_file = os.path.expanduser("~/dl.config")
    if not os.path.isfile(config_file):
        os.exit(f"Error: {config_file} does not exist.")

    if is_wsl():
        os.environ['win_home'] = win_home()

    with open(config_file, mode="rb") as file:
        config = yaml.safe_load(file)

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
def samba_mount(remote_node, remote_drive):
    slash = '' if remote_drive.startswith('/') else '/'
    mount_point = f"/mnt/{remote_node}{slash}{remote_drive}"
    if not os.path.isdir(mount_point):
        run(f"sudo mkdir {mount_point}", silent=not args.debug)
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

def set_mp3_dest():
    global mp3_dest
    if os.path.isdir(config['mp3s']['automount']):
        mp3_dest = config['mp3s']['automount']
    elif environ.get('mp3s') is not None:
        mp3_dest = environ.get('mp3s')
        if not os.path.isdir(mp3_dest):
            sys.exit(f"mp3s environment variable points to {mp3_dest}, but that directory does not exist.")
    elif os.path.expandvars(os.path.isdir(config['local']['mp3s'])):
        mp3_dest = os.path.expandvars(config['local']['mp3s'])
    else:
        mp3_dest = os.path.expanduser('~') + "/Music/mp3s"
        if not os.path.isdir(mp3_dest):
            sys.exit('Unable to find directory for mp3s.')

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

def doit(args):
    name = media_name(args.url)

    # See help(yt_dlp.postprocessor) for a list of available Postprocessors and their arguments
    if action == 'video':
        vdir = os.path.expandvars(config['local']['vdest'])
        if args.video_dest is not None:
            vdir = args.video_dest
        saved_filename = f"{vdir}/{name}"
        ydl_opts = {
            'format': 'mp4',
            'outtmpl': f"{saved_filename}.mp4"
        }
    else:
        saved_filename = os.path.expandvars(f"{config['local']['mp3s']}/{name}")
        ydl_opts = {
            'outtmpl': saved_filename,
            'postprocessors': [{  # Extract audio using ffmpeg
                'key': 'FFmpegExtractAudio',
                'preferredcodec': format,
            }]
        }
    print(os.path.expandvars(f"Saving {saved_filename}.{format}"))
    ydl_opts['quiet'] = True
    with yt_dlp.YoutubeDL(ydl_opts) as ydl:
        # print(json.dumps(ydl.sanitize_info(info)))
        error_code = ydl.download(args.url)

    remotes = config['remotes']
    if action == 'mp3':
        mp3_name = name.replace('.webm', '.mp3')
        mp3_name = mp3_name.replace('.mp4', '.mp3')

        # run(f"mp3tag {mp3_name}")
        for remote_name in list(remotes.keys()):
            remote = remotes[remote_name]
            if 'disabled' in remote and remote['disabled']:
                continue

            method = remote['method'] if 'method' in remote and remote['method'] else 'scp'
            mp3s = remote['mp3s']
            source = f"{saved_filename}.{format}"
            target = f"{remote_name}:{mp3s}"
            if method == 'samba':
                remote_drive, local_path = samba_parse(remote_name, remote['mp3s'])
                samba_root = samba_mount(remote_name, remote_drive)
                target = f"{samba_root}{local_path}/{mp3_name}.{format}"
                print(f"Copying to {target}")
                shutil.copyfile(source, target)
            else:
                print(f"Copying to {target}/{mp3_name}.{format}")
                run(f"{method} {source} {target}", silent=not args.debug)
    elif action == 'video':
        for remote_name in list(remotes.keys()):
            remote = remotes[remote_name]
            if 'disabled' in remote and remote['disabled']:
                continue

            method = remote['method'] if 'method' in remote and remote['method'] else 'scp'
            dest = remote['vdest']
            if args.xrated: dest = remote['xdest']
            print(f"Copying {saved_filename}.{format} to {remote_name}:{dest}/{name}.{format}")
            run(f"{method} {saved_filename}.{format} {remote_name}:{dest}", silent=not args.debug)
    else:
        sys.abort(f"Invalid action '{action}'")

read_config()
xdest = config['local']['xdest']
vdest = config['local']['vdest']
set_mp3_dest()

args = parse_args()
doit(args)
