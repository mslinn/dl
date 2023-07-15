#!/usr/bin/env python3

import inspect
import json
import re
import sys
import yt_dlp

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


format = 'mp3'
if len(sys.argv)==1:
    help("Error: No URL was provided.")
    sys.exit(2)
else:
    URL = sys.argv[1]

with yt_dlp.YoutubeDL({}) as ydl:
    info = ydl.extract_info(URL, download=False)

name = re.sub(r'[^A-Za-z0-9 ]+', '', info['title']).strip().replace(' ', '_')
name = re.sub('__+', '_', name)
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
    error_code = ydl.download(URL)
