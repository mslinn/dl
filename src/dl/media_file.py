import os
import sys
from pathlib import Path
from os import environ
from dl.dl_config import DLConfig

class MediaFile:
    def __init__(self, config: DLConfig, path: Path) -> None:
        self.config = config
        self.path = path
        # self.mp3_dest = self.set_mp3_dest()
        self.is_mp3 = self.path.name.endswith('.mp3')
        self.is_video = self.path.name.endswith('.mp4')
        self.file_type = self.path.suffix.removeprefix('.')

    def set_mp3_dest(self) -> None:
        """TODO: Is this a good idea?
            If mp3s directory specified in config file, use it
            Else if mp3s is defined as an environment variable, use it
            Else use ~/Music/mp3s if it exists
            Else crash exit
        """
        if isinstance(self.config.local, dict):
            mp3s = self.config.mp3s(self.config.local)
            if isinstance(mp3s, Path) and mp3s.is_dir():
                self.mp3_dest = mp3s
                return

        mp3s = environ.get('mp3s')
        if isinstance(mp3s, str) and len(str(mp3s))>0:
            self.mp3_dest = Path(mp3s)
            if not self.mp3_dest.is_dir():
                sys.exit(f"mp3s environment variable points to '{self.mp3_dest}', but that directory does not exist.")
            return

        self.mp3_dest = os.path.expanduser('~') + "/Music/mp3s"
        if not os.path.isdir(self.mp3_dest):
            sys.exit('Unable to find directory for mp3s.')
