#!/usr/bin/env python3

import argparse
import inspect
import json
import os
import re
import sys
import yt_dlp
from os import environ
from platform import uname

def help(msg):
    if len(msg.strip())>0:
        print(msg + "\n")
    print(inspect.cleandoc(f"""
        {sys.argv[0]} - Download videos from various sites, including YouTube.
        Can save video, or can just save audio as mp3 to $MP3_DEST on the local machine.
        Also copies to gojira.

        Usage: $( basename $0) [options] url

        Options are:
            -h       display this message and quit
            -v       download video to $VDEST
            -V dir   download video to any given directory
            -x       download video to $XDEST

        Last modified 2023-07-15
        """))
    exit(1)

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
    parser = argparse.ArgumentParser(prog='dl', description='Downloads media')
    parser.add_argument('url')
    parser.add_argument('-d', '--debug', action='store_true')
    parser.add_argument('-v', '--keep-video', action='store_true', help=f"Download video to {VDEST}")
    parser.add_argument('-x', '--xrated', action='store_true')
    parser.add_argument('-V', '--video_dest')
    args = parser.parse_args()

    action = 'mp3'
    format = 'mp3'
    if args.keep_video or args.video_dest or args.xrated:
        action = 'video'
        dl_options = '--keep-video'
        format = 'mp4'
    if args.xrated and 'pornhub.com' in args.url:
        dl_options += ' -ss 3'
    return args

def run(cmd):
    result = run(cmd, capture_output=True, shell=True)
    errors = result.stderr.strip()
    if len(errors)>0:
        print(errors)
    return result.stdout.strip()

def set_MP3_DEST():
    global MP3_DEST
    CLIP_JAM = '/media/mslinn/Clip Jam/Music/playlist'
    if os.path.isdir(CLIP_JAM):
        MP3_DEST = CLIP_JAM
    elif environ.get('mp3s') is not None:
        MP3_DEST = environ.get('mp3s')
        if not os.path.isdir(MP3_DEST):
            os.abort(f"mp3s environment variable points to {MP3_DEST}, but that directory does not exist.")
    elif os.path.isdir('/data/media/mp3s'):
        MP3_DEST = '/data/media/mp3s'
    elif is_wsl():
        MP3_DEST = f"{win_home()}/Music"
        if not os.path.isdir(MP3_DEST):
            os.abort('Unable to find directory for mp3s.')
    else:
        MP3_DEST = os.path.expanduser('~') + "/Music/mp3s"
        if not os.path.isdir(MP3_DEST):
            os.abort('Unable to find directory for mp3s.')

def set_XDEST_VDEST():
    global XDEST, VDEST
    XDEST="$HOME/Videos"
    VDEST="$XDEST"
    if is_wsl():
        VDEST = f"{win_home()}/Videos"
        XDEST = f"{wsl_subdir('storage')}/a_z/videos"
    elif os.path.isdir('/data'):
        VDEST = '/data/media/staging'
        XDEST = '/data/a_z/videos'
    else:
        os.abort('Unable to find staging and videos directory')

def win_home() -> str:
    path = os.path.expandvars('$PATH').split(':')
    for index, item in enumerate(path):
        if '/AppData/' in item:
            return item.split('/AppData/')[0]
    os.abort('Unable to determine Windows home directory.')

def wsl_subdir(subdir):
    for dir in enumerate(['c', 'e', 'f']):
        dir_fq = f"/mnt/{dir}/{subdir}"
        if os.path.isdir(dir_fq):
            return dir_fq
    os.abort(f"Unable to find '{subdir}' within /mnt/")

def doit():
    name = media_name(args.url)
    print(f"Saving {name}.{format}")
    ydl_opts = {
        'format': f'{format}/bestaudio/best',
        # See help(yt_dlp.postprocessor) for a list of available Postprocessors and their arguments
        'postprocessors': [{  # Extract audio using ffmpeg
            'key': 'FFmpegExtractAudio',
            'preferredcodec': format,
        }],
        'outtmpl': name
    }

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
        run(f"scp {mp3_name} mslinn@gojira:/data/media/mp3s/")
    elif action == 'video' and uname().node != 'gojira':
        print("Copying to gojira...")
        run(f"scp {name} mslinn@gojira:/data/a_z/videos/")
    else:
        sys.abort(f"Invalid action '{action}'")

args = parse_args()
doit()
