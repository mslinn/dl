import os
import sys
from . import *
from os import environ

class MediaFile:
    def __init__(self, config, path) -> None:
        self.config = config
        self.path = path
        mp3s = config['mp3s']
        vdest = config['vdest']
        xdest = config['xdest']
        self.set_mp3_dest()

    def is_mp3(self):
      return self.path.endswith('.mp3')

    def is_video(self):
      return self.path.endswith('.mp4')

    def set_mp3_dest(self):
        global mp3_dest
        if os.path.isdir(self.config['mp3s']['automount']):
            mp3_dest = self.config['mp3s']['automount']
        elif environ.get('mp3s') is not None:
            mp3_dest = environ.get('mp3s')
            if not os.path.isdir(mp3_dest):
                sys.exit(f"mp3s environment variable points to {mp3_dest}, but that directory does not exist.")
        elif os.path.expandvars(os.path.isdir(self.config['mp3s'])):
            mp3_dest = os.path.expandvars(self.config['mp3s'])
        else:
            mp3_dest = os.path.expanduser('~') + "/Music/mp3s"
            if not os.path.isdir(mp3_dest):
                sys.exit('Unable to find directory for mp3s.')
