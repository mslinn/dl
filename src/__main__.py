#!/usr/bin/env python3

import os
import re
import sys
try:
    from colorama import init as colorama_init
    from colorama import Fore, Back, Style
    from dl.argument_parse import ArgParse
    from dl.dl_config import DLConfig
    from dl.util import run, samba_mount, samba_parse
    from os.path import expandvars
    from shutil import copyfile
    from typing import Any
    from yt_dlp import YoutubeDL
except ModuleNotFoundError as e:
    print(f"Error: {e} - Is the right Python virtual environment active?")
    exit(1)

class DL:
    def __init__(self, arg_parse: ArgParse) -> None:
        self.arg_parse: ArgParse = arg_parse
        self.config: DLConfig = DLConfig()
        self.name: str = self.media_name()
        colorama_init()

    def media_name(self) -> str:
        info: (dict[str, Any] | None) = None
        with YoutubeDL({'quiet': True}) as ydl:
            info = ydl.extract_info(self.arg_parse.args.url, download=False)

        if isinstance(info, dict):
            name = re.sub(r'[^A-Za-z0-9 ]+', '', info['title']).strip().replace(' ', '_')
            name = re.sub('__+', '_', name)
            return name

        return "no_name"

    def set_ydl_opts(self) -> str:
        if self.arg_parse.action == 'video':
            if self.arg_parse.args.video_dest is not None:
                self.vdest = self.arg_parse.args.video_dest
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
                        'preferredcodec': self.arg_parse.format,
                    }]
                }
                print(expandvars(f"Saving {saved_filename}.{self.arg_parse.format}"))
                self.ydl_opts['quiet'] = True
                return saved_filename
            else:
                exit("Error: local mp3 destination is not defined in config")

    def doit(self) -> None:
        saved_filename = self.set_ydl_opts()
        # See help(yt_dlp.postprocessor) for a list of available Postprocessors and their arguments
        with YoutubeDL(self.ydl_opts) as ydl:
            url = self.arg_parse.args.url
            error_code = ydl.download([url])
            if error_code:
                exit(f"Error: ydl.download failed with error code {error_code}")

        remotes: (dict|None) = self.config.remotes
        if self.arg_parse.action == 'mp3':
            mp3_name = self.name.replace('.webm', '.mp3')
            mp3_name = mp3_name.replace('.mp4', '.mp3')

            # run(f"mp3tag {mp3_name}")
            if isinstance(remotes, dict):
                for remote_name in list(remotes.keys()):
                    try:
                        remote = remotes[remote_name]
                        if 'disabled' in remote and remote['disabled']:
                            continue

                        method = remote['method'] if 'method' in remote and remote['method'] else 'scp'
                        mp3s = remote['mp3s']
                        source = f"{saved_filename}.{self.arg_parse.format}"
                        target = f"{remote_name}:{mp3s}"
                        if method == 'samba':
                            remote_drive, local_path = samba_parse(remote['mp3s'])
                            samba_root = samba_mount(remote_name, remote_drive, self.arg_parse.args.debug)
                            target = f"{samba_root}{local_path}/{mp3_name}.{self.arg_parse.format}"
                            print(f"Copying to {target}")
                            copyfile(source, target)
                        else:
                            print(f"Copying to {target}/{mp3_name}.{self.arg_parse.format}")
                            run(f"{method} {source} {target}", silent=not self.arg_parse.args.debug)
                    except Exception as exception:
                        print(Fore.YELLOW + str(type(exception)) + f": while copying {saved_filename}.{self.arg_parse.format} to {dest}{Style.RESET_ALL}")
        elif self.arg_parse.action == 'video':
            if os.path.isfile(f"{saved_filename}.webm"):
                os.remove(f"{saved_filename}.webm")
            if isinstance(remotes, dict):
                for remote_name in list(remotes.keys()):
                    dest = f"{saved_filename}.mp4"
                    remote = remotes[remote_name]
                    if 'disabled' in remote and remote['disabled']:
                        continue

                    try:
                        if not self.arg_parse.other_dir and "vdest" not in remote:
                            print(f"Warning: remote {remote_name} has no key called vdest")
                            continue
                        if self.arg_parse.other_dir:
                            dest = remote['other']
                        else:
                            dest = remote['vdest']
                        method = remote['method'] if 'method' in remote and remote['method'] else 'scp'
                        if self.arg_parse.args.xrated: dest = remote['xdest']
                        print(f"Copying {saved_filename}.{self.arg_parse.format} to {remote_name}:{dest}/{self.name}.{self.arg_parse.format}")
                        run(f"{method} {saved_filename}.{self.arg_parse.format} {remote_name}:{dest}", silent=not self.arg_parse.args.debug)
                    except Exception as exception:
                        print(Fore.YELLOW + str(type(exception)) + f": while copying {saved_filename}.{self.arg_parse.format} to {dest}{Style.RESET_ALL}")
        else:
            sys.exit(f"Invalid action '{self.arg_parse.action}'")

if __name__ == '__main__':
    config: DLConfig = DLConfig()
    arg_parse: ArgParse = ArgParse(config)
    # media_file = MediaFile(config = config.local, path='')
    dl: DL = DL(arg_parse)
    dl.doit()
