#!/usr/bin/env python3

import argparse
import inspect
import json
import os
import re
import subprocess
import sys
import yt_dlp
from os import environ
from platform import uname
from time import sleep

def is_wsl() -> bool:
    return 'microsoft-standard' in uname().release

def media_name(URL):
    with yt_dlp.YoutubeDL({}) as ydl:
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

def run(cmd):
    print(f"Executing {cmd}")
    with subprocess.Popen(cmd, stderr=subprocess.STDOUT, stdout=subprocess.PIPE, shell=True) as p:
        while p.poll() == None:
            print(p.stdout.read1().decode('utf-8'), end="")
            sleep(0.1)

def set_mp3_dest():
    global mp3_dest
    clip_jam = '/media/mslinn/Clip Jam/Music/playlist'
    if os.path.isdir(clip_jam):
        mp3_dest = clip_jam
    elif environ.get('mp3s') is not None:
        mp3_dest = environ.get('mp3s')
        if not os.path.isdir(mp3_dest):
            sys.exit(f"mp3s environment variable points to {mp3_dest}, but that directory does not exist.")
    elif os.path.isdir('/data/media/mp3s'):
        mp3_dest = '/data/media/mp3s'
    elif is_wsl():
        mp3_dest = f"{win_home()}/Music"
        if not os.path.isdir(mp3_dest):
            sys.exit('Unable to find directory for mp3s.')
    else:
        mp3_dest = os.path.expanduser('~') + "/Music/mp3s"
        if not os.path.isdir(mp3_dest):
            sys.exit('Unable to find directory for mp3s.')

def set_xdest_vdest():
    global xdest, vdest
    xdest="$HOME/Videos"
    vdest="$xdest"
    if is_wsl():
        vdest = f"{win_home()}/Videos"
        xdest = f"{wsl_subdir('storage')}/a_z/videos"
    elif os.path.isdir('/data'):
        vdest = '/data/media/staging'
        xdest = '/data/a_z/videos'
    else:
        sys.exit('Unable to find staging and videos directory')

def win_home() -> str:
    path = os.path.expandvars('$PATH').split(':')
    for index, item in enumerate(path):
        if '/AppData/' in item:
            return item.split('/AppData/')[0]
    sys.exit('Unable to determine Windows home directory.')

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
        vdir = vdest
        if args.video_dest is not None:
            vdir = args.video_dest
        ydl_opts = {
            'format': 'mp4',
            'outtmpl': f"{vdir}/{name}.mp4"
        }
    else:
        ydl_opts = {
            # 'format': 'mp3/bestaudio/best',
            'outtmpl': f"{mp3_dest}/{name}",
            'postprocessors': [{  # Extract audio using ffmpeg
                'key': 'FFmpegExtractAudio',
                'preferredcodec': format,
            }]
        }
    print(f"Saving {ydl_opts['outtmpl']}")
    with yt_dlp.YoutubeDL(ydl_opts) as ydl:
        # print(json.dumps(ydl.sanitize_info(info)))
        error_code = ydl.download(args.url)

    if action == 'mp3':
        mp3_name = name.replace('.webm', '.mp3')
        mp3_name = mp3_name.replace('.mp4', '.mp3')

        # run(f"mp3tag {mp3_name}")
        # s3_name = f"s3://musicmslinn/{basename(mp3_name)})"
        # print(f"Uploading to {s3_name}")
        # run(f"aws s3 cp {mp3_name} {s3_name}")

        print("Copying to gojira...")
        run(f"scp {ydl_opts['outtmpl']['default']} mslinn@gojira:/data/media/mp3s/")
    elif action == 'video' and uname().node != 'gojira':
        print("Copying to gojira...")
        run(f"scp {ydl_opts['outtmpl']['default']} mslinn@gojira:/data/a_z/videos/")
    else:
        sys.abort(f"Invalid action '{action}'")

set_mp3_dest()
set_xdest_vdest()
args = parse_args()
doit(args)
