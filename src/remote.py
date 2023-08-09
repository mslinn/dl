import shutil
import src.util as util

class Remote:
    def __init__(self, node_name, disabled=False, method='scp', mp3s = None, vdest = None, xdest = None) -> None:
       self.disabled = disabled
       self.method = method
       self.node_name = node_name
       self.mp3s = mp3s
       self.vdest = vdest
       self.xdest = xdest

    def copy_to(self, other: 'Remote', path):
        if self.method == 'samba':
            remote_drive, local_path = util.samba_parse(other.node_name, other.mp3s)
            samba_root = util.samba_mount(other.node_name, path, self.debug)
            target = f"{samba_root}{local_path}/{self.mp3_name}.{format}"
            print(f"Copying to {target}")
            shutil.copyfile(source, target)
        else:
            print(f"Copying to {target}/{self.mp3_name}.{format}")
            util.run(f"{other.method} {source} {target}", silent=not self.debug)
