import shutil
import dl.util as util
from pathlib import Path
from typing import Union
from dl.media_file import MediaFile

PathNone = Union[Path, None]

class Remote:
    def __init__(self, node_name: str, disabled: bool=False, method: str='scp', mp3_path: PathNone = None, video_path:PathNone = None, x_path:PathNone = None) -> None:
       if not method in ['scp', 'samba']:
            print(f"Error: method '{self.method}' is invalid. Allowable values are: samba and scp")
            exit(1)

       self.disabled: bool = disabled
       self.method: str = method
       self.node_name: str = node_name
       self.mp3s: PathNone = mp3_path
       self.video_path: PathNone = video_path
       self.x_path: PathNone = x_path

    def copy_to(self, purpose: str, media_file: MediaFile, other: 'Remote'):
        """
        Copy source to remote, fails if the remote directory does not exist
        @return None
        """
        debug = False
        remote_path = None
        name = media_file.path.name
        match purpose:
            case 'mp3s':
                if isinstance(other.mp3s, Path):
                    remote_path = other.mp3s
                else:
                    exit(f"Remote {other.node_name} does not define a path for mp3s.")
            case 'vdest':
                if isinstance(other.video_path, Path):
                    remote_path = other.video_path
                else:
                    exit(f"Remote {other.node_name} does not define a path for videos.")
            case 'xdest':
                if isinstance(other.x_path, Path):
                    remote_path = other.x_path
                else:
                    exit(f"Remote {other.node_name} does not define a path for x-rated videos.")
            case _:
                exit(f"Unknown purpose '{purpose}'")
        if other.method == 'samba':
            remote_drive, remote_path = util.samba_parse(remote_path)
            samba_root = util.samba_mount(other.node_name, remote_drive, debug)
            target = f"{samba_root}{remote_path}/{name}"
            print(f"Copying to {target} using {other.method}")
            shutil.copyfile(media_file.path, target)
        else:
            target = f"//{other.node_name}/{remote_path}/{name}"
            print(f"Copying to {target} using {other.method}")
            util.run(f"{other.method} {media_file.path} {target}", silent=not debug)
