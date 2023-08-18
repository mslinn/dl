#!/usr/bin/env python3

import inspect
import os
import re
import shutil
import textwrap
import yaml
import yt_dlp
import sys
from media_file import *
from remote import *
from util import *
from argparse import ArgumentParser, HelpFormatter
from os import environ
from platform import uname

class RawFormatter(HelpFormatter):
    def _fill_text(self, text, width, indent):
        return "\n".join([textwrap.fill(line, width) for line in textwrap.indent(textwrap.dedent(text), indent).splitlines()])

def media_name(URL):
    with yt_dlp.YoutubeDL({'quiet': True}) as ydl:
        info = ydl.extract_info(URL, download=False)

    name = re.sub(r'[^A-Za-z0-9 ]+', '', info['title']).strip().replace(' ', '_')
    name = re.sub('__+', '_', name)
    return name

def not_disabled(x):
    name = x[0]
    props = x[1]
    return not 'disabled' in props or props['disabled']==False

def active_remotes():
    return map(lambda x: x[0], filter(not_disabled, config['remotes'].items()))

def parse_args():
    global action, dl_options, format
    aremotes = ", ".join(list(map(lambda x: x, active_remotes())))
    description = inspect.cleandoc(f"""
        Downloads media.
        Defaults to just downloading an MP3, even when the original is a video, unless the -x, -v or -V options are provided.

        Modify {config_file} to suit; at present,
        MP3s are downloaded to {config['local']['mp3s']},
        videos to {config['local']['vdest']}, and
        x-rated videos to {config['local']['xdest']}.
        Active remotes are: {aremotes}.
    """)
    parser = ArgumentParser(prog='dl',
                            description=os.path.expandvars(description),
                            epilog=f"Last modified 2023-07-16.",
                            formatter_class=RawFormatter)
    parser.add_argument('url')
    parser.add_argument('-d', '--debug', action='store_true', help="Enable debug mode")
    parser.add_argument('-v', '--keep-video', action='store_true', help=os.path.expandvars(f"Download video to {config['local']['vdest']} and remotes"))
    parser.add_argument('-x', '--xrated', action='store_true', help=os.path.expandvars(f"Download video to {config['local']['xdest']} and remotes"))
    parser.add_argument('-V', '--video_dest', help=f"download video to the specified directory")
    args = parser.parse_args()

    action = 'mp3'
    format = 'mp3'
    if args.keep_video or args.video_dest or args.xrated:
        action = 'video'
        format = 'mp4'
    return args

def read_config():
    global config, config_file
    config_file = os.path.expanduser("~/dl.config")
    if not os.path.isfile(config_file):
        os.exit(f"Error: {config_file} does not exist.")

    if util.is_wsl():
        os.environ['win_home'] = util.win_home()

    with open(config_file, mode="rb") as file:
        config = yaml.safe_load(file)

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
    args = parse_args()
    media_file = media_file.MediaFile(config = config['local'], path='')
    doit(args)
