import shutil
import dl.util as util
from dl.media_file import MediaFile

class Remote:
    def __init__(self, node_name, disabled=False, method='scp', mp3_path = None, video_path = None, x_path = None) -> None:
       if not method in ['scp', 'samba']:
            print(f"Error: method '{self.method}' is invalid. Allowable values are: samba and scp")
            exit(1)

       self.disabled = disabled
       self.method = method
       self.node_name = node_name
       self.mp3s = mp3_path
       self.video_path = video_path
       self.x_path = x_path

    # Copy source to remote
    def copy_to(self, media_file: MediaFile, other: 'Remote'):
        if other.method == 'samba':
            remote_drive, remote_path = util.samba_parse(other.mp3_path)
            samba_root = util.samba_mount(other.node_name, remote_drive, self.debug)
            target = f"{samba_root}{remote_path}/{self.mp3_name}.{format}"
            print(f"Copying to {target}")
            shutil.copyfile(media_file.path, target)
        else:
            target = f"//{other.node_name}/{remote_path}/{media_file.file_name}"
            print(f"Copying to {target} using {other.method}")
            util.run(f"{other.method} {media_file.path} {target}", silent=not self.debug)
