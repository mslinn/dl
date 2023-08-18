#!/usr/bin/env python3

import argument_parse
import re
import shutil
import yt_dlp
import sys
from config import read_config
from media_file import *
from remote import *
from util import *
from os import environ
from platform import uname

# See https://stackoverflow.com/a/64102901/553865

def media_name(URL):
    with yt_dlp.YoutubeDL({'quiet': True}) as ydl:
        info = ydl.extract_info(URL, download=False)

    name = re.sub(r'[^A-Za-z0-9 ]+', '', info['title']).strip().replace(' ', '_')
    name = re.sub('__+', '_', name)
    return name


def doit(args):
    global vdest
    name = media_name(args.url)

    # See help(yt_dlp.postprocessor) for a list of available Postprocessors and their arguments
    if action == 'video':
        vdest = os.path.expandvars(config['local']['vdest'])
        if args.video_dest is not None:
            vdest = args.video_dest
        saved_filename = f"{vdest}/{name}"
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
                remote_drive, local_path = util.samba_parse(remote_name, remote['mp3s'])
                samba_root = util.samba_mount(remote_name, remote_drive, args.debug)
                target = f"{samba_root}{local_path}/{mp3_name}.{format}"
                print(f"Copying to {target}")
                shutil.copyfile(source, target)
            else:
                print(f"Copying to {target}/{mp3_name}.{format}")
                util.run(f"{method} {source} {target}", silent=not args.debug)
    elif action == 'video':
        for remote_name in list(remotes.keys()):
            remote = remotes[remote_name]
            if 'disabled' in remote and remote['disabled']:
                continue

            method = remote['method'] if 'method' in remote and remote['method'] else 'scp'
            dest = remote['vdest']
            if args.xrated: dest = remote['xdest']
            print(f"Copying {saved_filename}.{format} to {remote_name}:{dest}/{name}.{format}")
            util.run(f"{method} {saved_filename}.{format} {remote_name}:{dest}", silent=not args.debug)
    else:
        sys.abort(f"Invalid action '{action}'")

if __name__ == '__main__':
    read_config()
    args = argument_parse.parse_args()
    media_file = media_file.MediaFile(config = config['local'], path='')
    doit(args)
