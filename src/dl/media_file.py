import os
import sys
from pathlib import Path
from os import environ
from dl.dl_config import DLConfig

class MediaFile:
    def __init__(self, config: DLConfig, path: Path) -> None:
        self.config = config
        self.path = path
        self.mp3_dest = self.set_mp3_dest()

    def is_mp3(self) -> bool:
        """Indicate if is this MediaFile is an MP3 audio file

        Returns:
            bool: true if this is MediaFile is an MP3 audio file
        """
        return self.path.name.endswith('.mp3')

    def is_video(self) -> bool:
        """Indicate if is this MediaFile is a video file

        Returns:
            bool: true if this is MediaFile is a video file
        """
        return self.path.name.endswith('.mp4')

    def file_type(self) -> str:
        """Obtains filetype (extension) of this MediaFile

        Returns:
            str: filetype of this MediaFile
        """
        return self.path.suffix

    def set_mp3_dest(self) -> None:
        """TODO: Is this a good idea?
            If mp3s directory specified in config file, use it
            Else if mp3s is defined as an environment variable, use it
            Else use ~/Music/mp3s if it exists
            Else crash exit
        """
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
