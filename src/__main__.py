#!/usr/bin/env python3

import re
from dl.argument_parse import ArgParse
from dl.dl_config import DLConfig
from dl.media_file import MediaFile
from dl.remote import Remote
from dl.util import run, samba_mount, samba_parse
from os import environ
from os.path import expandvars
from shutil import copyfile
import sys
from yt_dlp import YoutubeDL

class DL:
    def __init__(self, args) -> None:
        self.args = args
        self.name = self.media_name()

    def media_name(self) -> str:
        with YoutubeDL({'quiet': True}) as ydl:
            info = ydl.extract_info(self.args.args.url, download=False)

        name = re.sub(r'[^A-Za-z0-9 ]+', '', info['title']).strip().replace(' ', '_')
        name = re.sub('__+', '_', name)
        return name

    def set_ydl_opts(self) -> None:
        if args.action == 'video':
            self.vdest = expandvars(config.local['vdest'])
            if args.video_dest is not None:
                self.vdest = args.video_dest
            saved_filename = f"{self.vdest}/{self.name}"
            self.ydl_opts = {
                'format': 'mp4',
                'outtmpl': f"{saved_filename}.mp4"
            }
        else:
            saved_filename = expandvars(f"{config.local['mp3s']}/{self.name}")
            self.ydl_opts = {
                'outtmpl': saved_filename,
                'postprocessors': [{  # Extract audio using ffmpeg
                    'key': 'FFmpegExtractAudio',
                    'preferredcodec': format,
                }]
            }
        print(expandvars(f"Saving {saved_filename}.{format}"))
        self.ydl_opts['quiet'] = True

    def doit(self) -> None:
        self.set_ydl_opts()
        # See help(yt_dlp.postprocessor) for a list of available Postprocessors and their arguments
        with YoutubeDL(self.ydl_opts) as ydl:
            # print(json.dumps(ydl.sanitize_info(info)))
            error_code = ydl.download(args.args.url)

        remotes = config['remotes']
        if args.action == 'mp3':
            mp3_name = self.name.replace('.webm', '.mp3')
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
                    samba_root = samba_mount(remote_name, remote_drive, args.debug)
                    target = f"{samba_root}{local_path}/{mp3_name}.{format}"
                    print(f"Copying to {target}")
                    copyfile(source, target)
                else:
                    print(f"Copying to {target}/{mp3_name}.{format}")
                    run(f"{method} {source} {target}", silent=not args.debug)
        elif args.action == 'video':
            for remote_name in list(remotes.keys()):
                remote = remotes[remote_name]
                if 'disabled' in remote and remote['disabled']:
                    continue

                method = remote['method'] if 'method' in remote and remote['method'] else 'scp'
                dest = remote['vdest']
                if args.xrated: dest = remote['xdest']
                print(f"Copying {saved_filename}.{format} to {remote_name}:{dest}/{self.name}.{format}")
                run(f"{method} {saved_filename}.{format} {remote_name}:{dest}", silent=not args.debug)
        else:
            sys.exit(f"Invalid action '{args.action}'")

if __name__ == '__main__':
    config = DLConfig()
    args = ArgParse(config)
    # media_file = MediaFile(config = config.local, path='')
    dl = DL(args)
    dl.doit(args)
