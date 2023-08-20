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

    # Copy source to remote
    def copy_to(self, media_file: MediaFile, other: 'Remote'):
        if other.method == 'samba':
            if isinstance(other.mp3s, Path):
                remote_drive, remote_path = util.samba_parse(other.mp3s)
                samba_root = util.samba_mount(other.node_name, remote_drive, self.debug)
                target = f"{samba_root}{remote_path}/{self.mp3s.name}.{format}"
                print(f"Copying to {target}")
                shutil.copyfile(media_file.path, target)
        else:
            target = f"//{other.node_name}/{remote_path}/{media_file.file_name}"
            print(f"Copying to {target} using {other.method}")
            util.run(f"{other.method} {media_file.path} {target}", silent=not self.debug)
