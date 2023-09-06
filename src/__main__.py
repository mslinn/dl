#!/usr/bin/env python3

import re
import sys
from argparse import Namespace
from dl.argument_parse import ArgParse
from dl.dl_config import DLConfig
from dl.media_file import MediaFile
from dl.remote import Remote
from dl.util import run, samba_mount, samba_parse
from os import environ
from os.path import expandvars
from shutil import copyfile
from typing import Any
from yt_dlp import YoutubeDL

class DL:
    def __init__(self, arg_parse: ArgParse) -> None:
        self.args: Namespace = arg_parse.args
        self.config: DLConfig = DLConfig()
        self.name: str = self.media_name()

    def media_name(self) -> str:
        info: (dict[str, Any] | None) = None
        with YoutubeDL({'quiet': True}) as ydl:
            info = ydl.extract_info(self.args.url, download=False)

        if isinstance(info, dict):
            name = re.sub(r'[^A-Za-z0-9 ]+', '', info['title']).strip().replace(' ', '_')
            name = re.sub('__+', '_', name)
            return name

        return "no_name"

    def set_ydl_opts(self) -> str:
        if self.args.action == 'video':
            if self.args.video_dest is not None:
                self.vdest = self.args.video_dest
            elif isinstance(self.config.local, dict) and self.config.local['vdest'] is not None:
                self.vdest = expandvars(self.config.local['vdest'])
            else:
                exit("Error: local video destination is not defined in config and was not specified on the command line")
            saved_filename = f"{self.vdest}/{self.name}"
            self.ydl_opts = {
                'format': 'mp4',
                'outtmpl': f"{saved_filename}.mp4"
            }
            return saved_filename
        else:
            saved_filename = None
            if isinstance(self.config.local, dict) and self.config.local['mp3s'] is not None:
                saved_filename = expandvars(f"{self.config.local['mp3s']}/{self.name}")
                self.ydl_opts = {
                    'outtmpl': saved_filename,
                    'postprocessors': [{  # Extract audio using ffmpeg
                        'key': 'FFmpegExtractAudio',
                        'preferredcodec': format,
                    }]
                }
                print(expandvars(f"Saving {saved_filename}.{format}"))
                self.ydl_opts['quiet'] = True
                return saved_filename
            else:
                exit("Error: local mp3 destination is not defined in config")

    def doit(self) -> None:
        saved_filename = self.set_ydl_opts()
        # See help(yt_dlp.postprocessor) for a list of available Postprocessors and their arguments
        with YoutubeDL(self.ydl_opts) as ydl:
            # print(json.dumps(ydl.sanitize_info(info)))
            error_code = ydl.download(self.args.url)
            if error_code:
                exit(f"Error: ydl.download failed with error code {error_code}")

        remotes: (dict|None) = self.config.remotes
        if self.args.action == 'mp3':
            mp3_name = self.name.replace('.webm', '.mp3')
            mp3_name = mp3_name.replace('.mp4', '.mp3')

            # run(f"mp3tag {mp3_name}")
            if isinstance(remotes, dict):
                for remote_name in list(remotes.keys()):
                    remote = remotes[remote_name]
                    if 'disabled' in remote and remote['disabled']:
                        continue

                    method = remote['method'] if 'method' in remote and remote['method'] else 'scp'
                    mp3s = remote['mp3s']
                    source = f"{saved_filename}.{format}"
                    target = f"{remote_name}:{mp3s}"
                    if method == 'samba':
                        remote_drive, local_path = samba_parse(remote['mp3s'])
                        samba_root = samba_mount(remote_name, remote_drive, self.args.debug)
                        target = f"{samba_root}{local_path}/{mp3_name}.{format}"
                        print(f"Copying to {target}")
                        copyfile(source, target)
                    else:
                        print(f"Copying to {target}/{mp3_name}.{format}")
                        run(f"{method} {source} {target}", silent=not self.args.debug)
        elif self.args.action == 'video':
            if isinstance(remotes, dict):
                for remote_name in list(remotes.keys()):
                    remote = remotes[remote_name]
                    if 'disabled' in remote and remote['disabled']:
                        continue

                    method = remote['method'] if 'method' in remote and remote['method'] else 'scp'
                    dest = remote['vdest']
                    if self.args.xrated: dest = remote['xdest']
                    print(f"Copying {saved_filename}.{format} to {remote_name}:{dest}/{self.name}.{format}")
                    run(f"{method} {saved_filename}.{format} {remote_name}:{dest}", silent=not self.args.debug)
        else:
            sys.exit(f"Invalid action '{self.args.action}'")

if __name__ == '__main__':
    config: DLConfig = DLConfig()
    arg_parse: ArgParse = ArgParse(config)
    # media_file = MediaFile(config = config.local, path='')
    dl: DL = DL(arg_parse)
    dl.doit()
