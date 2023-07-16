#!/usr/bin/env python3

import argparse
import os
import re
import subprocess
import sys
import yaml
import yt_dlp
from os import environ
from platform import uname
from time import sleep

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
    parser = argparse.ArgumentParser(prog='dl', description='Downloads media', epilog="Last modified 2023-07-16")
    parser.add_argument('url')
    parser.add_argument('-d', '--debug', action='store_true', help="Enable debug mode")
    parser.add_argument('-v', '--keep-video', action='store_true', help=f"Download video to {vdest}")
    parser.add_argument('-x', '--xrated', action='store_true', help=f"Download video to {xdest}")
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
    # print(f"Executing {cmd}")
    with subprocess.Popen(cmd, stderr=subprocess.STDOUT, stdout=subprocess.PIPE, shell=True) as p:
        while p.poll() == None:
            stdout = p.stdout.read1()
            if not silent:
                print(stdout.decode('utf-8'), end="")
            sleep(0.1)

def set_mp3_dest():
    global mp3_dest
    if os.path.isdir(config['mp3s']['automount']):
        mp3_dest = config['mp3s']['automount']
    elif environ.get('mp3s') is not None:
        mp3_dest = environ.get('mp3s')
        if not os.path.isdir(mp3_dest):
            sys.exit(f"mp3s environment variable points to {mp3_dest}, but that directory does not exist.")
    elif os.path.isdir(config['local']['mp3s']):
        mp3_dest = config['local']['mp3s']
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
        vdir = config['local']['vdest']
        if args.video_dest is not None:
            vdir = args.video_dest
        ydl_opts = {
            'format': 'mp4',
            'outtmpl': f"{vdir}/{name}.mp4"
        }
    else:
        ydl_opts = {
            'outtmpl': f"{config['local']['mp3s']}/{name}",
            'postprocessors': [{  # Extract audio using ffmpeg
                'key': 'FFmpegExtractAudio',
                'preferredcodec': format,
            }]
        }
    print(f"Saving {ydl_opts['outtmpl']}")
    ydl_opts['quiet'] = True
    with yt_dlp.YoutubeDL(ydl_opts) as ydl:
        # print(json.dumps(ydl.sanitize_info(info)))
        error_code = ydl.download(args.url)

    if action == 'mp3':
        mp3_name = name.replace('.webm', '.mp3')
        mp3_name = mp3_name.replace('.mp4', '.mp3')

        # run(f"mp3tag {mp3_name}")

        print("Copying to gojira...")
        run(f"scp {ydl_opts['outtmpl']['default']}.{format} mslinn@gojira:/data/media/mp3s/")
    elif action == 'video':
        remotes = config['remotes']
        for remote_name in list(remotes.keys()):
            remote = remotes[remote_name]
            if 'disabled' in remote and remote['disabled']:
                continue

            vdest = remote['vdest']
            print(f"Copying {ydl_opts['outtmpl']['default']} to {remote_name}:{vdest}...")
            run(f"scp {ydl_opts['outtmpl']['default']} {remote_name}:{vdest}")
    else:
        sys.abort(f"Invalid action '{action}'")

read_config()
xdest = config['local']['xdest']
vdest = config['local']['vdest']
set_mp3_dest()

args = parse_args()
doit(args)
